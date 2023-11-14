// ⚠️ Don't forget to Add Your new
// Table Here : must sorted by developer.

// Package entity contains all the structs that will be used to build tables in the database.
package entity

import "log"

// Table interface is contract that make developer
// not forget to add TableName method for struct
type Table interface {
	TableName() string
}

// AllTables func serve/ return all structs that
// developer has been created. This func used in
// database migration scripts.
func AllTables() []any {
	allTables := []any{
		&User{},
		&UserHasRoles{},
		&Role{},
		&RoleHasPermission{},
		&Permission{},

		// ...
		// Add more tables/structs
	}
	for _, table := range allTables {
		_, ok := table.(Table)
		if !ok {
			log.Fatal("please add TableName() func to all structs")
		}
	}
	return allTables
}
