package controller

import (
	"math"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/Lukman-e/try-gost/domain/base"
	"github.com/Lukman-e/try-gost/domain/model"
	"github.com/Lukman-e/try-gost/internal/constants"
	"github.com/Lukman-e/try-gost/internal/response"
	service "github.com/Lukman-e/try-gost/service/role"
)

type RoleController interface {

	// Create func creates a new role
	Create(c *fiber.Ctx) error

	// Connect func connects a role with some permissions
	// and storing data in role_has_permissions table
	Connect(c *fiber.Ctx) error

	// Get func gets a role
	Get(c *fiber.Ctx) error

	// GetAll func gets some roles
	GetAll(c *fiber.Ctx) error

	// Update func updates a role
	Update(c *fiber.Ctx) error

	// Delete func deletes a role
	Delete(c *fiber.Ctx) error
}

type RoleControllerImpl struct {
	service service.RoleService
}

var (
	roleControllerImpl     *RoleControllerImpl
	roleControllerImplOnce sync.Once
)

func NewRoleController(service service.RoleService) RoleController {
	roleControllerImplOnce.Do(func() {
		roleControllerImpl = &RoleControllerImpl{
			service: service,
		}
	})
	return roleControllerImpl
}

func (ctr *RoleControllerImpl) Create(c *fiber.Ctx) error {
	var role model.RoleCreate
	if err := c.BodyParser(&role); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}
	// hashmap
	idCheckers := make(map[int]bool)
	for _, id := range role.PermissionsID {
		if id < 1 {
			return response.BadRequest(c, "One of the permission IDs is invalid")
		}
		if idCheckers[id] {
			return response.BadRequest(c, "Permission IDs contain the same value")
		}
		idCheckers[id] = true
	}
	validate := validator.New()
	if err := validate.Struct(&role); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}

	ctx := c.Context()
	id, createErr := ctr.service.Create(ctx, role)
	if createErr != nil {
		fiberErr, ok := createErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, constants.ServerErr+createErr.Error())
	}
	data := map[string]any{
		"id": id,
	}
	return response.SuccessCreated(c, data)
}

func (ctr *RoleControllerImpl) Connect(c *fiber.Ctx) error {
	var role model.RoleConnectToPermissions
	if err := c.BodyParser(&role); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}
	// hashmap
	idCheckers := make(map[int]bool)
	for _, id := range role.PermissionsID {
		if id < 1 {
			return response.BadRequest(c, "One of the permission IDs is invalid")
		}
		if idCheckers[id] {
			return response.BadRequest(c, "Permission IDs contain the same value")
		}
		idCheckers[id] = true
	}
	validate := validator.New()
	if err := validate.Struct(&role); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}

	ctx := c.Context()
	connectErr := ctr.service.ConnectPermissions(ctx, role)
	if connectErr != nil {
		fiberErr, ok := connectErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, constants.ServerErr+connectErr.Error())
	}
	return response.SuccessCreated(c, "role and permissions success connected")
}

func (ctr *RoleControllerImpl) Get(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return response.BadRequest(c, constants.InvalidID)
	}

	ctx := c.Context()
	role, getErr := ctr.service.GetByID(ctx, id)
	if getErr != nil {
		fiberErr, ok := getErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, constants.ServerErr+getErr.Error())
	}
	return response.SuccessLoaded(c, role)
}

func (ctr *RoleControllerImpl) GetAll(c *fiber.Ctx) error {
	request := base.RequestGetAll{
		Page:    c.QueryInt("page", 1),
		Limit:   c.QueryInt("limit", 20),
		Keyword: c.Query("search"),
		Sort:    c.Query("sort"),
	}
	if request.Page <= 0 || request.Limit <= 0 {
		return response.BadRequest(c, "invalid page or limit value")
	}

	ctx := c.Context()
	roles, total, getErr := ctr.service.GetAll(ctx, request)
	if getErr != nil {
		return response.Error(c, constants.ServerErr+getErr.Error())
	}

	data := make([]interface{}, len(roles))
	for i := range roles {
		data[i] = roles[i]
	}
	responseData := base.GetAllResponse{
		Meta: base.PageMeta{
			Total: total,
			Pages: int(math.Ceil(float64(total) / float64(request.Limit))),
			Page:  request.Page,
		},
		Data: data,
	}
	return response.SuccessLoaded(c, responseData)
}

func (ctr *RoleControllerImpl) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return response.BadRequest(c, constants.InvalidID)
	}
	var role model.RoleUpdate
	role.ID = id
	if err := c.BodyParser(&role); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(&role); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}

	ctx := c.Context()
	updateErr := ctr.service.Update(ctx, role)
	if updateErr != nil {
		fiberErr, ok := updateErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, constants.ServerErr+updateErr.Error())
	}
	return response.SuccessNoContent(c)
}

func (ctr *RoleControllerImpl) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return response.BadRequest(c, constants.InvalidID)
	}

	ctx := c.Context()
	deleteErr := ctr.service.Delete(ctx, id)
	if deleteErr != nil {
		fiberErr, ok := deleteErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, constants.ServerErr+deleteErr.Error())
	}
	return response.SuccessNoContent(c)
}
