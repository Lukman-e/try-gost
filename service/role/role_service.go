package service

import (
	"context"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/Lukman-e/try-gost/domain/base"
	"github.com/Lukman-e/try-gost/domain/entity"
	"github.com/Lukman-e/try-gost/domain/model"
	repository "github.com/Lukman-e/try-gost/repository/role"
	permService "github.com/Lukman-e/try-gost/service/permission"
)

type RoleService interface {

	// Create func create one role.
	Create(ctx context.Context, data model.RoleCreate) (id int, err error)

	// ConnectPermissions func connect one role with one or more permissions.
	ConnectPermissions(ctx context.Context, data model.RoleConnectToPermissions) (err error)

	// GetByID func get one role.
	GetByID(ctx context.Context, id int) (role *entity.Role, err error)

	// GetAll func get some roles.
	GetAll(ctx context.Context, filter base.RequestGetAll) (roles []model.RoleResponse, total int, err error)

	// Update func update one role.
	Update(ctx context.Context, data model.RoleUpdate) (err error)

	// Delete func delete one role.
	Delete(ctx context.Context, id int) (err error)
}

type RoleServiceImpl struct {
	repository        repository.RoleRepository
	servicePermission permService.PermissionService
}

var (
	roleServiceImpl     *RoleServiceImpl
	roleServiceImplOnce sync.Once
)

const roleNotFound = "role/s not found"

func NewRoleService(servicePermission permService.PermissionService) RoleService {
	roleServiceImplOnce.Do(func() {
		roleServiceImpl = &RoleServiceImpl{
			repository:        repository.NewRoleRepository(),
			servicePermission: servicePermission,
		}
	})
	return roleServiceImpl
}

func (svc *RoleServiceImpl) Create(ctx context.Context, data model.RoleCreate) (id int, err error) {
	data.Name = strings.ToLower(data.Name)
	for _, id := range data.PermissionsID {
		permission, getErr := svc.servicePermission.GetByID(ctx, id)
		if getErr != nil || permission == nil {
			return 0, fiber.NewError(fiber.StatusNotFound, "one of permissions isn't found")
		}
	}
	role, getErr := svc.repository.GetByName(ctx, data.Name)
	if getErr == nil || role != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "role name has been used")
	}

	entityRole := entity.Role{
		Name:        data.Name,
		Description: data.Description,
	}
	entityRole.SetCreateTime()
	id, err = svc.repository.Create(ctx, entityRole, data.PermissionsID)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (svc *RoleServiceImpl) ConnectPermissions(ctx context.Context, data model.RoleConnectToPermissions) (err error) {
	role, getErr := svc.repository.GetByID(ctx, data.RoleID)
	if getErr != nil {
		if getErr == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, roleNotFound)
		}
		return getErr
	}
	if role == nil {
		return fiber.NewError(fiber.StatusNotFound, roleNotFound)
	}
	for _, id := range data.PermissionsID {
		permission, getErr := svc.servicePermission.GetByID(ctx, id)
		if getErr != nil || permission == nil {
			return fiber.NewError(fiber.StatusNotFound, "one of permissions isn't found")
		}
	}

	connectErr := svc.repository.ConnectToPermission(ctx, data.RoleID, data.PermissionsID)
	if connectErr != nil {
		return connectErr
	}
	return nil
}

func (svc *RoleServiceImpl) GetByID(ctx context.Context, id int) (role *entity.Role, err error) {
	role, err = svc.repository.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.NewError(fiber.StatusNotFound, roleNotFound)
		}
		return nil, err
	}
	if role == nil {
		return nil, fiber.NewError(fiber.StatusNotFound, roleNotFound)
	}
	return role, nil
}

func (svc *RoleServiceImpl) GetAll(ctx context.Context, filter base.RequestGetAll) (roles []model.RoleResponse, total int, err error) {
	roleEntities, total, err := svc.repository.GetAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	roles = []model.RoleResponse{}
	for _, roleEntity := range roleEntities {
		newRole := model.RoleResponse{
			ID:          roleEntity.ID,
			Name:        roleEntity.Name,
			Description: roleEntity.Description,
		}
		roles = append(roles, newRole)
	}
	return roles, total, nil
}

func (svc *RoleServiceImpl) Update(ctx context.Context, data model.RoleUpdate) (err error) {
	data.Name = strings.ToLower(data.Name)
	roleByName, getErr := svc.repository.GetByName(ctx, data.Name)
	if getErr != nil && getErr != gorm.ErrRecordNotFound {
		return getErr
	}
	if roleByName != nil && roleByName.ID != data.ID {
		return fiber.NewError(fiber.StatusBadRequest, "role name has been used")
	}

	roleByID, getErr := svc.repository.GetByID(ctx, data.ID)
	if getErr != nil {
		if getErr == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, roleNotFound)
		}
		return getErr
	}
	if roleByID == nil {
		return fiber.NewError(fiber.StatusNotFound, roleNotFound)
	}

	entityRole := entity.Role{
		ID:          data.ID,
		Name:        data.Name,
		Description: data.Description,
	}
	entityRole.SetUpdateTime()
	err = svc.repository.Update(ctx, entityRole)
	if err != nil {
		return err
	}
	return nil
}

func (svc *RoleServiceImpl) Delete(ctx context.Context, id int) (err error) {
	role, getErr := svc.repository.GetByID(ctx, id)
	if getErr != nil {
		if getErr == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, roleNotFound)
		}
		return getErr
	}
	if role == nil {
		return fiber.NewError(fiber.StatusNotFound, roleNotFound)
	}
	err = svc.repository.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}
