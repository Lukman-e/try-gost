// 📌 Origin Github Repository: https://github.com/Lukmanern<slash>gost

// 🔍 README
// Application package configures middleware, error management, and
// handles OS signals for gracefully stopping the server when receiving
// an interrupt signal. This package provides routes related to user
// management and role-based access control (RBAC). And so on.

package application

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/Lukman-e/try-gost/database/connector"
	"github.com/Lukman-e/try-gost/internal/env"
)

var (
	port int

	// Create a new fiber instance with custom config
	router = fiber.New(fiber.Config{
		// Override default error handler
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			// Status code defaults to 500
			code := fiber.StatusInternalServerError

			// Retrieve the custom status code
			// if it's a *fiber.Error
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}

			// Send custom error page
			err = ctx.Status(code).JSON(fiber.Map{
				"message": e.Message,
			})
			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).
					SendString("Internal Server Error")
			}
			return nil
		},
		// memory management
		// ReduceMemoryUsage: true,
		// ReadBufferSize: 5120,
	})
)

func setup() {
	// Check env and database
	env.ReadConfig("./.env")
	config := env.Configuration()
	privKey := config.GetPrivateKey()
	pubKey := config.GetPublicKey()
	if privKey == nil || pubKey == nil {
		log.Fatal("private and public keys are not valid or not found")
	}
	port = config.AppPort

	connector.LoadDatabase()
	connector.LoadRedisCache()
}

func RunApp() {
	setup()
	router.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))
	router.Use(logger.New())
	// Custom File Writer
	_ = os.MkdirAll("./log", os.ModePerm)
	fileName := fmt.Sprintf("./log/%s.log", time.Now().Format("20060102"))
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer file.Close()
	router.Use(logger.New(logger.Config{
		Output: file,
	}))

	// Create channel for idle connections.
	idleConnsClosed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt) // Catch OS signals.
		<-sigint

		// Received an interrupt signal, shutdown.
		// ctrl+c
		if err := router.Shutdown(); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("Oops... Server is not shutting down! Reason: %v", err)
		}

		close(idleConnsClosed)
	}()

	getUserManagementRoutes(router) // user CRUD without auth ⚠️
	getDevopmentRouter(router)      // experimental without auth ⚠️
	getUserRoutes(router)           // user with auth
	getRolePermissionRoutes(router) // RBAC CRUD with auth

	if err := router.Listen(fmt.Sprintf(":%d", port)); err != nil {
		log.Printf("Oops... Server is not running! Reason: %v", err)
	}

	<-idleConnsClosed
}
