package relay

//Database is an abstraction around the pip boy database
type Database struct {
	entries  map[uint32]*DataEntry
	parents  map[uint32]uint32
	children map[uint32][]uint32
}

//Update updates a database based on entry list
func (db *Database) Update(des []*DataEntry) {
	if db.entries == nil {
		db.entries = make(map[uint32]*DataEntry, 30000)
		db.parents = make(map[uint32]uint32, 30000)
		db.children = make(map[uint32][]uint32, 4000)
	}

	for i, de := range des {
		switch de.Type {
		case ModifyEntry:
			for _, ins := range de.Value.(InsRemove).Insert {
				//Handle mappings
				db.parents[ins.Ref] = de.ID
				if children, ok := db.children[de.ID]; ok {
					children = append(children, ins.Ref)
				} else {
					db.children[de.ID] = []uint32{ins.Ref}
				}

				if existing, ok := db.entries[ins.Ref]; ok {
					existing.Name = ins.Name
				} else {
					db.entries[ins.Ref] = &DataEntry{
						ID:   ins.Ref,
						Name: ins.Name,
					}
				}
			}
		case ListEntry:
			list := de.Value.([]uint32)
			for _, p := range list {
				db.parents[p] = de.ID
			}
			if children, ok := db.children[de.ID]; ok {
				children = append(children, list...)
			} else {
				db.children[de.ID] = list
			}
		}

		if existing, ok := db.entries[de.ID]; ok {
			existing.Value = de.Value
			des[i] = existing
		} else {
			db.entries[de.ID] = de
		}
	}
}

func (db *Database) ToJSON(item uint32) interface{} {
	de := db.entries[item]

	js := de.Value

	switch de.Type {
	case ModifyEntry:
		ins := de.Value.(InsRemove)
		m := make(map[string]interface{})
		for _, i := range ins.Insert {
			m[i.Name] = db.ToJSON(i.Ref)
		}
		js = m
	case ListEntry:
		l := make([]interface{}, 0, 10)
		for _, ref := range de.Value.([]uint32) {
			l = append(l, db.ToJSON(ref))
		}
		js = l
	}

	return js
}
