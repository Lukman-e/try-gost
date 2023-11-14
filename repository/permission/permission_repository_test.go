package repository

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/Lukman-e/try-gost/database/connector"
	"github.com/Lukman-e/try-gost/domain/base"
	"github.com/Lukman-e/try-gost/domain/entity"
	"github.com/Lukman-e/try-gost/internal/env"
)

var (
	permissionRepoImpl PermissionRepositoryImpl
	timeNow            time.Time
	ctx                context.Context
)

func init() {
	filePath := "./../../.env"
	env.ReadConfig(filePath)
	timeNow = time.Now()
	ctx = context.Background()
	permissionRepoImpl = PermissionRepositoryImpl{
		db: connector.LoadDatabase(),
	}

}

func createOnePermission(t *testing.T, namePrefix string) *entity.Permission {
	permission := entity.Permission{
		Name:        "valid-permission-name-" + namePrefix,
		Description: "valid-permission-description-" + namePrefix,
		TimeFields: base.TimeFields{
			CreatedAt: &timeNow,
			UpdatedAt: &timeNow,
		},
	}
	id, createErr := permissionRepoImpl.Create(ctx, permission)
	if createErr != nil {
		t.Errorf("error while creating permission")
	}
	permission.ID = id
	return &permission
}

func TestNewPermissionRepository(t *testing.T) {
	permRepo := NewPermissionRepository()
	if permRepo == nil {
		t.Error("should not nil")
	}
}

func TestPermissionRepositoryImplCreate(t *testing.T) {
	permission := createOnePermission(t, "create-same-name")
	if permission == nil {
		t.Error("failed creating permission : permission is nil")
	}
	defer func() {
		permissionRepoImpl.Delete(ctx, permission.ID)
	}()

	type args struct {
		ctx        context.Context
		permission entity.Permission
	}
	tests := []struct {
		name      string
		repo      PermissionRepositoryImpl
		args      args
		wantErr   bool
		wantPanic bool
	}{
		{
			name:      "error while creating with the same name",
			wantErr:   true,
			wantPanic: true,
			args: args{
				ctx: ctx,
				permission: entity.Permission{
					Name:        permission.Name,
					Description: "",
					TimeFields: base.TimeFields{
						CreatedAt: &timeNow,
						UpdatedAt: &timeNow,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil && tt.wantPanic {
					t.Errorf("create() do not panic")
				}
			}()
			gotID, err := tt.repo.Create(tt.args.ctx, tt.args.permission)
			if (err != nil) != tt.wantErr {
				t.Errorf("PermissionRepositoryImpl.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotID <= 0 {
				t.Errorf("ID should be positive")
			}
		})
	}
}

func TestPermissionRepositoryImplGetByID(t *testing.T) {
	permission := createOnePermission(t, "TestGetByID")
	if permission == nil {
		t.Error("failed creating permission : permission is nil")
	}
	defer func() {
		permissionRepoImpl.Delete(ctx, permission.ID)
	}()

	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name    string
		repo    PermissionRepositoryImpl
		args    args
		wantErr bool
	}{
		{
			name: "success get permission by valid id",
			repo: permissionRepoImpl,
			args: args{
				ctx: ctx,
				id:  permission.ID,
			},
			wantErr: false,
		},
		{
			name: "failed get permission by invalid id",
			repo: permissionRepoImpl,
			args: args{
				ctx: ctx,
				id:  -10,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPermission, err := tt.repo.GetByID(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("PermissionRepositoryImpl.GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gotPermission == nil {
				t.Error("permission should not nil")
			}
		})
	}
}

func TestPermissionRepositoryImplGetByName(t *testing.T) {
	permission := createOnePermission(t, "TestGetByName")
	if permission == nil {
		t.Error("failed creating permission : permission is nil")
	}
	defer func() {
		permissionRepoImpl.Delete(ctx, permission.ID)
	}()

	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name           string
		repo           PermissionRepositoryImpl
		args           args
		wantPermission *entity.Permission
		wantErr        bool
	}{
		{
			name: "success get permission by valid id",
			repo: permissionRepoImpl,
			args: args{
				ctx:  ctx,
				name: permission.Name,
			},
			wantErr: false,
		},
		{
			name: "failed get permission by invalid id",
			repo: permissionRepoImpl,
			args: args{
				ctx:  ctx,
				name: "unknown name",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPermission, err := tt.repo.GetByName(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("PermissionRepositoryImpl.GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gotPermission == nil {
				t.Error("permission should not nil")
			}
		})
	}
}

func TestPermissionRepositoryImplGetAll(t *testing.T) {
	permissions := make([]entity.Permission, 0)
	for i := 0; i < 10; i++ {
		permission := createOnePermission(t, "TestGetAll-"+strconv.Itoa(i))
		if permission == nil {
			continue
		}
		defer func() {
			permissionRepoImpl.Delete(ctx, permission.ID)
		}()

		permissions = append(permissions, *permission)
	}
	lenPermissions := len(permissions)

	type args struct {
		ctx    context.Context
		filter base.RequestGetAll
	}
	tests := []struct {
		name    string
		repo    PermissionRepositoryImpl
		args    args
		wantErr bool
	}{
		{
			name: "success get all",
			repo: permissionRepoImpl,
			args: args{
				ctx: ctx,
				filter: base.RequestGetAll{
					Limit: 1000,
					Page:  1,
				},
			},
			wantErr: false,
		},
		{
			name: "success get all",
			repo: permissionRepoImpl,
			args: args{
				ctx: ctx,
				filter: base.RequestGetAll{
					Limit: 1,
					Page:  1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPermissions, gotTotal, err := tt.repo.GetAll(tt.args.ctx, tt.args.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("PermissionRepositoryImpl.GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.args.filter.Limit > lenPermissions && len(gotPermissions) < lenPermissions {
				t.Error("permissions should be $lenPermissions or more")
			}
			if tt.args.filter.Limit > lenPermissions && gotTotal < lenPermissions {
				t.Error("total permissions should be $lenPermissions or more")
			}
			if tt.args.filter.Limit < lenPermissions && len(gotPermissions) > lenPermissions {
				t.Error("permissions should be less than $lenPermission")
			}
		})
	}
}

func TestPermissionRepositoryImplUpdate(t *testing.T) {
	permission := createOnePermission(t, "TestUpdateByID")
	if permission == nil {
		t.Error("failed creating permission : permission is nil")
	}
	defer func() {
		permissionRepoImpl.Delete(ctx, permission.ID)
	}()

	type args struct {
		ctx        context.Context
		permission entity.Permission
	}
	tests := []struct {
		name    string
		repo    PermissionRepositoryImpl
		wantErr bool
		args    args
	}{
		{
			name:    "success update name and desc",
			repo:    permissionRepoImpl,
			wantErr: false,
			args: args{
				ctx: ctx,
				permission: entity.Permission{
					ID:          permission.ID,
					Name:        "updated name",
					Description: "updated description",
				},
			},
		},
		{
			name:    "failed update name and desc with invalid id",
			repo:    permissionRepoImpl,
			wantErr: true,
			args: args{
				ctx: ctx,
				permission: entity.Permission{
					ID:          -10,
					Name:        "updated name",
					Description: "updated description",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.repo.Update(tt.args.ctx, tt.args.permission); (err != nil) != tt.wantErr {
				t.Errorf("PermissionRepositoryImpl.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			p, err := tt.repo.GetByID(tt.args.ctx, permission.ID)
			if err != nil {
				t.Error("error while getting permission")
			}
			if p.Name != tt.args.permission.Name || p.Description != tt.args.permission.Description {
				t.Error("name and description failed to update")
			}
		})
	}
}

func TestPermissionRepositoryImplDelete(t *testing.T) {
	permission := createOnePermission(t, "TestDeleteByID")
	if permission == nil {
		t.Error("failed creating permission : permission is nil")
	}
	defer func() {
		permissionRepoImpl.Delete(ctx, permission.ID)
	}()

	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name    string
		repo    PermissionRepositoryImpl
		args    args
		wantErr bool
	}{
		{
			name:    "success update permission",
			repo:    permissionRepoImpl,
			wantErr: false,
			args: args{
				ctx: ctx,
				id:  permission.ID,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.repo.Delete(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("PermissionRepositoryImpl.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			permission, err := tt.repo.GetByID(tt.args.ctx, tt.args.id)
			if !tt.wantErr && err == nil {
				t.Error("should error")
			}
			if !tt.wantErr && permission != nil {
				t.Error("permission should nil")
			}
		})
	}
}
