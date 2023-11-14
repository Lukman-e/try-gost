// Don't run test per file without -p 1
// or simply run test per func or run
// project test using make test command
// check Makefile file
package controller

import (
	"net/http"
	"testing"

	"github.com/Lukman-e/try-gost/database/connector"
	"github.com/Lukman-e/try-gost/internal/env"
	"github.com/Lukman-e/try-gost/internal/helper"
	"github.com/gofiber/fiber/v2"
)

type handlerF = func(c *fiber.Ctx) error

func init() {
	// Check env and database
	env.ReadConfig("./../../.env")

	connector.LoadDatabase()
	connector.LoadRedisCache()
}

func TestNewDevControllerImpl(t *testing.T) {
	ctr := NewDevControllerImpl()
	c := helper.NewFiberCtx()
	if ctr == nil || c == nil {
		t.Error("should not error")
	}

	pingDbErr := ctr.PingDatabase(c)
	if pingDbErr != nil {
		t.Error("err: ", pingDbErr)
	}

	pingRedisErr := ctr.PingRedis(c)
	if pingRedisErr != nil {
		t.Error("err: ", pingRedisErr)
	}

	panicErr := ctr.Panic(c)
	if panicErr != nil {
		t.Error("err: ", panicErr)
	}

	storingErr := ctr.StoringToRedis(c)
	if storingErr != nil {
		t.Error("err: ", storingErr)
	}

	getErr := ctr.GetFromRedis(c)
	if getErr != nil {
		t.Error("err: ", getErr)
	}

	checkRoleErr := ctr.CheckNewRole(c)
	if checkRoleErr != nil {
		t.Error("err: ", checkRoleErr)
	}

	checkPermErr := ctr.CheckNewPermission(c)
	if checkPermErr != nil {
		t.Error("err: ", checkPermErr)
	}
}

func TestMethods(t *testing.T) {
	c := helper.NewFiberCtx()
	ctr := NewDevControllerImpl()
	if ctr == nil || c == nil {
		t.Error("should not nil")
	}

	testCases := []struct {
		caseName string
		method   handlerF
		respCode int
	}{
		{"PingDatabase", ctr.PingDatabase, http.StatusOK},
		{"PingRedis", ctr.PingRedis, http.StatusOK},
		{"Panic", ctr.Panic, http.StatusInternalServerError},
		{"StoringToRedis", ctr.StoringToRedis, http.StatusCreated},
		{"GetFromRedis", ctr.GetFromRedis, http.StatusOK},
		{"CheckNewRole", ctr.CheckNewRole, http.StatusOK},
		{"CheckNewPermission", ctr.CheckNewPermission, http.StatusOK},
		{"UploadFile", ctr.UploadFile, http.StatusBadRequest},
		{"RemoveFile", ctr.RemoveFile, http.StatusBadRequest},
		{"GetFilesList", ctr.GetFilesList, http.StatusOK},
	}

	for _, tc := range testCases {
		c := helper.NewFiberCtx()
		c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		tc.method(c)
		resp := c.Response()
		if resp.StatusCode() != tc.respCode {
			t.Errorf("Expected response code %d, but got %d on: %s", tc.respCode, resp.StatusCode(), tc.caseName)
		}
	}
}
