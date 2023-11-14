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
	repository "github.com/Lukman-e/try-gost/repository/permission"
)

type PermissionService interface {
	// Create func create one permission.
	Create(ctx context.Context, permission model.PermissionCreate) (id int, err error)

	// GetByID func get one permission by ID.
	GetByID(ctx context.Context, id int) (permission *model.PermissionResponse, err error)

	// GetAll func get some permissions with payload.
	GetAll(ctx context.Context, filter base.RequestGetAll) (permissions []model.PermissionResponse, total int, err error)

	// Update func update one permission by ID and payload.
	Update(ctx context.Context, permission model.PermissionUpdate) (err error)

	// Delete func delete one permission by ID.
	Delete(ctx context.Context, id int) (err error)
}

type PermissionServiceImpl struct {
	repository repository.PermissionRepository
}

var (
	permissionServiceImpl     *PermissionServiceImpl
	permissionServiceImplOnce sync.Once
)

const permNotFound = "permission/s not found"

func NewPermissionService() PermissionService {
	permissionServiceImplOnce.Do(func() {
		permissionServiceImpl = &PermissionServiceImpl{
			repository: repository.NewPermissionRepository(),
		}
	})
	return permissionServiceImpl
}

func (svc *PermissionServiceImpl) Create(ctx context.Context, permission model.PermissionCreate) (id int, err error) {
	permission.Name = strings.ToLower(permission.Name)

	checkPermission, getErr := svc.repository.GetByName(ctx, permission.Name)
	if getErr == nil || checkPermission != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "permission name has been used")
	}
	entityPermission := entity.Permission{
		Name:        permission.Name,
		Description: permission.Description,
	}
	entityPermission.SetCreateTime()
	id, err = svc.repository.Create(ctx, entityPermission)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (svc *PermissionServiceImpl) GetByID(ctx context.Context, id int) (permission *model.PermissionResponse, err error) {
	permissionEntity, getErr := svc.repository.GetByID(ctx, id)
	if getErr != nil {
		if getErr == gorm.ErrRecordNotFound {
			return nil, fiber.NewError(fiber.StatusNotFound, permNotFound)
		}
		return nil, getErr
	}
	if permissionEntity == nil {
		return nil, fiber.NewError(fiber.StatusNotFound, permNotFound)
	}

	permission = &model.PermissionResponse{
		ID:          permissionEntity.ID,
		Name:        permissionEntity.Name,
		Description: permissionEntity.Description,
	}
	return permission, nil
}

func (svc *PermissionServiceImpl) GetAll(ctx context.Context, filter base.RequestGetAll) (permissions []model.PermissionResponse, total int, err error) {
	permissionEntities, total, err := svc.repository.GetAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	permissions = []model.PermissionResponse{}
	for _, permissionEntity := range permissionEntities {
		newPermission := model.PermissionResponse{
			ID:          permissionEntity.ID,
			Name:        permissionEntity.Name,
			Description: permissionEntity.Description,
		}

		permissions = append(permissions, newPermission)
	}
	return permissions, total, nil
}

func (svc *PermissionServiceImpl) Update(ctx context.Context, data model.PermissionUpdate) (err error) {
	data.Name = strings.ToLower(data.Name)
	permissionByName, getErr := svc.repository.GetByName(ctx, data.Name)
	if getErr != nil && getErr != gorm.ErrRecordNotFound {
		return getErr
	}
	if permissionByName != nil && permissionByName.ID != data.ID {
		return fiber.NewError(fiber.StatusBadRequest, "permission name has been used")
	}

	permissionByID, getErr := svc.repository.GetByID(ctx, data.ID)
	if getErr != nil {
		if getErr == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, permNotFound)
		}
		return getErr
	}
	if permissionByID == nil {
		return fiber.NewError(fiber.StatusNotFound, permNotFound)
	}

	entityRole := entity.Permission{
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

func (svc *PermissionServiceImpl) Delete(ctx context.Context, id int) (err error) {
	permission, getErr := svc.repository.GetByID(ctx, id)
	if getErr != nil {
		if getErr == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, permNotFound)
		}
		return getErr
	}
	if permission == nil {
		return fiber.NewError(fiber.StatusNotFound, permNotFound)
	}
	err = svc.repository.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}
