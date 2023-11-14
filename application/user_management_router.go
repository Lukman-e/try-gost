// 📌 Origin Github Repository: https://github.com/Lukmanern<slash>gost

// 🔍 README
// User Management Routes provides create, read (get & getAll), update, and
// delete functionalities for user data management without JWT authentication
// in header. ⚠️ So, don't forget to commented on the line of code that routes
// getUserManagementRoutes in the app.go file.

package application

import (
	"github.com/gofiber/fiber/v2"

	controller "github.com/Lukman-e/try-gost/controller/user_management"
	service "github.com/Lukman-e/try-gost/service/user_management"
)

var (
	userDevService    service.UserManagementService
	userDevController controller.UserManagementController
)

func getUserManagementRoutes(router fiber.Router) {
	userDevService = service.NewUserManagementService()
	userDevController = controller.NewUserManagementController(userDevService)

	userDevRoute := router.Group("user-management")
	userDevRoute.Post("create", userDevController.Create)
	userDevRoute.Get("", userDevController.GetAll)
	userDevRoute.Get(":id", userDevController.Get)
	userDevRoute.Put(":id", userDevController.Update)
	userDevRoute.Delete(":id", userDevController.Delete)
}
