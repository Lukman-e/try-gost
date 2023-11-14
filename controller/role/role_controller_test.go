// Don't run test per file without -p 1
// or simply run test per func or run
// project test using make test command
// check Makefile file
package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/Lukman-e/try-gost/database/connector"
	"github.com/Lukman-e/try-gost/domain/model"
	"github.com/Lukman-e/try-gost/internal/constants"
	"github.com/Lukman-e/try-gost/internal/env"
	"github.com/Lukman-e/try-gost/internal/helper"
	"github.com/Lukman-e/try-gost/internal/response"
	"github.com/gofiber/fiber/v2"

	permissionController "github.com/Lukman-e/try-gost/controller/permission"
	permSvc "github.com/Lukman-e/try-gost/service/permission"
	service "github.com/Lukman-e/try-gost/service/role"
)

var (
	permService    permSvc.PermissionService
	roleService    service.RoleService
	roleController RoleController
	permController permissionController.PermissionController
	appURL         string
)

func init() {
	env.ReadConfig("./../../.env")
	config := env.Configuration()
	appURL = config.AppURL

	connector.LoadDatabase()
	connector.LoadRedisCache()

	permService = permSvc.NewPermissionService()
	permController = permissionController.NewPermissionController(permService)

	roleService = service.NewRoleService(permService)
	roleController = NewRoleController(roleService)
}

func TestRoleCreate(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}

	permIDs := make([]int, 0)
	for i := 0; i < 4; i++ {
		// create 1 permission
		permID, createErr := permService.Create(ctx, model.PermissionCreate{
			Name:        helper.RandomString(11),
			Description: helper.RandomString(30),
		})
		if createErr != nil || permID < 1 {
			t.Fatal("should not error while creating permission")
		}
		defer func() {
			permService.Delete(ctx, permID)
		}()

		permIDs = append(permIDs, permID)
	}

	createdRole := model.RoleCreate{
		Name:          helper.RandomString(9),
		Description:   helper.RandomString(30),
		PermissionsID: permIDs,
	}
	roleID, createErr := roleService.Create(ctx, createdRole)
	if createErr != nil || roleID <= 0 {
		t.Fatal("should not error while creating new Role")
	}
	roleByID, getErr := roleService.GetByID(ctx, roleID)
	if getErr != nil || roleByID == nil {
		t.Fatal("should not error while getting Role")
	}
	if len(roleByID.Permissions) != 4 {
		t.Error("the length should equal")
	}
	defer func() {
		roleService.Delete(ctx, roleID)
	}()

	testCases := []struct {
		caseName string
		respCode int
		payload  model.RoleCreate
	}{
		{
			caseName: "success create -1",
			respCode: http.StatusCreated,
			payload: model.RoleCreate{
				Name:        helper.RandomString(10),
				Description: helper.RandomString(30),
			},
		},
		{
			caseName: "success create -2",
			respCode: http.StatusCreated,
			payload: model.RoleCreate{
				Name:        helper.RandomString(10),
				Description: helper.RandomString(30),
			},
		},
		{
			caseName: "failed create: permissions not found",
			respCode: http.StatusNotFound,
			payload: model.RoleCreate{
				Name:          helper.RandomString(10),
				Description:   helper.RandomString(30),
				PermissionsID: []int{permIDs[0] + 90},
			},
		},
		{
			caseName: "failed create: invalid name, too short",
			respCode: http.StatusBadRequest,
			payload: model.RoleCreate{
				Name:          "",
				Description:   helper.RandomString(30),
				PermissionsID: []int{permIDs[0] - 90},
			},
		},
		{
			caseName: "failed create: invalid description, too short",
			respCode: http.StatusBadRequest,
			payload: model.RoleCreate{
				Name:          helper.RandomString(10),
				Description:   "",
				PermissionsID: []int{permIDs[0]},
			},
		},
	}

	for _, tc := range testCases {
		log.Println(":::::::" + tc.caseName)
		jsonObject, err := json.Marshal(tc.payload)
		if err != nil {
			t.Error(constants.ShouldNotErr, err.Error())
		}
		url := appURL + "role"
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonObject))
		if err != nil {
			t.Error(constants.ShouldNotErr, err.Error())
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Post("/role", roleController.Create)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(constants.ShouldNotErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal, want", tc.respCode, "but got", resp.StatusCode)
		}
		var data response.Response
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			t.Fatal("failed to decode JSON:", err)
		}

		if resp.StatusCode == http.StatusCreated {
			data, ok1 := data.Data.(map[string]interface{})
			if !ok1 {
				t.Error("should ok1")
			}
			anyID, ok2 := data["id"]
			if !ok2 {
				t.Error("should ok2")
			}
			intID, ok3 := anyID.(float64)
			if !ok3 {
				t.Error("should ok3")
			}
			deleteErr := roleService.Delete(ctx, int(intID))
			if deleteErr != nil {
				t.Error(constants.ShouldNotErr)
			}
		}
	}
}

func TestRoleConnect(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}

	permIDs := make([]int, 0)
	for i := 0; i < 4; i++ {
		// create 1 permission
		permID, createErr := permService.Create(ctx, model.PermissionCreate{
			Name:        helper.RandomString(11),
			Description: helper.RandomString(30),
		})
		if createErr != nil || permID < 1 {
			t.Fatal("should not error while creating permission")
		}
		defer func() {
			permService.Delete(ctx, permID)
		}()

		permIDs = append(permIDs, permID)
	}

	createdRole := model.RoleCreate{
		Name:          helper.RandomString(9),
		Description:   helper.RandomString(30),
		PermissionsID: permIDs,
	}
	roleID, createErr := roleService.Create(ctx, createdRole)
	if createErr != nil || roleID <= 0 {
		t.Fatal("should not error while creating new Role")
	}
	roleByID, getErr := roleService.GetByID(ctx, roleID)
	if getErr != nil || roleByID == nil {
		t.Fatal("should not error while getting Role")
	}
	if len(roleByID.Permissions) != 4 {
		t.Error("the length should equal")
	}
	defer func() {
		roleService.Delete(ctx, roleID)
	}()

	testCases := []struct {
		caseName string
		respCode int
		payload  model.RoleConnectToPermissions
	}{
		{
			caseName: "success connect -1",
			respCode: http.StatusCreated,
			payload: model.RoleConnectToPermissions{
				RoleID:        roleID,
				PermissionsID: permIDs,
			},
		},
		{
			caseName: "success connect -2",
			respCode: http.StatusCreated,
			payload: model.RoleConnectToPermissions{
				RoleID:        roleID,
				PermissionsID: permIDs,
			},
		},
		{
			caseName: "failed connect: status not found",
			respCode: http.StatusNotFound,
			payload: model.RoleConnectToPermissions{
				RoleID:        roleID + 99,
				PermissionsID: permIDs,
			},
		},
		{
			caseName: "failed connect: invalid role id",
			respCode: http.StatusBadRequest,
			payload: model.RoleConnectToPermissions{
				RoleID:        -1,
				PermissionsID: permIDs,
			},
		},
		{
			caseName: "failed connect: invalid id",
			respCode: http.StatusBadRequest,
			payload: model.RoleConnectToPermissions{
				RoleID:        roleID,
				PermissionsID: []int{-1, 2, 3},
			},
		},
	}

	for _, tc := range testCases {
		log.Println(":::::::" + tc.caseName)
		jsonObject, err := json.Marshal(tc.payload)
		if err != nil {
			t.Error(constants.ShouldNotErr, err.Error())
		}
		url := appURL + "role/connect"
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonObject))
		if err != nil {
			t.Error(constants.ShouldNotErr, err.Error())
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Post("/role/connect", roleController.Connect)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(constants.ShouldNotErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal, want", tc.respCode, "but got", resp.StatusCode)
		}
		var data response.Response
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			t.Fatal("failed to decode JSON:", err)
		}
		if resp.StatusCode == http.StatusOK {
			role, getErr := roleService.GetByID(ctx, tc.payload.RoleID)
			if getErr != nil || role == nil {
				t.Fatal("should not error while getting role")
			}

			if len(role.Permissions) != len(tc.payload.PermissionsID) {
				t.Error("should equal")
			}
		}
	}
}

func TestRoleGet(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}

	permIDs := make([]int, 0)
	for i := 0; i < 4; i++ {
		// create 1 permission
		permID, createErr := permService.Create(ctx, model.PermissionCreate{
			Name:        helper.RandomString(11),
			Description: helper.RandomString(30),
		})
		if createErr != nil || permID < 1 {
			t.Fatal("should not error while creating permission")
		}
		defer func() {
			permService.Delete(ctx, permID)
		}()

		permIDs = append(permIDs, permID)
	}

	createdRole := model.RoleCreate{
		Name:          helper.RandomString(9),
		Description:   helper.RandomString(30),
		PermissionsID: permIDs,
	}
	roleID, createErr := roleService.Create(ctx, createdRole)
	if createErr != nil || roleID <= 0 {
		t.Fatal("should not error while creating new Role")
	}
	roleByID, getErr := roleService.GetByID(ctx, roleID)
	if getErr != nil || roleByID == nil {
		t.Fatal("should not error while getting Role")
	}
	if len(roleByID.Permissions) != 4 {
		t.Error("the length should equal")
	}
	defer func() {
		roleService.Delete(ctx, roleID)
	}()

	testCases := []struct {
		caseName string
		respCode int
		roleID   int
	}{
		{
			caseName: "success get -1",
			respCode: http.StatusOK,
			roleID:   roleID,
		},
		{
			caseName: "success get -2",
			respCode: http.StatusOK,
			roleID:   roleID,
		},
		{
			caseName: "failed get: status not found",
			respCode: http.StatusNotFound,
			roleID:   roleID + 99,
		},
		{
			caseName: "failed get: invalid id",
			respCode: http.StatusBadRequest,
			roleID:   -10,
		},
	}

	for _, tc := range testCases {
		url := fmt.Sprintf(appURL+"role/%d", tc.roleID)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			t.Error(constants.ShouldNotErr, err.Error())
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Get("/role/:id", roleController.Get)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(constants.ShouldNotErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal, want", tc.respCode, "but got", resp.StatusCode)
		}
		var data response.Response
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			t.Fatal("failed to decode JSON:", err)
		}
	}
}

func TestRoleGetAll(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}

	permIDs := make([]int, 0)
	for i := 0; i < 4; i++ {
		// create 1 permission
		permID, createErr := permService.Create(ctx, model.PermissionCreate{
			Name:        helper.RandomString(11),
			Description: helper.RandomString(30),
		})
		if createErr != nil || permID < 1 {
			t.Fatal("should not error while creating permission")
		}
		defer func() {
			permService.Delete(ctx, permID)
		}()

		permIDs = append(permIDs, permID)
	}

	createdRole := model.RoleCreate{
		Name:          helper.RandomString(9),
		Description:   helper.RandomString(30),
		PermissionsID: permIDs,
	}
	roleID, createErr := roleService.Create(ctx, createdRole)
	if createErr != nil || roleID <= 0 {
		t.Fatal("should not error while creating new Role")
	}
	roleByID, getErr := roleService.GetByID(ctx, roleID)
	if getErr != nil || roleByID == nil {
		t.Fatal("should not error while getting Role")
	}
	if len(roleByID.Permissions) != 4 {
		t.Error("the length should equal")
	}
	defer func() {
		roleService.Delete(ctx, roleID)
	}()

	testCases := []struct {
		caseName string
		respCode int
		params   string
	}{
		{
			caseName: "success getAll -1",
			respCode: http.StatusOK,
			params:   "limit=10&page=1",
		},
		{
			caseName: "success getAll -2",
			respCode: http.StatusOK,
			params:   "limit=100&page=1",
		},
		{
			caseName: "failed getAll: invalid limit/page",
			respCode: http.StatusBadRequest,
			params:   "limit=-10&page=-1",
		},
	}

	for _, tc := range testCases {
		url := appURL + "role?" + tc.params
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			t.Error(constants.ShouldNotErr, err.Error())
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Get("/role", roleController.GetAll)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(constants.ShouldNotErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal, want", tc.respCode, "but got", resp.StatusCode)
		}
		var data response.Response
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			t.Fatal("failed to decode JSON:", err)
		}
	}
}

func TestRoleUpdate(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}

	permIDs := make([]int, 0)
	for i := 0; i < 4; i++ {
		// create 1 permission
		permID, createErr := permService.Create(ctx, model.PermissionCreate{
			Name:        helper.RandomString(11),
			Description: helper.RandomString(30),
		})
		if createErr != nil || permID < 1 {
			t.Fatal("should not error while creating permission")
		}
		defer func() {
			permService.Delete(ctx, permID)
		}()

		permIDs = append(permIDs, permID)
	}

	createdRole := model.RoleCreate{
		Name:          helper.RandomString(9),
		Description:   helper.RandomString(30),
		PermissionsID: permIDs,
	}
	roleID, createErr := roleService.Create(ctx, createdRole)
	if createErr != nil || roleID <= 0 {
		t.Fatal("should not error while creating new Role")
	}
	roleByID, getErr := roleService.GetByID(ctx, roleID)
	if getErr != nil || roleByID == nil {
		t.Fatal("should not error while getting Role")
	}
	if len(roleByID.Permissions) != 4 {
		t.Error("the length should equal")
	}
	defer func() {
		roleService.Delete(ctx, roleID)
	}()

	testCases := []struct {
		caseName string
		respCode int
		roleID   int
		payload  model.RoleUpdate
	}{
		{
			caseName: "success update -1",
			respCode: http.StatusNoContent,
			roleID:   roleID,
			payload: model.RoleUpdate{
				ID:          roleID,
				Name:        helper.RandomString(12),
				Description: helper.RandomString(20),
			},
		},
		{
			caseName: "success update -2",
			respCode: http.StatusNoContent,
			roleID:   roleID,
			payload: model.RoleUpdate{
				ID:          roleID,
				Name:        helper.RandomString(12),
				Description: helper.RandomString(20),
			},
		},
		{
			caseName: "failed update: invalid name/description",
			respCode: http.StatusBadRequest,
			roleID:   roleID,
			payload: model.RoleUpdate{
				ID:          roleID,
				Name:        "",
				Description: "",
			},
		},
		{
			caseName: "failed update: invalid id",
			respCode: http.StatusBadRequest,
			roleID:   -10,
		},
		{
			caseName: "failed update: data not found",
			respCode: http.StatusNotFound,
			roleID:   roleID + 99,
			payload: model.RoleUpdate{
				ID:          roleID + 99,
				Name:        helper.RandomString(12),
				Description: helper.RandomString(20),
			},
		},
	}

	for _, tc := range testCases {
		log.Println(":::::::" + tc.caseName)
		jsonObject, err := json.Marshal(tc.payload)
		if err != nil {
			t.Error(constants.ShouldNotErr, err.Error())
		}
		url := fmt.Sprintf(appURL+"role/%d", tc.roleID)
		req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(jsonObject))
		if err != nil {
			t.Error(constants.ShouldNotErr, err.Error())
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Put("/role/:id", roleController.Update)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(constants.ShouldNotErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal, want", tc.respCode, "but got", resp.StatusCode)
		}
		if resp.StatusCode != http.StatusNoContent {
			var data response.Response
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				t.Fatal("failed to decode JSON:", err)
			}
		}
		if resp.StatusCode == http.StatusNoContent {
			perm, getErr := roleService.GetByID(ctx, roleID)
			if getErr != nil || perm == nil {
				t.Error("should not nil while get permission")
			}
			if perm.Name != strings.ToLower(tc.payload.Name) ||
				perm.Description != tc.payload.Description {
				t.Error("should equal")
			}
		}
	}
}

func TestRoleDelete(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}

	permIDs := make([]int, 0)
	for i := 0; i < 4; i++ {
		// create 1 permission
		permID, createErr := permService.Create(ctx, model.PermissionCreate{
			Name:        helper.RandomString(11),
			Description: helper.RandomString(30),
		})
		if createErr != nil || permID < 1 {
			t.Fatal("should not error while creating permission")
		}
		defer func() {
			permService.Delete(ctx, permID)
		}()

		permIDs = append(permIDs, permID)
	}

	createdRole := model.RoleCreate{
		Name:          helper.RandomString(9),
		Description:   helper.RandomString(30),
		PermissionsID: permIDs,
	}
	roleID, createErr := roleService.Create(ctx, createdRole)
	if createErr != nil || roleID <= 0 {
		t.Fatal("should not error while creating new Role")
	}
	roleByID, getErr := roleService.GetByID(ctx, roleID)
	if getErr != nil || roleByID == nil {
		t.Fatal("should not error while getting Role")
	}
	if len(roleByID.Permissions) != 4 {
		t.Error("the length should equal")
	}
	defer func() {
		roleService.Delete(ctx, roleID)
	}()

	testCases := []struct {
		caseName string
		respCode int
		roleID   int
	}{
		{
			caseName: "success update -1",
			respCode: http.StatusNoContent,
			roleID:   roleID,
		},
		{
			caseName: "success update -2",
			respCode: http.StatusNotFound,
			roleID:   roleID,
		},
		{
			caseName: "failed update: invalid id",
			respCode: http.StatusBadRequest,
			roleID:   -10,
		},
		{
			caseName: "failed update: data not found",
			respCode: http.StatusNotFound,
			roleID:   roleID + 99,
		},
	}

	for _, tc := range testCases {
		url := fmt.Sprintf(appURL+"role/%d", tc.roleID)
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			t.Error(constants.ShouldNotErr, err.Error())
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Delete("/role/:id", roleController.Delete)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(constants.ShouldNotErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal, want", tc.respCode, "but got", resp.StatusCode)
		}
		if resp.StatusCode != http.StatusNoContent {
			var data response.Response
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				t.Fatal("failed to decode JSON:", err)
			}
		}
		if resp.StatusCode == http.StatusNoContent {
			role, getErr := roleService.GetByID(ctx, tc.roleID)
			if getErr == nil || role != nil {
				t.Error("should error while get role")
			}
		}
	}
}
