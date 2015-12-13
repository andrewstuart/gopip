package relay

//Database is an abstraction around the pip boy database
type Database struct {
	entries, parents map[uint32]*DataEntry
	byName           map[string]*DataEntry
}

//Update updates a database based on entry list
func (db *Database) Update(des []*DataEntry) {
	if db.entries == nil {
		db.entries = make(map[uint32]*DataEntry, 30000)
	}

	for i, de := range des {
		switch de.Type {
		case ModifyEntry:
			for _, ins := range de.Value.(InsRemove).Insert {
				if existing, ok := db.entries[ins.Ref]; ok {
					existing.Name = ins.Name
				} else {
					db.entries[ins.Ref] = &DataEntry{
						ID:   ins.Ref,
						Name: ins.Name,
					}
				}
			}
		default:
			if existing, ok := db.entries[de.ID]; ok {
				existing.Value = de.Value
				des[i] = existing
			} else {
				db.entries[de.ID] = de
			}
		}
	}
}

func (db *Database) Get(s string) *DataEntry {
	return db.byName[s]
}
