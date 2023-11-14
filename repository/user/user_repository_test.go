package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Lukman-e/try-gost/database/connector"
	"github.com/Lukman-e/try-gost/domain/base"
	"github.com/Lukman-e/try-gost/domain/entity"
	"github.com/Lukman-e/try-gost/internal/env"
)

var (
	timeNow time.Time
	ctx     context.Context
)

func init() {
	filePath := "./../../.env"
	env.ReadConfig(filePath)
	timeNow = time.Now()
	ctx = context.Background()
}

func TestNewUserRepository(t *testing.T) {
	userRepository := NewUserRepository()
	if userRepository == nil {
		t.Error("should not nil")
	}
}

func TestUserRepositoryImplCreate(t *testing.T) {
	userRepositoryImpl := UserRepositoryImpl{
		db: connector.LoadDatabase(),
	}

	type args struct {
		ctx  context.Context
		user entity.User
	}
	tests := []struct {
		name      string
		repo      UserRepositoryImpl
		wantErr   bool
		wantPanic bool
		args      args
	}{
		{
			name:      "success create new user",
			repo:      userRepositoryImpl,
			wantErr:   false,
			wantPanic: false,
			args: args{
				ctx: context.Background(),
				user: entity.User{
					Name:     "validname",
					Email:    "valid1@email.com",
					Password: "example-password",
					TimeFields: base.TimeFields{
						CreatedAt: &timeNow,
						UpdatedAt: &timeNow,
					},
				},
			},
		},
		{
			name:      "success create new user with void data",
			repo:      userRepositoryImpl,
			wantErr:   false,
			wantPanic: false,
		},
		{
			name: "failed create new user with void data and nil repository",
			repo: UserRepositoryImpl{
				db: nil,
			},
			wantErr:   true,
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantPanic {
				gotID, err := tt.repo.Create(tt.args.ctx, tt.args.user, 1)
				if (err != nil) != tt.wantErr {
					t.Errorf("UserRepositoryImpl.Create() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				gotID2, err2 := tt.repo.Create(tt.args.ctx, tt.args.user, 1)
				if err2 == nil || gotID2 != 0 {
					t.Error("should be error, couse email is already used")
				}
				tt.repo.Delete(tt.args.ctx, gotID)

				return
			}
			// want panic
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("create() do not panic")
				}
			}()
			userID, err := tt.repo.Create(tt.args.ctx, tt.args.user, 1)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserRepositoryImpl.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			userRepositoryImpl.Delete(ctx, userID)
		})
	}
}

func TestUserRepositoryImplGetByID(t *testing.T) {
	// create user
	user := entity.User{
		Name:     "validname",
		Email:    "valid2@email.com",
		Password: "example-password",
		TimeFields: base.TimeFields{
			CreatedAt: &timeNow,
			UpdatedAt: &timeNow,
		},
	}
	userRepositoryImpl := UserRepositoryImpl{
		db: connector.LoadDatabase(),
	}
	id, createErr := userRepositoryImpl.Create(ctx, user, 1)
	if createErr != nil {
		t.Errorf("error while creating user")
	}
	defer func() {
		userRepositoryImpl.Delete(ctx, id)
	}()

	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name     string
		repo     UserRepositoryImpl
		wantErr  bool
		wantUser bool
		args     args
	}{
		{
			name:     "Success get user by id",
			repo:     userRepositoryImpl,
			wantErr:  false,
			wantUser: true,
			args: args{
				ctx: ctx,
				id:  id,
			},
		},
		{
			name:     "Failed get user by negative id",
			repo:     userRepositoryImpl,
			wantErr:  true,
			wantUser: false,
			args: args{
				ctx: ctx,
				id:  -10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUser, err := tt.repo.GetByID(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserRepositoryImpl.GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantUser {
				if gotUser == nil {
					t.Error("error user shouldn't nil")
				}
			}
		})
	}
}

func TestUserRepositoryImplGetByEmail(t *testing.T) {
	// create user
	user := entity.User{
		Name:     "validname",
		Email:    "valid3@email.com",
		Password: "example-password",
		TimeFields: base.TimeFields{
			CreatedAt: &timeNow,
			UpdatedAt: &timeNow,
		},
	}
	userRepositoryImpl := UserRepositoryImpl{
		db: connector.LoadDatabase(),
	}
	id, createErr := userRepositoryImpl.Create(ctx, user, 1)
	if createErr != nil {
		t.Errorf("error while creating user")
	}
	defer func() {
		userRepositoryImpl.Delete(ctx, id)
	}()

	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name     string
		repo     UserRepositoryImpl
		wantUser bool
		wantErr  bool
		args     args
	}{
		{
			name:     "Success get user by valid email",
			repo:     userRepositoryImpl,
			wantErr:  false,
			wantUser: true,
			args: args{
				ctx:   ctx,
				email: user.Email,
			},
		},
		{
			name:     "Failed get user by invalid-email",
			repo:     userRepositoryImpl,
			wantErr:  true,
			wantUser: false,
			args: args{
				ctx:   ctx,
				email: "invalid-email",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUser, err := tt.repo.GetByEmail(tt.args.ctx, tt.args.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserRepositoryImpl.GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantUser {
				if gotUser == nil {
					t.Error("error user shouldn't nil")
				}
			}
		})
	}
}

func TestUserRepositoryImplGetAll(t *testing.T) {
	// create user
	allUsersID := make([]int, 0)
	userRepositoryImpl := UserRepositoryImpl{
		db: connector.LoadDatabase(),
	}
	for _, id := range []string{"4", "5", "6", "7", "8"} {
		user := entity.User{
			Name:     "validname",
			Email:    "valid" + id + "@email.com", // email is unique
			Password: "example-password",
			TimeFields: base.TimeFields{
				CreatedAt: &timeNow,
				UpdatedAt: &timeNow,
			},
		}
		newUserID, createErr := userRepositoryImpl.Create(ctx, user, 1)
		if createErr != nil {
			t.Errorf("error while creating user :" + id)
		}
		allUsersID = append(allUsersID, newUserID)
	}
	defer func() {
		for _, userID := range allUsersID {
			userRepositoryImpl.Delete(ctx, userID)
		}
	}()

	type args struct {
		ctx    context.Context
		filter base.RequestGetAll
	}
	tests := []struct {
		name    string
		repo    UserRepositoryImpl
		wantErr bool
		args    args
	}{
		{
			name:    "success get 5 or more users",
			repo:    userRepositoryImpl,
			wantErr: false,
			args: args{
				ctx: ctx,
				filter: base.RequestGetAll{
					Page:    1,
					Limit:   1000,
					Keyword: "",
				},
			},
		},
		{
			name:    "success get less than 5",
			repo:    userRepositoryImpl,
			wantErr: false,
			args: args{
				ctx: ctx,
				filter: base.RequestGetAll{
					Page:    1,
					Limit:   1,
					Keyword: "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUsers, gotTotal, err := tt.repo.GetAll(tt.args.ctx, tt.args.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserRepositoryImpl.GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.args.filter.Limit > 5 && len(gotUsers) < 5 {
				t.Error("users should be 5 or more")
			}
			if tt.args.filter.Limit > 5 && gotTotal < 5 {
				t.Error("total users should be 5 or more")
			}
			if tt.args.filter.Limit < 5 && len(gotUsers) > 5 {
				t.Error("users should be less than 5")
			}
		})
	}
}

func TestUserRepositoryImplUpdate(t *testing.T) {
	// create user
	user := entity.User{
		Name:     "validname",
		Email:    "valid9@email.com",
		Password: "example-password",
		TimeFields: base.TimeFields{
			CreatedAt: &timeNow,
			UpdatedAt: &timeNow,
		},
	}
	userRepositoryImpl := UserRepositoryImpl{
		db: connector.LoadDatabase(),
	}
	id, createErr := userRepositoryImpl.Create(ctx, user, 1)
	if createErr != nil {
		t.Errorf("error while creating user")
	}
	// add id to user
	user.ID = id
	defer func() {
		userRepositoryImpl.Delete(ctx, id)
	}()

	type args struct {
		ctx  context.Context
		user entity.User
	}
	tests := []struct {
		name        string
		repo        UserRepositoryImpl
		wantErr     bool
		newUserName string
		args        args
	}{
		{
			name:        "success update user's name",
			repo:        userRepositoryImpl,
			wantErr:     false,
			newUserName: "test-update-001",
			args: args{
				ctx:  ctx,
				user: user,
			},
		},
	}
	for _, tt := range tests {
		tt.args.user.Name = tt.newUserName
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.repo.Update(tt.args.ctx, tt.args.user); (err != nil) != tt.wantErr {
				t.Errorf("UserRepositoryImpl.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			getUser, getErr := tt.repo.GetByID(tt.args.ctx, id)
			if getErr != nil {
				t.Error("error while getting user")
			}
			if getUser.Name != tt.newUserName {
				t.Error("update name failed")
			}
		})
	}
}

func TestUserRepositoryImplDelete(t *testing.T) {
	userRepository := NewUserRepository()
	if userRepository == nil {
		t.Error("shouldn't nil")
	}

	ctx := context.Background()
	err := userRepository.Delete(ctx, -2)
	if err != nil {
		t.Error("delete shouldn't error")
	}
	if ctx.Err() != nil {
		t.Error("delete shouldn't error")
	}
}

func TestUserRepositoryImplUpdatePassword(t *testing.T) {
	// create user
	user := entity.User{
		Name:     "validname",
		Email:    "valid10@email.com",
		Password: "example-password",
		TimeFields: base.TimeFields{
			CreatedAt: &timeNow,
			UpdatedAt: &timeNow,
		},
	}
	userRepositoryImpl := UserRepositoryImpl{
		db: connector.LoadDatabase(),
	}
	id, createErr := userRepositoryImpl.Create(ctx, user, 1)
	if createErr != nil {
		t.Errorf("error while creating user")
	}
	// add id to user
	user.ID = id
	defer func() {
		userRepositoryImpl.Delete(ctx, id)
	}()

	type args struct {
		ctx            context.Context
		id             int
		passwordHashed string
	}
	tests := []struct {
		name    string
		repo    UserRepositoryImpl
		args    args
		wantErr bool
	}{
		{
			name:    "success update user's password",
			repo:    userRepositoryImpl,
			wantErr: false,
			args: args{
				ctx:            ctx,
				id:             id,
				passwordHashed: "new-password-hashed",
			},
		},
		{
			name:    "failed getting user with negative id",
			repo:    userRepositoryImpl,
			wantErr: true,
			args: args{
				ctx: ctx,
				id:  -100,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.repo.UpdatePassword(tt.args.ctx, tt.args.id, tt.args.passwordHashed); (err != nil) != tt.wantErr {
				t.Errorf("UserRepositoryImpl.UpdatePassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				getUser, getErr := tt.repo.GetByID(tt.args.ctx, id)
				if getErr != nil {
					t.Error("error while getting user")
				}
				if getUser.Password != tt.args.passwordHashed {
					t.Error("failed to update user's password")
				}
			}
		})
	}
}

func TestUserRepositoryImplGetByConditions(t *testing.T) {
	user := entity.User{
		Name:     "validname",
		Email:    "valid10@email.com",
		Password: "example-password",
		TimeFields: base.TimeFields{
			CreatedAt: &timeNow,
			UpdatedAt: &timeNow,
		},
	}
	userRepositoryImpl := UserRepositoryImpl{
		db: connector.LoadDatabase(),
	}
	id, createErr := userRepositoryImpl.Create(ctx, user, 1)
	if createErr != nil {
		t.Errorf("error while creating user")
	}
	// add id to user
	user.ID = id
	defer func() {
		userRepositoryImpl.Delete(ctx, id)
	}()

	type args struct {
		ctx   context.Context
		conds map[string]any
	}
	tests := []struct {
		name    string
		repo    UserRepositoryImpl
		args    args
		wantErr bool
	}{
		{
			name: "success get data",
			repo: userRepositoryImpl,
			args: args{
				ctx: ctx,
				conds: map[string]any{
					"name =": user.Name,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUser, err := tt.repo.GetByConditions(tt.args.ctx, tt.args.conds)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserRepositoryImpl.GetByConditions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUser.ID != user.ID || gotUser.Email != user.Email || gotUser.Password != user.Password {
				t.Error("should got same ID/ Email/ Password")
			}
		})
	}
}
