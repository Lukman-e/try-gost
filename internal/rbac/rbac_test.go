package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllPermissions(t *testing.T) {
	defer func() {
		r := recover()
		assert.Nil(t, r, "should not panic, but got:", r)
	}()

	permissions := AllPermissions()
	for _, permission := range permissions {
		assert.NotEmpty(t, permission.Name, "permission name should not be empty")
	}
}

func TestCountPermissions(t *testing.T) {
	hashMapPermissions := make(map[string]int, 0)

	for _, permission := range AllPermissions() {
		hashMapPermissions[permission.Name]++
		assert.LessOrEqual(t, hashMapPermissions[permission.Name], 1, "should be 1, not more; non-unique permission detected: %s", permission.Name)
	}

	assert.Equal(t, len(hashMapPermissions), len(AllPermissions()), "should have equal length; non-unique permission detected")
}

func TestAllRoles(t *testing.T) {
	roles := AllRoles()
	for _, role := range roles {
		assert.NotEmpty(t, role.Name, "name should not be empty")
	}
}

func TestCountRoles(t *testing.T) {
	hashMapRoles := make(map[string]int, 0)

	for _, role := range AllRoles() {
		hashMapRoles[role.Name]++
		assert.LessOrEqual(t, hashMapRoles[role.Name], 1, "should be 1, not more; non-unique role (role:name) detected: %s", role.Name)
	}

	assert.Equal(t, len(hashMapRoles), len(AllRoles()), "should have equal length; non-unique role detected")
}
