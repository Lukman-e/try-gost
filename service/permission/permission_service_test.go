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
)

func init() {
	// Check env and database
	env.ReadConfig("./../../.env")

	connector.LoadDatabase()
	connector.LoadRedisCache()
}

func TestNewPermissionService(t *testing.T) {
	svc := NewPermissionService()
	if svc == nil {
		t.Error(constants.ShouldNotNil)
	}
}

// Create 1 role
// -> get by id
// -> get all and check >= 1
// -> update
// -> delete
// -> get by id

func TestSuccessCrudPermission(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	svc := NewPermissionService()
	if svc == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}
	modelPerm := model.PermissionCreate{
		Name:        strings.ToLower(helper.RandomString(10)),
		Description: helper.RandomString(30),
	}
	permID, createErr := svc.Create(ctx, modelPerm)
	if createErr != nil || permID < 1 {
		t.Error("should not error and permID should more than one, but got", createErr.Error())
	}
	defer func() {
		svc.Delete(ctx, permID)
	}()

	permByID, getErr := svc.GetByID(ctx, permID)
	if getErr != nil || permByID == nil {
		t.Error("should not error and permByID should not nil")
	}
	if permByID.Name != modelPerm.Name || permByID.Description != modelPerm.Description {
		t.Error("name and desc should same")
	}

	perms, total, getAllErr := svc.GetAll(ctx, base.RequestGetAll{Limit: 10, Page: 1})
	if len(perms) < 1 || total < 1 || getAllErr != nil {
		t.Error("should more than or equal one and not error at all")
	}

	updatePermModel := model.PermissionUpdate{
		ID:          permID,
		Name:        strings.ToLower(helper.RandomString(11)),
		Description: helper.RandomString(31),
	}
	updateErr := svc.Update(ctx, updatePermModel)
	if updateErr != nil {
		t.Error(constants.ShouldNotErr)
	}

	// value reset
	permByID = nil
	getErr = nil
	permByID, getErr = svc.GetByID(ctx, permID)
	if getErr != nil || permByID == nil {
		t.Error("should not error and permByID should not nil")
	}
	if permByID.Name != updatePermModel.Name || permByID.Description != updatePermModel.Description {
		t.Error("name and desc should same")
	}

	deleteErr := svc.Delete(ctx, permID)
	if deleteErr != nil {
		t.Error(constants.ShouldNotErr)
	}

	// value reset
	permByID = nil
	getErr = nil
	permByID, getErr = svc.GetByID(ctx, permID)
	if getErr == nil || permByID != nil {
		t.Error("should error and permByID should nil")
	}
}

func TestFailedCrudPermission(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	svc := NewPermissionService()
	if svc == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}
	modelPerm := model.PermissionCreate{
		Name:        strings.ToLower(helper.RandomString(10)),
		Description: helper.RandomString(30),
	}
	permID, createErr := svc.Create(ctx, modelPerm)
	if createErr != nil || permID < 1 {
		t.Error("should not error and permID should more than one")
	}
	defer func() {
		svc.Delete(ctx, permID)
	}()

	permByID, getErr := svc.GetByID(ctx, -10)
	if getErr == nil || permByID != nil {
		t.Error("should error and permByID should nil")
	}

	updatePermModel := model.PermissionUpdate{
		ID:          -10,
		Name:        strings.ToLower(helper.RandomString(11)),
		Description: helper.RandomString(31),
	}
	updateErr := svc.Update(ctx, updatePermModel)
	if updateErr == nil {
		t.Error(constants.ShouldErr)
	}

	deleteErr := svc.Delete(ctx, -10)
	if deleteErr == nil {
		t.Error(constants.ShouldErr)
	}
}
