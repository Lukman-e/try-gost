package repository

import (
	"context"
	"sync"

	"github.com/Lukman-e/try-gost/database/connector"
	"github.com/Lukman-e/try-gost/domain/base"
	"github.com/Lukman-e/try-gost/domain/entity"
	"gorm.io/gorm"
)

type PermissionRepository interface {
	// Create adds a new permission to the repository.
	Create(ctx context.Context, permission entity.Permission) (id int, err error)

	// GetByID retrieves a permission by its unique identifier.
	GetByID(ctx context.Context, id int) (permission *entity.Permission, err error)

	// GetByName retrieves a permission by its name.
	GetByName(ctx context.Context, name string) (permission *entity.Permission, err error)

	// GetAll retrieves all permissions based on a filter for pagination.
	GetAll(ctx context.Context, filter base.RequestGetAll) (permissions []entity.Permission, total int, err error)

	// Update modifies permission information in the repository.
	Update(ctx context.Context, permission entity.Permission) (err error)

	// Delete removes a permission from the repository by its ID.
	Delete(ctx context.Context, id int) (err error)
}

type PermissionRepositoryImpl struct {
	db *gorm.DB
}

var (
	permissionRepositoryImpl     *PermissionRepositoryImpl
	permissionRepositoryImplOnce sync.Once
)

func NewPermissionRepository() PermissionRepository {
	permissionRepositoryImplOnce.Do(func() {
		permissionRepositoryImpl = &PermissionRepositoryImpl{
			db: connector.LoadDatabase(),
		}
	})
	return permissionRepositoryImpl
}

func (repo *PermissionRepositoryImpl) Create(ctx context.Context, permission entity.Permission) (id int, err error) {
	err = repo.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Create(&permission)
		if res.Error != nil {
			tx.Rollback()
			return res.Error
		}
		id = permission.ID
		return nil
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (repo *PermissionRepositoryImpl) GetByID(ctx context.Context, id int) (permission *entity.Permission, err error) {
	permission = &entity.Permission{}
	result := repo.db.Where("id = ?", id).First(&permission)
	if result.Error != nil {
		return nil, result.Error
	}
	return permission, nil
}

func (repo *PermissionRepositoryImpl) GetByName(ctx context.Context, name string) (permission *entity.Permission, err error) {
	permission = &entity.Permission{}
	result := repo.db.Where("name = ?", name).First(&permission)
	if result.Error != nil {
		return nil, result.Error
	}
	return permission, nil
}

func (repo *PermissionRepositoryImpl) GetAll(ctx context.Context, filter base.RequestGetAll) (permissions []entity.Permission, total int, err error) {
	var count int64
	args := []interface{}{"%" + filter.Keyword + "%"}
	cond := "name LIKE ?"
	result := repo.db.Where(cond, args...).Find(&permissions)
	count = result.RowsAffected
	if result.Error != nil {
		return nil, 0, result.Error
	}
	permissions = []entity.Permission{}
	skip := int64(filter.Limit * (filter.Page - 1))
	limit := int64(filter.Limit)
	result = repo.db.Where(cond, args...).Limit(int(limit)).Offset(int(skip)).Find(&permissions)
	if result.Error != nil {
		return nil, 0, result.Error
	}
	total = int(count)
	return permissions, total, nil
}

func (repo *PermissionRepositoryImpl) Update(ctx context.Context, permission entity.Permission) (err error) {
	err = repo.db.Transaction(func(tx *gorm.DB) error {
		var oldData entity.Permission
		result := tx.Where("id = ?", permission.ID).First(&oldData)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}

		oldData.Name = permission.Name
		oldData.Description = permission.Description
		oldData.UpdatedAt = permission.UpdatedAt
		result = tx.Save(&oldData)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
		return nil
	})

	return err
}

func (repo *PermissionRepositoryImpl) Delete(ctx context.Context, id int) (err error) {
	deleted := entity.Permission{}
	result := repo.db.Where("id = ?", id).Delete(&deleted)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
