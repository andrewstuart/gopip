package relay

//Database is an abstraction around the pip boy database
type Database struct {
	entries map[uint32]*DataEntry
	parents map[uint32]*DataEntry
	byName  map[string]*DataEntry
}

//Update updates a database based on entry list
func (db *Database) Update(des []*DataEntry) {
	if db.entries == nil {
		db.entries = make(map[uint32]*DataEntry, 30000)
		db.parents = make(map[uint32]*DataEntry, 30000)
		db.byName = make(map[string]*DataEntry, 30000)
	}

	for i, de := range des {
		switch de.Type {
		case ModifyEntry: //reference types
			upd := de.Value.(InsRemove)

			for _, u := range upd.Insert {
				//Set parent
				db.parents[u.Ref] = de

				//If in database
				if child, ok := db.entries[u.Ref]; ok {
					child.Name = u.Name
					de.Children = append(de.Children, child)
				} else {
					child = &DataEntry{
						ID:   u.Ref,
						Name: u.Name,
					}

					db.entries[u.Ref] = child
					de.Children = append(de.Children, child)
				}
			}
		case ListEntry:
			de.Name = "$LIST"
			for _, ref := range de.Value.([]uint32) {
				//Double pointer
				db.parents[ref] = de
			}
		}

		if existing, ok := db.entries[de.ID]; ok {
			existing.Value = de.Value
			des[i] = existing
		} else {
			db.entries[de.ID] = de

			if parent, ok := db.parents[de.ID]; ok {
				parent.Children = append(parent.Children, de)
			}
		}

	}
}

func (db *Database) Get(s string) *DataEntry {
	return db.byName[s]
}
