// used by user auth service

package repository

import (
	"context"
	"sync"

	"gorm.io/gorm"

	"github.com/Lukman-e/try-gost/database/connector"
	"github.com/Lukman-e/try-gost/domain/base"
	"github.com/Lukman-e/try-gost/domain/entity"
)

type UserRepository interface {
	// Create adds a new user to the repository with a specified role.
	Create(ctx context.Context, user entity.User, roleID int) (id int, err error)

	// GetByID retrieves a user by their unique identifier.
	GetByID(ctx context.Context, id int) (user *entity.User, err error)

	// GetByEmail retrieves a user by their email address.
	GetByEmail(ctx context.Context, email string) (user *entity.User, err error)

	// GetByConditions retrieves a user based on specified conditions.
	GetByConditions(ctx context.Context, conds map[string]any) (user *entity.User, err error)

	// GetAll retrieves all users based on a filter for pagination.
	GetAll(ctx context.Context, filter base.RequestGetAll) (users []entity.User, total int, err error)

	// Update modifies user information in the repository.
	Update(ctx context.Context, user entity.User) (err error)

	// Delete removes a user from the repository by their ID.
	Delete(ctx context.Context, id int) (err error)

	// UpdatePassword updates a user's password in the repository.
	UpdatePassword(ctx context.Context, id int, passwordHashed string) (err error)
}

type UserRepositoryImpl struct {
	db *gorm.DB
}

var (
	userRepositoryImpl     *UserRepositoryImpl
	userRepositoryImplOnce sync.Once
)

func NewUserRepository() UserRepository {
	userRepositoryImplOnce.Do(func() {
		userRepositoryImpl = &UserRepositoryImpl{
			db: connector.LoadDatabase(),
		}
	})
	return userRepositoryImpl
}

func (repo *UserRepositoryImpl) Create(ctx context.Context, user entity.User, roleID int) (id int, err error) {
	err = repo.db.Transaction(func(tx *gorm.DB) error {
		if res := tx.Create(&user); res.Error != nil {
			tx.Rollback()
			return res.Error
		}
		id = user.ID

		if res := tx.Create(&entity.UserHasRoles{
			UserID: id,
			RoleID: roleID,
		}); res.Error != nil {
			tx.Rollback()
			return res.Error
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (repo *UserRepositoryImpl) GetByID(ctx context.Context, id int) (user *entity.User, err error) {
	user = &entity.User{}
	result := repo.db.Where("id = ?", id).Preload("Roles.Permissions").First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (repo *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (user *entity.User, err error) {
	user = &entity.User{}
	result := repo.db.Where("email = ?", email).Preload("Roles.Permissions").First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (repo *UserRepositoryImpl) GetByConditions(ctx context.Context, conds map[string]any) (user *entity.User, err error) {
	// this func is easy-contain-vunarable by default
	user = &entity.User{}
	query := repo.db
	for con, val := range conds {
		query = query.Where(con+" ?", val)
	}
	result := query.Preload("Roles.Permissions").First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (repo *UserRepositoryImpl) GetAll(ctx context.Context, filter base.RequestGetAll) (users []entity.User, total int, err error) {
	var count int64
	args := []interface{}{"%" + filter.Keyword + "%"}
	cond := "name LIKE ?"
	result := repo.db.Where(cond, args...).Find(&users)
	count = result.RowsAffected
	if result.Error != nil {
		return nil, 0, result.Error
	}
	users = []entity.User{}
	skip := int64(filter.Limit * (filter.Page - 1))
	limit := int64(filter.Limit)
	result = repo.db.Where(cond, args...).Limit(int(limit)).Offset(int(skip)).Find(&users)
	if result.Error != nil {
		return nil, 0, result.Error
	}
	total = int(count)
	return users, total, nil
}

func (repo *UserRepositoryImpl) Update(ctx context.Context, user entity.User) (err error) {
	err = repo.db.Transaction(func(tx *gorm.DB) error {
		var oldData entity.User
		result := tx.Where("id = ?", user.ID).First(&oldData)
		if result.Error != nil {
			return result.Error
		}

		oldData.Name = user.Name
		oldData.ActivatedAt = user.ActivatedAt
		oldData.VerificationCode = user.VerificationCode
		oldData.UpdatedAt = user.UpdatedAt
		result = tx.Save(&oldData)
		if result.Error != nil {
			return result.Error
		}
		return nil
	})

	return err
}

func (repo *UserRepositoryImpl) Delete(ctx context.Context, id int) (err error) {
	deleted := entity.User{}
	result := repo.db.Where("id = ?", id).Delete(&deleted)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (repo *UserRepositoryImpl) UpdatePassword(ctx context.Context, id int, passwordHashed string) (err error) {
	err = repo.db.Transaction(func(tx *gorm.DB) error {
		var user entity.User
		result := tx.Where("id = ?", id).First(&user)
		if result.Error != nil {
			return result.Error
		}
		user.Password = passwordHashed
		user.SetUpdateTime()
		result = tx.Save(&user)
		if result.Error != nil {
			return result.Error
		}
		return nil
	})

	return err
}
