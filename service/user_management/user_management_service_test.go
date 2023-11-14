// Don't run test per file without -p 1
// or simply run test per func or run
// project test using make test command
// check Makefile file
package service

import (
	"testing"

	"github.com/Lukman-e/try-gost/database/connector"
	"github.com/Lukman-e/try-gost/domain/base"
	"github.com/Lukman-e/try-gost/domain/model"
	"github.com/Lukman-e/try-gost/internal/constants"
	"github.com/Lukman-e/try-gost/internal/env"
	"github.com/Lukman-e/try-gost/internal/helper"
	"github.com/gofiber/fiber/v2"
)

func init() {
	// Check env and database
	env.ReadConfig("./../../.env")

	connector.LoadDatabase()
	connector.LoadRedisCache()
}

func TestNewUserManagementService(t *testing.T) {
	svc := NewUserManagementService()
	if svc == nil {
		t.Error(constants.ShouldNotNil)
	}
}

// Create 1 user
// -> get by id
// -> get by email
// -> get all and check >= 1
// -> update
// -> delete
// -> get by id (checking)
// -> get by email (checking)

func TestSuccessCrud(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	svc := NewUserManagementService()
	if svc == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}

	userModel := model.UserCreate{
		Name:     "John Doe",
		Email:    helper.RandomEmail(),
		Password: "password",
		IsAdmin:  true,
	}
	userID, createErr := svc.Create(ctx, userModel)
	if createErr != nil || userID < 1 {
		t.Error("should not error or id should more than or equal one")
	}
	defer func() {
		svc.Delete(ctx, userID)
	}()

	userByID, getByIDErr := svc.GetByID(ctx, userID)
	if getByIDErr != nil || userByID == nil {
		t.Error("should not error or user should not nil")
	}
	if userByID.Name != userModel.Name || userByID.Email != userModel.Email {
		t.Error("name and email should same")
	}

	userByEmail, getByEmailErr := svc.GetByEmail(ctx, userModel.Email)
	if getByEmailErr != nil || userByEmail == nil {
		t.Error("should not error or user should not nil")
	}
	if userByEmail.Name != userModel.Name || userByEmail.Email != userModel.Email {
		t.Error("name and email should same")
	}

	users, total, getAllErr := svc.GetAll(ctx, base.RequestGetAll{Limit: 10, Page: 1})
	if len(users) < 1 || total < 1 || getAllErr != nil {
		t.Error("should more than or equal one and not error at all")
	}

	updateUserData := model.UserProfileUpdate{
		ID:   userID,
		Name: "John Doe Update",
	}
	updateErr := svc.Update(ctx, updateUserData)
	if updateErr != nil {
		t.Error(constants.ShouldNotErr)
	}

	// reset value
	getByIDErr = nil
	userByID = nil
	userByID, getByIDErr = svc.GetByID(ctx, userID)
	if getByIDErr != nil || userByID == nil {
		t.Error("should not error or user should not nil")
	}
	if userByID.Name != updateUserData.Name || userByID.Email != userModel.Email {
		t.Error("name and email should same")
	}

	deleteErr := svc.Delete(ctx, userID)
	if deleteErr != nil {
		t.Error(constants.ShouldNotErr)
	}

	// reset value
	getByIDErr = nil
	userByID = nil
	userByID, getByIDErr = svc.GetByID(ctx, userID)
	if getByIDErr == nil || userByID != nil {
		t.Error("should error and user should nil")
	}
	fiberErr, ok := getByIDErr.(*fiber.Error)
	if ok {
		if fiberErr.Code != fiber.StatusNotFound {
			t.Error("should error 404")
		}
	}

	// reset value
	userByEmail = nil
	getByEmailErr = nil
	userByEmail, getByEmailErr = svc.GetByEmail(ctx, userModel.Email)
	if getByEmailErr == nil || userByEmail != nil {
		t.Error("should error or user should nil")
	}

	fiberErr, ok = getByEmailErr.(*fiber.Error)
	if ok {
		if fiberErr.Code != fiber.StatusNotFound {
			t.Error("should error 404")
		}
	}
}
