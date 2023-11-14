// 📌 Origin Github Repository: https://github.com/Lukmanern<slash>gost

// 🔍 README
// User Routes provides some features and action that user can use.
// User Routes provide the typical web application authentication flow,
// such as registration, sending verification codes, and verifying accounts
// with a verification code.

package application

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Lukman-e/try-gost/internal/middleware"

	controller "github.com/Lukman-e/try-gost/controller/user"
	service "github.com/Lukman-e/try-gost/service/user"

	permSvc "github.com/Lukman-e/try-gost/service/permission"
	roleSvc "github.com/Lukman-e/try-gost/service/role"
)

var (
	userPermService permSvc.PermissionService
	userRoleService roleSvc.RoleService
	userService     service.UserService
	userController  controller.UserController
)

func getUserRoutes(router fiber.Router) {
	userPermService = permSvc.NewPermissionService()
	userRoleService = roleSvc.NewRoleService(userPermService)
	userService = service.NewUserService(userRoleService)
	userController = controller.NewUserController(userService)
	jwtHandler := middleware.NewJWTHandler()

	userRoute := router.Group("user")
	userRoute.Post("login", userController.Login)
	userRoute.Post("register", userController.Register)
	userRoute.Post("verification", userController.AccountActivation)
	userRoute.Post("request-delete", userController.DeleteAccountActivation)
	userRoute.Post("forget-password", userController.ForgetPassword)
	userRoute.Post("reset-password", userController.ResetPassword)

	userRouteAuth := userRoute.Use(jwtHandler.IsAuthenticated)
	userRouteAuth.Post("logout", userController.Logout)
	userRouteAuth.Get("my-profile", userController.MyProfile)
	userRouteAuth.Put("profile-update", userController.UpdateProfile)
	userRouteAuth.Post("update-password", userController.UpdatePassword)
}
