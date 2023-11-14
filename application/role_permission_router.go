// 📌 Origin Github Repository: https://github.com/Lukmanern<slash>gost

// 🔍 README
// Role-Permission Routes provides des create, read (get & getAll), update, and
// delete functionalities for Role and Permission entities including connecting
// both of them. This routes can be access by user that has admin-role (see database/migration).

package application

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Lukman-e/try-gost/internal/middleware"
	"github.com/Lukman-e/try-gost/internal/rbac"

	permCtr "github.com/Lukman-e/try-gost/controller/permission"
	roleCtr "github.com/Lukman-e/try-gost/controller/role"
	permSvc "github.com/Lukman-e/try-gost/service/permission"
	roleSvc "github.com/Lukman-e/try-gost/service/role"
)

var (
	roleService    roleSvc.RoleService
	roleController roleCtr.RoleController

	permissionService    permSvc.PermissionService
	permissionController permCtr.PermissionController
)

func getRolePermissionRoutes(router fiber.Router) {
	jwtHandler := middleware.NewJWTHandler()

	permissionService = permSvc.NewPermissionService()
	permissionController = permCtr.NewPermissionController(permissionService)
	permissionRouter := router.Group("permission").Use(jwtHandler.IsAuthenticated)

	// create-permission is unused
	permissionRouter.Post("", jwtHandler.CheckHasPermission(rbac.PermCreatePermission.ID), permissionController.Create)
	permissionRouter.Get("", jwtHandler.CheckHasPermission(rbac.PermViewPermission.ID), permissionController.GetAll)
	permissionRouter.Get(":id", jwtHandler.CheckHasPermission(rbac.PermViewPermission.ID), permissionController.Get)
	permissionRouter.Put(":id", jwtHandler.CheckHasPermission(rbac.PermUpdatePermission.ID), permissionController.Update)
	permissionRouter.Delete(":id", jwtHandler.CheckHasPermission(rbac.PermDeletePermission.ID), permissionController.Delete)

	roleService = roleSvc.NewRoleService(permissionService)
	roleController = roleCtr.NewRoleController(roleService)
	roleRouter := router.Group("role").Use(jwtHandler.IsAuthenticated)

	roleRouter.Post("", jwtHandler.CheckHasPermission(rbac.PermCreateRole.ID), roleController.Create)
	roleRouter.Post("connect", jwtHandler.CheckHasPermission(rbac.PermCreateRole.ID), roleController.Connect)
	roleRouter.Get("", jwtHandler.CheckHasPermission(rbac.PermViewRole.ID), roleController.GetAll)
	roleRouter.Get(":id", jwtHandler.CheckHasPermission(rbac.PermViewRole.ID), roleController.Get)
	roleRouter.Put(":id", jwtHandler.CheckHasPermission(rbac.PermUpdateRole.ID), roleController.Update)
	roleRouter.Delete(":id", jwtHandler.CheckHasPermission(rbac.PermDeleteRole.ID), roleController.Delete)
}
