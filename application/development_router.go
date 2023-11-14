// 📌 Origin Github Repository: https://github.com/Lukmanern<slash>gost

// 🔍 README
// Development Routes provides experimental/ developing/ testing
// for routes, middleware, connection and many more without JWT
// authentication in header. ⚠️ So, don't forget to commented
// on the line of code that routes getDevopmentRouter
// in the app.go file.

package application

import (
	"github.com/gofiber/fiber/v2"

	controller "github.com/Lukman-e/try-gost/controller/development"
)

var (
	devController controller.DevController
)

func getDevopmentRouter(router fiber.Router) {
	devController = controller.NewDevControllerImpl()
	// Developement 'helper' Process
	devRouter := router.Group("development")
	devRouter.Get("ping/db", devController.PingDatabase)
	devRouter.Get("ping/redis", devController.PingRedis)
	devRouter.Get("panic", devController.Panic)
	devRouter.Get("storing-to-redis", devController.StoringToRedis)
	devRouter.Get("get-from-redis", devController.GetFromRedis)
	devRouter.Post("upload-file", devController.UploadFile)
	devRouter.Post("get-files-list", devController.GetFilesList)
	devRouter.Delete("remove-file", devController.RemoveFile)

	// you should create new role named new-role-001 and new permission
	// named new-permission-001 from RBAC-endpoints to test these endpoints
	// jwtHandler := middleware.NewJWTHandler()
	// devRouterAuth := devRouter.Use(jwtHandler.IsAuthenticated)
	// devRouterAuth.Get("test-new-role",
	// 	jwtHandler.CheckHasRole("new-role-001"), devController.CheckNewRole)
	// devRouterAuth.Get("test-new-permission",
	// 	jwtHandler.CheckHasPermission(21), devController.CheckNewPermission)
}
