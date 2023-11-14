// Don't run test without -p 1
// Please check Makefile file
// or simply just run this : go test -p 1 ./application/...

package application

import (
	"context"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/Lukman-e/try-gost/internal/env"
	"github.com/Lukman-e/try-gost/internal/middleware"
	repository "github.com/Lukman-e/try-gost/repository/user"
)

var (
	jwtHandler *middleware.JWTHandler
	timeNow    time.Time
	userRepo   repository.UserRepository
	ctx        context.Context
	appURL     string
)

func init() {
	env.ReadConfig("./../.env")
	c := env.Configuration()
	appURL = c.AppURL

	jwtHandler = middleware.NewJWTHandler()
	timeNow = time.Now()
	userRepo = repository.NewUserRepository()
	ctx = context.Background()
}

func TestRunApp(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Error("should not panic", r)
		}
	}()

	go RunApp()
	time.Sleep(3 * time.Second)
}

func TestAppRouter(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Error("should not panic : ", r)
		}
	}()

	if router == nil {
		t.Error("Router should not be nil")
	}
	if router.Server() == nil {
		t.Error("Router's server should not be nil")
	}
	if router.Config().ReadBufferSize <= 0 {
		t.Error("Router's ReadBufferSize should be more than 0")
	}
	if router.Config().WriteBufferSize <= 0 {
		t.Error("Router's WriteBufferSize should be more than 0")
	}
	if router.Config().ServerHeader != "" {
		t.Error("Router's ServerHeader should be empty")
	}
	if router.Config().ProxyHeader != "" {
		t.Error("Router's ProxyHeader should be empty")
	}
	setup()
}

func TestRoutes(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Error("should not panic : ", r)
		}
	}()
	router := fiber.New()

	t.Run("getDevelopmentRouter", func(t *testing.T) {
		getDevopmentRouter(router)
	})

	t.Run("getRBACAuthRoutes", func(t *testing.T) {
		getRolePermissionRoutes(router)
	})

	t.Run("getUserAuthRoutes", func(t *testing.T) {
		getUserRoutes(router)
	})

	t.Run("getUserRoutes", func(t *testing.T) {
		getUserManagementRoutes(router)
	})
}
