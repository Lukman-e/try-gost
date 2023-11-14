// Don't run test per file without -p 1
// or simply run test per func or run
// project test using make test command
// check Makefile file
package service

import (
	"strings"
	"testing"

	"github.com/Lukman-e/try-gost/database/connector"
	"github.com/Lukman-e/try-gost/domain/base"
	"github.com/Lukman-e/try-gost/domain/model"
	"github.com/Lukman-e/try-gost/internal/constants"
	"github.com/Lukman-e/try-gost/internal/env"
	"github.com/Lukman-e/try-gost/internal/helper"
	permService "github.com/Lukman-e/try-gost/service/permission"
)

func init() {
	// Check env and database
	env.ReadConfig("./../../.env")

	connector.LoadDatabase()
	connector.LoadRedisCache()
}

func TestNewRoleService(t *testing.T) {
	permSvc := permService.NewPermissionService()
	svc := NewRoleService(permSvc)
	if svc == nil {
		t.Error(constants.ShouldNotNil)
	}
}

// create 1 role, create 4 permissions
// trying to connect
func TestSuccessCrudRole(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	permSvc := permService.NewPermissionService()
	if permSvc == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}
	svc := NewRoleService(permSvc)
	if svc == nil {
		t.Error(constants.ShouldNotNil)
	}

	modelRole := model.RoleCreate{
		Name:        strings.ToLower(helper.RandomString(10)),
		Description: helper.RandomString(30),
	}
	roleID, createErr := svc.Create(ctx, modelRole)
	if createErr != nil || roleID < 1 {
		t.Error("should not error and id should more than zero")
	}

	// Save the ID for deleting the permissions
	permsID := make([]int, 0)
	for i := 0; i < 3; i++ {
		modelPerm := model.PermissionCreate{
			Name:        strings.ToLower(helper.RandomString(10)),
			Description: helper.RandomString(30),
		}
		permID, createErr := permSvc.Create(ctx, modelPerm)
		if createErr != nil || permID < 1 {
			t.Error("should not error and permID should be more than one")
		}

		permsID = append(permsID, permID)
	}

	defer func() {
		svc.Delete(ctx, roleID)
		for _, id := range permsID {
			permSvc.Delete(ctx, id)
		}
	}()

	// Success connect
	modelConnect := model.RoleConnectToPermissions{
		RoleID:        roleID,
		PermissionsID: permsID,
	}
	connectErr := svc.ConnectPermissions(ctx, modelConnect)
	if connectErr != nil {
		t.Error(constants.ShouldNotErr)
	}

	roleByID, getErr := svc.GetByID(ctx, roleID)
	if getErr != nil || roleByID == nil {
		t.Error("should not error and role not nil")
	}
	if len(roleByID.Permissions) != len(permsID) {
		t.Error("total of permissions connected by role should be equal")
	}

	roles, total, getAllErr := svc.GetAll(ctx, base.RequestGetAll{Limit: 10, Page: 1})
	if len(roles) < 1 || total < 1 || getAllErr != nil {
		t.Error("should be more than or equal to one and not error at all")
	}

	updateRoleModel := model.RoleUpdate{
		ID:          roleID,
		Name:        strings.ToLower(helper.RandomString(11)),
		Description: helper.RandomString(31),
	}
	updateErr := svc.Update(ctx, updateRoleModel)
	if updateErr != nil {
		t.Error(constants.ShouldNotErr)
	}

	// Value reset
	roleByID = nil
	getErr = nil
	roleByID, getErr = svc.GetByID(ctx, roleID)
	if getErr != nil || roleByID == nil {
		t.Error("should not error and roleByID should not be nil")
	}
	if roleByID.Name != updateRoleModel.Name || roleByID.Description != updateRoleModel.Description {
		t.Error("name and description should be the same")
	}

	deleteErr := svc.Delete(ctx, roleID)
	if deleteErr != nil {
		t.Error(constants.ShouldNotErr)
	}

	// Value reset
	roleByID = nil
	getErr = nil
	roleByID, getErr = svc.GetByID(ctx, roleID)
	if getErr == nil || roleByID != nil {
		t.Error("should error and roleByID should be nil")
	}
}

func TestFailedCrudRoles(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	permSvc := permService.NewPermissionService()
	if permSvc == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}
	svc := NewRoleService(permSvc)
	if svc == nil {
		t.Error(constants.ShouldNotNil)
	}

	// failed create: permissions not found
	func() {
		modelRole := model.RoleCreate{
			Name:          strings.ToLower(helper.RandomString(10)),
			Description:   helper.RandomString(30),
			PermissionsID: []int{-1, -2, -3},
		}
		roleID, createErr := svc.Create(ctx, modelRole)
		if createErr == nil || roleID != 0 {
			t.Error("should error and id should zero")
		}
	}()

	// success create
	modelRole := model.RoleCreate{
		Name:        strings.ToLower(helper.RandomString(10)),
		Description: helper.RandomString(30),
	}
	roleID, createErr := svc.Create(ctx, modelRole)
	if createErr != nil || roleID < 1 {
		t.Error("should not error and id should more than zero")
	}

	defer func() {
		svc.Delete(ctx, roleID)
	}()

	// failed connect
	modelConnectFailed := model.RoleConnectToPermissions{
		RoleID:        roleID,
		PermissionsID: []int{-3, -2, -1},
	}
	connectErr := svc.ConnectPermissions(ctx, modelConnectFailed)
	if connectErr == nil {
		t.Error(constants.ShouldErr)
	}

	modelConnectFailed = model.RoleConnectToPermissions{
		RoleID:        -1,
		PermissionsID: []int{},
	}
	connectErr = nil
	connectErr = svc.ConnectPermissions(ctx, modelConnectFailed)
	if connectErr == nil {
		t.Error(constants.ShouldErr)
	}

	// failed update
	updateRoleModel := model.RoleUpdate{
		ID:          -1,
		Name:        strings.ToLower(helper.RandomString(11)),
		Description: helper.RandomString(31),
	}
	updateErr := svc.Update(ctx, updateRoleModel)
	if updateErr == nil {
		t.Error(constants.ShouldErr)
	}

	// failed delete
	deleteErr := svc.Delete(ctx, -1)
	if deleteErr == nil {
		t.Error(constants.ShouldErr)
	}
}
