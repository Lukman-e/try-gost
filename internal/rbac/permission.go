package rbac

import (
	"log"

	"github.com/Lukman-e/try-gost/domain/entity"
)

// AllPermissions func return all permissions entities
// that has been created by developer. This func run self
// audit that check for name and id should be unique value.
// ⚠️ Do not forget to run TestAllPermissions to audit
func AllPermissions() []entity.Permission {
	permissions := []entity.Permission{
		// user
		PermCreateUser, PermViewUser, PermUpdateUser, PermDeleteUser,
		// user has role
		PermCreateUserHasRole, PermViewUserHasRole, PermUpdateUserHasRole, PermDeleteUserHasRole,
		// role
		PermCreateRole, PermViewRole, PermUpdateRole, PermDeleteRole,
		// role has permissions
		PermCreateRoleHasPermissions, PermViewRoleHasPermissions, PermUpdateRoleHasPermissions, PermDeleteRoleHasPermissions,
		// permission
		PermCreatePermission, PermViewPermission, PermUpdatePermission, PermDeletePermission,
		// ...
		// add more permissions
	}

	// self tested check id and name
	// id and name should unique
	checkIDs := make(map[int]int)
	checkNames := make(map[string]int)
	for _, perm := range permissions {
		if perm.ID < 1 || len(perm.Name) <= 1 {
			log.Fatal("permission name too short or invalid id at:", perm)
		}
		checkIDs[perm.ID]++
		checkNames[perm.Name]++
		if checkIDs[perm.ID] > 1 || checkNames[perm.Name] > 1 {
			log.Fatal("permission name or id should unique, but got:", perm)
		}
	}

	return permissions
}

var (
	PermCreateUser = entity.Permission{ID: 1, Name: "create-user", Description: "CRUD for User Entity"}
	PermViewUser   = entity.Permission{ID: 2, Name: "view-user", Description: "CRUD for User Entity"}
	PermUpdateUser = entity.Permission{ID: 3, Name: "update-user", Description: "CRUD for User Entity"}
	PermDeleteUser = entity.Permission{ID: 4, Name: "delete-user", Description: "CRUD for User Entity"}

	PermCreateUserHasRole = entity.Permission{ID: 5, Name: "create-user-has-role", Description: "CRUD for User-Has-Role entity"}
	PermViewUserHasRole   = entity.Permission{ID: 6, Name: "view-user-has-role", Description: "CRUD for User-Has-Role entity"}
	PermUpdateUserHasRole = entity.Permission{ID: 7, Name: "update-user-has-role", Description: "CRUD for User-Has-Role entity"}
	PermDeleteUserHasRole = entity.Permission{ID: 8, Name: "delete-user-has-role", Description: "CRUD for User-Has-Role entity"}

	PermCreateRole = entity.Permission{ID: 9, Name: "create-role", Description: "CRUD for Role Entity"}
	PermViewRole   = entity.Permission{ID: 10, Name: "view-role", Description: "CRUD for Role Entity"}
	PermUpdateRole = entity.Permission{ID: 11, Name: "update-role", Description: "CRUD for Role Entity"}
	PermDeleteRole = entity.Permission{ID: 12, Name: "delete-role", Description: "CRUD for Role Entity"}

	PermCreateRoleHasPermissions = entity.Permission{ID: 13, Name: "create-role-has-permissions", Description: "CRUD for Role-Has-Permission Entity"}
	PermViewRoleHasPermissions   = entity.Permission{ID: 14, Name: "view-role-has-permissions", Description: "CRUD for Role-Has-Permission Entity"}
	PermUpdateRoleHasPermissions = entity.Permission{ID: 15, Name: "update-role-has-permissions", Description: "CRUD for Role-Has-Permission Entity"}
	PermDeleteRoleHasPermissions = entity.Permission{ID: 16, Name: "delete-role-has-permissions", Description: "CRUD for Role-Has-Permission Entity"}

	PermCreatePermission = entity.Permission{ID: 17, Name: "create-permission", Description: "CRUD for Role Entity"}
	PermViewPermission   = entity.Permission{ID: 18, Name: "read-permission", Description: "CRUD for Role Entity"}
	PermUpdatePermission = entity.Permission{ID: 19, Name: "update-permission", Description: "CRUD for Role Entity"}
	PermDeletePermission = entity.Permission{ID: 20, Name: "delete-permission", Description: "CRUD for Role Entity"}

	// ...
	// add more permissions
	// Rule :
	// Name should unique
	// ID +1 from the ID before
)
