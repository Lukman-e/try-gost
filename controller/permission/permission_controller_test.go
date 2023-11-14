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
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/Lukman-e/try-gost/database/connector"
	"github.com/Lukman-e/try-gost/domain/base"
	"github.com/Lukman-e/try-gost/domain/model"
	"github.com/Lukman-e/try-gost/internal/constants"
	"github.com/Lukman-e/try-gost/internal/env"
	"github.com/Lukman-e/try-gost/internal/helper"
	"github.com/Lukman-e/try-gost/internal/middleware"
	"github.com/Lukman-e/try-gost/internal/response"

	userController "github.com/Lukman-e/try-gost/controller/user"
	userRepository "github.com/Lukman-e/try-gost/repository/user"
	service "github.com/Lukman-e/try-gost/service/permission"
	roleService "github.com/Lukman-e/try-gost/service/role"
	userService "github.com/Lukman-e/try-gost/service/user"
)

var (
	userRepo       userRepository.UserRepository
	permService    service.PermissionService
	permController PermissionController
	appURL         string
)

func init() {
	env.ReadConfig("./../../.env")
	config := env.Configuration()
	appURL = config.AppURL

	connector.LoadDatabase()
	connector.LoadRedisCache()

	userRepo = userRepository.NewUserRepository()
	permService = service.NewPermissionService()
	permController = NewPermissionController(permService)
}

func TestPermNewPermissionController(t *testing.T) {
	permSvc := service.NewPermissionService()
	permCtr := NewPermissionController(permSvc)

	if permSvc == nil || permCtr == nil {
		t.Error(constants.ShouldNotNil)
	}
}

func TestPermCreate(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}

	userID, userToken := createUserAndToken()
	if userID < 1 || len(userToken) < 2 {
		t.Error("should more")
	}

	defer func() {
		userRepo.Delete(ctx, userID)
	}()

	testCases := []struct {
		caseName string
		respCode int
		payload  model.PermissionCreate
		token    string // jwt into claims (fake claims)
	}{
		{
			caseName: "success create",
			respCode: http.StatusCreated,
			payload: model.PermissionCreate{
				Name:        "example-permission-001",
				Description: "example-description-of-permission-001",
			},
			token: userToken,
		},
		{
			caseName: "failed create: name already used",
			respCode: http.StatusBadRequest,
			payload: model.PermissionCreate{
				Name:        "example-permission-001",
				Description: "example-description-of-permission-001",
			},
			token: userToken,
		},
		{
			caseName: "failed create: name/desc too short",
			respCode: http.StatusBadRequest,
			payload:  model.PermissionCreate{},
			token:    userToken,
		},
	}

	createdIDs := make([]float64, 0)
	jwtHandler := middleware.NewJWTHandler()
	for _, tc := range testCases {
		c := helper.NewFiberCtx()
		c.Request().Header.Set(fiber.HeaderAuthorization, fmt.Sprintf("Bearer %s", tc.token))
		c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		requestBody, err := json.Marshal(tc.payload)
		if err != nil {
			t.Fatal("Error while serializing payload to request body")
		}
		c.Request().SetBody(requestBody)

		fakeClaims := jwtHandler.GenerateClaims(tc.token)
		if fakeClaims != nil {
			c.Locals("claims", fakeClaims)
		}
		ctr.Create(c)
		resp := c.Response()
		if resp.StatusCode() != tc.respCode {
			t.Error("should equal, but got", resp.StatusCode())
		}

		if resp.StatusCode() == http.StatusCreated {
			respBody := c.Response().Body()
			respString := string(respBody)
			respStruct := struct {
				Message string         `json:"message"`
				Success bool           `json:"success"`
				Data    map[string]any `json:"data"`
			}{}

			err := json.Unmarshal([]byte(respString), &respStruct)
			if err != nil {
				t.Errorf("Failed to parse response JSON: %v", err)
			}
			if !respStruct.Success {
				t.Error("Expected success")
			}
			if respStruct.Message != response.MessageSuccessCreated {
				t.Error("Expected message to be equal")
			}
			if id, ok := respStruct.Data["id"].(float64); !ok || id < 1 {
				t.Error("should be a positive integer")
			} else {
				createdIDs = append(createdIDs, id)
			}
		}
	}

	for _, id := range createdIDs {
		err := permService.Delete(ctx, int(id))
		if err != nil {
			t.Error("deletingc created permission/s should not error")
		}
	}
}

func TestPermGet(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}

	testCases := []struct {
		caseName string
		respCode int
		permID   int
	}{
		{
			caseName: "success get -1",
			respCode: http.StatusOK,
			permID:   1,
		},
		{
			caseName: "success get -2",
			respCode: http.StatusOK,
			permID:   1,
		},
		{
			caseName: "failed get: invalid id",
			respCode: http.StatusBadRequest,
			permID:   -10,
		},
		{
			caseName: "failed get: data not found",
			respCode: http.StatusNotFound,
			permID:   9999,
		},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/permission/%d", tc.permID), nil)
		app := fiber.New()
		app.Get("/permission/:id", permController.Get)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(constants.ShouldNotErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal, want", tc.respCode, "but got", resp.StatusCode)
		}
		if resp.StatusCode == http.StatusOK {
			var respStruct struct {
				Message string              `json:"message"`
				Success bool                `json:"success"`
				Data    base.GetAllResponse `json:"data"`
			}
			err := json.NewDecoder(resp.Body).Decode(&respStruct)
			if err != nil {
				t.Errorf("Failed to parse response JSON: %v", err)
			}
			if !respStruct.Success {
				t.Error("Expected success")
			}
			if respStruct.Message != response.MessageSuccessLoaded {
				t.Error("Expected message to be equal")
			}
		}
	}
}

func TestPermGetAll(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}

	testCases := []struct {
		caseName string
		respCode int
		payload  base.RequestGetAll
	}{
		{
			caseName: "success get -1",
			respCode: http.StatusOK,
			payload: base.RequestGetAll{
				Limit: 10,
				Page:  1,
			},
		},
		{
			caseName: "success get -2",
			respCode: http.StatusOK,
			payload: base.RequestGetAll{
				Limit: 100,
				Page:  2,
			},
		},
		{
			caseName: "failed get: invalid payload",
			respCode: http.StatusBadRequest,
			payload: base.RequestGetAll{
				Limit: -1,
				Page:  -1,
			},
		},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/permission?page=%d&limit=%d", tc.payload.Page, tc.payload.Limit), nil)
		app := fiber.New()
		app.Get("/permission", permController.GetAll)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(constants.ShouldNotErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal")
		}
		if resp.StatusCode == http.StatusOK {
			var respStruct struct {
				Message string              `json:"message"`
				Success bool                `json:"success"`
				Data    base.GetAllResponse `json:"data"`
			}
			err := json.NewDecoder(resp.Body).Decode(&respStruct)
			if err != nil {
				t.Errorf("Failed to parse response JSON: %v", err)
			}
			if !respStruct.Success {
				t.Error("Expected success")
			}
			if respStruct.Message != response.MessageSuccessLoaded {
				t.Error("Expected message to be equal")
			}
			if len(respStruct.Data.Data.([]any)) > tc.payload.Limit {
				t.Error("should less or equal", len(respStruct.Data.Data.([]any)))
			}
		}
	}
}

func TestPermUpdate(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}

	// create 1 permission
	permID, createErr := permService.Create(ctx, model.PermissionCreate{
		Name:        strings.ToLower(helper.RandomString(12)),
		Description: "description-of-example-permission-001",
	})
	if createErr != nil || permID < 1 {
		t.Fatal("should not error while creating permission")
	}

	defer func() {
		permService.Delete(ctx, permID)
	}()

	testCases := []struct {
		caseName string
		respCode int
		permID   int
		payload  model.PermissionUpdate
	}{
		{
			caseName: "success update -1",
			respCode: http.StatusNoContent,
			permID:   permID,
			payload: model.PermissionUpdate{
				ID:          permID,
				Name:        helper.RandomString(12),
				Description: helper.RandomString(20),
			},
		},
		{
			caseName: "success update -2",
			respCode: http.StatusNoContent,
			permID:   permID,
			payload: model.PermissionUpdate{
				ID:          permID,
				Name:        helper.RandomString(12),
				Description: helper.RandomString(20),
			},
		},
		{
			caseName: "failed update: invalid name/description",
			respCode: http.StatusBadRequest,
			permID:   permID,
			payload: model.PermissionUpdate{
				ID:          permID,
				Name:        "",
				Description: "",
			},
		},
		{
			caseName: "failed update: invalid id",
			respCode: http.StatusBadRequest,
			permID:   -10,
		},
		{
			caseName: "failed update: data not found",
			respCode: http.StatusNotFound,
			permID:   permID + 99,
			payload: model.PermissionUpdate{
				ID:          permID + 99,
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
		url := fmt.Sprintf(appURL+"permission/%d", tc.permID)
		req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(jsonObject))
		if err != nil {
			t.Error(constants.ShouldNotErr, err.Error())
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Put("/permission/:id", permController.Update)
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
			perm, getErr := permService.GetByID(ctx, permID)
			if getErr != nil || perm == nil {
				t.Error("should not error while get permission")
			}
			if perm.Name != strings.ToLower(tc.payload.Name) ||
				perm.Description != tc.payload.Description {
				t.Error("should equal")
			}
		}
	}
}

func TestPermDelete(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}

	// create 1 permission
	permID, createErr := permService.Create(ctx, model.PermissionCreate{
		Name:        strings.ToLower(helper.RandomString(12)),
		Description: "description-of-example-permission-001",
	})
	if createErr != nil || permID < 1 {
		t.Fatal("should not error while creating permission")
	}

	defer func() {
		permService.Delete(ctx, permID)
	}()

	testCases := []struct {
		caseName string
		respCode int
		permID   int
	}{
		{
			caseName: "success get -1",
			respCode: http.StatusNoContent,
			permID:   permID,
		},
		{
			caseName: "failed: not found / already deleted",
			respCode: http.StatusNotFound,
			permID:   permID,
		},
		{
			caseName: "failed: not found",
			respCode: http.StatusNotFound,
			permID:   permID + 100,
		},
		{
			caseName: "failed: invalid id",
			respCode: http.StatusBadRequest,
			permID:   -10,
		},
	}

	for _, tc := range testCases {
		url := appURL + "permission/" + strconv.Itoa(tc.permID)
		req, httpReqErr := http.NewRequest(http.MethodDelete, url, nil)
		if httpReqErr != nil || req == nil {
			t.Fatal(constants.ShouldNotNil)
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Delete("/permission/:id", permController.Delete)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(constants.ShouldNotErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal, want", tc.respCode, "but got", resp.StatusCode)
		}
	}
}

func createUserAndToken() (userID int, token string) {
	permService := service.NewPermissionService()
	roleService := roleService.NewRoleService(permService)
	userSvc := userService.NewUserService(roleService)
	userCtr := userController.NewUserController(userSvc)

	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	if ctr == nil || c == nil || ctx == nil {
		log.Fatal(constants.ShouldNotNil)
	}

	createdUser := model.UserRegister{
		Name:     helper.RandomString(10),
		Email:    helper.RandomEmail(),
		Password: helper.RandomString(10),
		RoleID:   1, // admin
	}
	userID, createErr := userSvc.Register(ctx, createdUser)
	if createErr != nil || userID <= 0 {
		log.Fatal("should success create user, user failed to create")
	}
	userByID, getErr := userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		log.Fatal("should success get user by id")
	}
	vCode := userByID.VerificationCode
	if vCode == nil || userByID.ActivatedAt != nil {
		log.Fatal("user should inactivate for now, but its get activated/ nulling vCode")
	}

	verifyErr := userSvc.Verification(ctx, model.UserVerificationCode{
		Code:  *userByID.VerificationCode,
		Email: userByID.Email,
	})
	if verifyErr != nil {
		log.Fatal(constants.ShouldNotErr)
	}
	userByID = nil
	userByID, getErr = userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		log.Fatal("should success get user by id")
	}
	if userByID.VerificationCode != nil || userByID.ActivatedAt == nil {
		log.Fatal("user should active for now, verification code should nil")
	}

	userToken, loginErr := userSvc.Login(ctx, model.UserLogin{
		Email:    createdUser.Email,
		Password: createdUser.Password,
		IP:       helper.RandomIPAddress(),
	})
	if userToken == "" || loginErr != nil {
		log.Fatal("login should success")
	}

	return userID, userToken
}
