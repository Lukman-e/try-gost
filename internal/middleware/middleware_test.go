package middleware

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/Lukman-e/try-gost/internal/constants"
	"github.com/Lukman-e/try-gost/internal/env"
	"github.com/Lukman-e/try-gost/internal/helper"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

type GenTokenParams struct {
	ID      int
	Email   string
	Role    string
	Per     map[int]int
	Exp     time.Time
	wantErr bool
}

var (
	params GenTokenParams
)

func init() {
	filepath := "./../../.env"
	env.ReadConfig(filepath)

	timeNow := time.Now()
	params = GenTokenParams{
		ID:    1,
		Email: helper.RandomEmail(),
		Role:  "test-role",
		Per: map[int]int{
			1: 1,
			2: 1,
			3: 1,
			4: 1,
			5: 1,
			6: 1,
			7: 1,
			8: 1,
		},
		Exp:     timeNow.Add(5 * time.Minute),
		wantErr: false,
	}
}

func TestNewJWTHandler(t *testing.T) {
	jwtHandler := NewJWTHandler()
	if jwtHandler.publicKey == nil {
		t.Errorf("Public key parsing should have failed")
	}

	if jwtHandler.privateKey == nil {
		t.Errorf("Private key parsing should have failed")
	}
}

func TestGenerateClaims(t *testing.T) {
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(1, params.Email, params.Role, params.Per, params.Exp)
	if err != nil || token == "" {
		t.Fatal("should not error")
	}

	testCases := []struct {
		token    string
		isResNil bool
	}{
		{
			token:    "",
			isResNil: true,
		},
		{
			token:    token,
			isResNil: false,
		},
	}

	for _, tc := range testCases {
		claims := jwtHandler.GenerateClaims(tc.token)
		if claims == nil && !tc.isResNil {
			t.Error("should not nil")
		}
		if claims != nil && tc.isResNil {
			t.Error("should nil")
		}
	}
}

func TestJWTHandlerInvalidateToken(t *testing.T) {
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(params.ID, params.Email, params.Role, params.Per, params.Exp)
	if err != nil {
		t.Error("error while generating token")
	}
	if token == "" {
		t.Error("error : token void")
	}

	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	invalidErr1 := jwtHandler.InvalidateToken(c)
	if invalidErr1 != nil {
		t.Error("Should error: Expected error for no token")
	}

	c.Request().Header.Add(fiber.HeaderAuthorization, "Bearer "+token)
	invalidErr2 := jwtHandler.InvalidateToken(c)
	if invalidErr2 != nil {
		t.Error("Expected no error for a valid token, but got an error.")
	}
}

func TestJWTHandlerIsBlacklisted(t *testing.T) {
	jwtHandler := NewJWTHandler()
	cookie, err := jwtHandler.GenerateJWT(1000,
		helper.RandomEmail(), "example-role",
		params.Per, time.Now().Add(1*time.Hour))
	if err != nil {
		t.Error("generate cookie/token should not error")
	}

	type args struct {
		cookie string
	}
	tests := []struct {
		name string
		j    JWTHandler
		args args
		want bool
	}{
		{
			name: "check : false blacklisted",
			j:    *jwtHandler,
			args: args{cookie: cookie},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.j.IsBlacklisted(tt.args.cookie); got != tt.want {
				t.Errorf("JWTHandler.IsBlacklisted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJWTHandlerIsAuthenticated(t *testing.T) {
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(params.ID, params.Email, params.Role, params.Per, params.Exp)
	if err != nil {
		t.Error("error while generating token")
	}
	if token == "" {
		t.Error("error : token void")
	}

	func() {
		jwtHandler1 := NewJWTHandler()
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		jwtHandler1.IsAuthenticated(c)
		c.Status(fiber.StatusUnauthorized)
		if c.Context().Response.StatusCode() != fiber.StatusUnauthorized {
			t.Error("Expected error for no token")
		}
	}()

	func() {
		defer func() {
			r := recover()
			if r != nil {
				t.Error("should not panic")
			}
		}()
		jwtHandler3 := NewJWTHandler()
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		c.Request().Header.Add(fiber.HeaderAuthorization, " "+token)
		c.Status(fiber.StatusUnauthorized)
		jwtHandler3.IsAuthenticated(c)
		if c.Context().Response.StatusCode() != fiber.StatusUnauthorized {
			t.Error("Expected error for no token")
		}
	}()
}

func TestJWTHandlerHasPermission(t *testing.T) {
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(params.ID, params.Email, params.Role, params.Per, params.Exp)
	if err != nil {
		t.Error("Error while generating token:", err)
	}
	if token == "" {
		t.Error("Error: Token is empty")
	}

	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	c.Request().Header.Add(fiber.HeaderAuthorization, "Bearer "+token)
	jwtHandler.HasPermission(c, 25)
	if c.Response().Header.StatusCode() != fiber.StatusUnauthorized {
		t.Error("Should authorized")
	}
}

func TestJWTHandlerHasRole(t *testing.T) {
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(params.ID, params.Email, params.Role, params.Per, params.Exp)
	if err != nil {
		t.Error("Error while generating token:", err)
	}
	if token == "" {
		t.Error("Error: Token is empty")
	}

	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	c.Request().Header.Add(fiber.HeaderAuthorization, "Bearer "+token)
	jwtHandler.HasRole(c, "test-role")
	if c.Response().Header.StatusCode() != fiber.StatusUnauthorized {
		t.Error(constants.Unauthorized)
	}
}

func TestJWTHandlerCheckHasPermission(t *testing.T) {
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(params.ID, params.Email, params.Role, params.Per, params.Exp)
	if err != nil {
		t.Error("Error while generating token:", err)
	}
	if token == "" {
		t.Error("Error: Token is empty")
	}

	err2 := jwtHandler.CheckHasPermission(9999)
	if err2 == nil {
		t.Error(constants.Unauthorized)
	}
}

func TestJWTHandlerCheckHasRole(t *testing.T) {
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(params.ID, params.Email, params.Role, params.Per, params.Exp)
	if err != nil {
		t.Error("Error while generating token:", err)
	}
	if token == "" {
		t.Error("Error: Token is empty")
	}

	err2 := jwtHandler.CheckHasRole("permission-1")
	if err2 == nil {
		t.Error(constants.Unauthorized)
	}
}

func TestPermissionBitGroup(t *testing.T) {
	d := 8
	testCases := []struct {
		input  int
		result map[int]int
	}{
		{
			input: d,
			result: map[int]int{
				1: int(math.Pow(2, 7)),
			},
		},
		{
			input: 10 * d,
			result: map[int]int{
				10: int(math.Pow(2, 7)),
			},
		},
		{
			input: d + 7,
			result: map[int]int{
				2: int(math.Pow(2, 6)),
			},
		},
		{
			input: d,
			result: map[int]int{
				1: int(math.Pow(2, 7)),
			},
		},
	}

	for _, tc := range testCases {
		result := BuildBitGroups(tc.input)
		if !reflect.DeepEqual(result, tc.result) {
			t.Error("should same, but got", result, "want", tc.result)
		}
	}

	permIDs := make([]int, 0)
	for i := 1; i < 90; i++ {
		if i%2 != 0 {
			continue
		}
		permIDs = append(permIDs, i)
	}

	result := BuildBitGroups(permIDs...)
	for group, bits := range result {
		fmt.Printf("%d : %08b\n", group, bits)
	}
}

func TestCheckHasPermission(t *testing.T) {
	// user perms
	permIDs := make([]int, 0)
	for i := 1; i <= 19; i++ {
		permIDs = append(permIDs, i)
	}

	bitGroups := BuildBitGroups(permIDs...)
	for i := 1; i <= 30; i++ {
		fmt.Println(i, ":", CheckHasPermission(i, bitGroups))
	}
}
