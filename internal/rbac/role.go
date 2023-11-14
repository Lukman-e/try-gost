package rbac

import "github.com/Lukman-e/try-gost/domain/entity"

// AllRoles func return all roles entities that has been
// created by developer. This func run self audit that
// check for name should be unique value.
// ⚠️ Do not forget to put new role here.
func AllRoles() []entity.Role {
	roleNames := []string{
		RoleAdmin,
		RoleUser,

		// ...
		// add more here
	}

	roles := []entity.Role{}
	for _, name := range roleNames {
		newRoleEntity := entity.Role{
			Name: name,
		}
		newRoleEntity.SetCreateTime()
		roles = append(roles, newRoleEntity)
	}

	return roles
}

const (
	RoleAdmin = "admin"
	RoleUser  = "user"

	// ...
	// add more here
)
