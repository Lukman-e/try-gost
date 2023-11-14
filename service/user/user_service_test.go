// Don't run test per file without -p 1
// or simply run test per func or run
// project test using make test command
// check Makefile file
package service

import (
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/Lukman-e/try-gost/database/connector"
	"github.com/Lukman-e/try-gost/domain/model"
	"github.com/Lukman-e/try-gost/internal/constants"
	"github.com/Lukman-e/try-gost/internal/env"
	"github.com/Lukman-e/try-gost/internal/helper"
	"github.com/Lukman-e/try-gost/internal/middleware"
	repository "github.com/Lukman-e/try-gost/repository/user"
	permService "github.com/Lukman-e/try-gost/service/permission"
	roleService "github.com/Lukman-e/try-gost/service/role"
)

func init() {
	// Check env and database
	env.ReadConfig("./../../.env")

	connector.LoadDatabase()
	connector.LoadRedisCache()
}

func TestNewUserService(t *testing.T) {
	permSvc := permService.NewPermissionService()
	roleSvc := roleService.NewRoleService(permSvc)
	svc := NewUserService(roleSvc)
	if svc == nil {
		t.Error(constants.ShouldNotNil)
	}
}

func TestSuccessRegister(t *testing.T) {
	defer func() {
		connector.LoadRedisCache().FlushAll()
	}()
	permSvc := permService.NewPermissionService()
	roleSvc := roleService.NewRoleService(permSvc)
	svc := NewUserService(roleSvc)
	c := helper.NewFiberCtx()
	ctx := c.Context()
	if svc == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}

	userRepo := repository.NewUserRepository()
	if userRepo == nil {
		t.Error(constants.ShouldNotNil)
	}

	modelUserRegis := model.UserRegister{
		Name:     helper.RandomString(12),
		Email:    helper.RandomEmail(),
		Password: helper.RandomString(12),
		RoleID:   1, // admin
	}
	userID, regisErr := svc.Register(ctx, modelUserRegis)
	if regisErr != nil || userID < 1 {
		t.Error("should not error and id should more than zero")
	}

	defer func() {
		userRepo.Delete(ctx, userID)
	}()

	userByID, getErr := userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		t.Error("should not error and id should not nil")
	}
	if userByID.Name != helper.ToTitle(modelUserRegis.Name) ||
		userByID.Email != modelUserRegis.Email ||
		userByID.Roles[0].ID != modelUserRegis.RoleID {
		t.Error("should equal")
	}
	if userByID.VerificationCode == nil {
		t.Error(constants.ShouldNotNil)
	}
	if userByID.ActivatedAt != nil {
		t.Error(constants.ShouldNil)
	}

	// failed login : account is created,
	// but account is inactive
	modelUserLogin := model.UserLogin{
		Email:    modelUserRegis.Email,
		Password: modelUserRegis.Password,
		IP:       helper.RandomIPAddress(),
	}
	token, loginErr := svc.Login(ctx, modelUserLogin)
	if loginErr == nil || token != "" {
		t.Error("should error login and token should nil-string")
	}
	fiberErr, ok := loginErr.(*fiber.Error)
	if ok {
		if fiberErr.Code != fiber.StatusBadRequest {
			t.Error("should error 400BadReq")
		}
	}

	// failed forget password : account is created,
	// but account is inactive
	forgetPassErr := svc.ForgetPassword(ctx, model.UserForgetPassword{Email: modelUserRegis.Email})
	if forgetPassErr == nil {
		t.Error("should error login and token should nil-string")
	}
	fiberErr, ok = forgetPassErr.(*fiber.Error)
	if ok {
		if fiberErr.Code != fiber.StatusBadRequest {
			t.Error("should error 400BadReq")
		}
	}

	// failed forget password : account is created,
	// but account is inactive
	resetPasswdErr := svc.ResetPassword(ctx, model.UserResetPassword{Code: "wrongCode"})
	if resetPasswdErr == nil {
		t.Error("should error login and token should nil-string")
	}

	vCode := userByID.VerificationCode

	verifErr := svc.Verification(ctx, model.UserVerificationCode{
		Code:  *vCode,
		Email: userByID.Email,
	})
	if verifErr != nil {
		t.Error(constants.ShouldNotNil)
	}

	// value reset
	userByID = nil
	getErr = nil
	userByID, getErr = userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		t.Error("should not error and id should not nil")
	}
	if userByID.VerificationCode != nil {
		t.Error(constants.ShouldNotNil)
	}
	if userByID.ActivatedAt == nil {
		t.Error(constants.ShouldNil)
	}

	// reset value
	token = ""
	loginErr = nil
	modelUserLogin = model.UserLogin{
		Email:    modelUserRegis.Email,
		Password: modelUserRegis.Password,
		IP:       helper.RandomIPAddress(),
	}
	token, loginErr = svc.Login(ctx, modelUserLogin)
	if loginErr != nil || token == "" {
		t.Error("should not error login and token should not nil-string")
	}

	jwtHandler := middleware.NewJWTHandler()
	if jwtHandler.IsBlacklisted(token) {
		t.Error("should not in black-list")
	}

	modelUserForgetPasswd := model.UserForgetPassword{
		Email: modelUserLogin.Email,
	}
	forgetPwErr := svc.ForgetPassword(ctx, modelUserForgetPasswd)
	if forgetPwErr != nil {
		t.Error(constants.ShouldNotErr)
	}

	// value reset
	userByID = nil
	getErr = nil
	userByID, getErr = userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		t.Error("should not error and id should not nil")
	}
	if userByID.VerificationCode == nil {
		t.Error(constants.ShouldNotNil)
	}
	if userByID.ActivatedAt == nil {
		t.Error(constants.ShouldNotNil)
	}

	passwd := helper.RandomString(12)
	modelUserResetPasswd := model.UserResetPassword{
		Email:              userByID.Email,
		Code:               *userByID.VerificationCode,
		NewPassword:        passwd,
		NewPasswordConfirm: passwd,
	}
	resetErr := svc.ResetPassword(ctx, modelUserResetPasswd)
	if resetErr != nil {
		t.Error(constants.ShouldNotErr)
	}

	// reset value, login failed
	token = ""
	loginErr = nil
	modelUserLogin = model.UserLogin{
		Email:    modelUserRegis.Email,
		Password: modelUserRegis.Password,
		IP:       helper.RandomIPAddress(),
	}
	token, loginErr = svc.Login(ctx, modelUserLogin)
	if loginErr == nil || token != "" {
		t.Error("should error login and token should nil-string")
	}

	// reset value, login success
	token = ""
	loginErr = nil
	modelUserLogin = model.UserLogin{
		Email:    modelUserRegis.Email,
		Password: modelUserResetPasswd.NewPassword,
		IP:       helper.RandomIPAddress(),
	}
	token, loginErr = svc.Login(ctx, modelUserLogin)
	if loginErr != nil || token == "" {
		t.Error("should not error login and token should not nil-string")
	}

	passwd = helper.RandomString(14)
	modelUserUpdatePasswd := model.UserPasswordUpdate{
		ID:                 userID,
		OldPassword:        modelUserResetPasswd.NewPassword,
		NewPassword:        passwd,
		NewPasswordConfirm: passwd,
	}
	updatePasswdErr := svc.UpdatePassword(ctx, modelUserUpdatePasswd)
	if updatePasswdErr != nil {
		t.Error(constants.ShouldNotErr)
	}

	// reset value, login success
	token = ""
	loginErr = nil
	modelUserLogin = model.UserLogin{
		Email:    modelUserRegis.Email,
		Password: modelUserUpdatePasswd.NewPassword,
		IP:       helper.RandomIPAddress(),
	}
	token, loginErr = svc.Login(ctx, modelUserLogin)
	if loginErr != nil || token == "" {
		t.Error("should not error login and token should not nil-string")
	}

	modelUserUpdate := model.UserProfileUpdate{
		ID:   userID,
		Name: helper.RandomString(10),
	}
	updateProfileErr := svc.UpdateProfile(ctx, modelUserUpdate)
	if updateProfileErr != nil {
		t.Error(constants.ShouldNotErr)
	}

	profile, getErr := svc.MyProfile(ctx, userID)
	if getErr != nil {
		t.Error(constants.ShouldNotErr)
	}
	if profile.Name != helper.ToTitle(modelUserUpdate.Name) {
		t.Error("should equal")
	}

	// success logout
	cForLogout := helper.NewFiberCtx()
	logoutErr := svc.Logout(cForLogout)
	if logoutErr != nil {
		t.Error("should no error")
	}
}

func TestFailedRegister(t *testing.T) {
	defer func() {
		connector.LoadRedisCache().FlushAll()
	}()
	permSvc := permService.NewPermissionService()
	roleSvc := roleService.NewRoleService(permSvc)
	svc := NewUserService(roleSvc)
	c := helper.NewFiberCtx()
	ctx := c.Context()
	if svc == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}

	userRepo := repository.NewUserRepository()
	if userRepo == nil {
		t.Error(constants.ShouldNotNil)
	}

	modelUserRegis := model.UserRegister{
		Name:     helper.RandomString(12),
		Email:    helper.RandomEmail(),
		Password: helper.RandomString(12),
		RoleID:   -10, // failed
	}
	userID, regisErr := svc.Register(ctx, modelUserRegis)
	if regisErr == nil || userID != 0 {
		t.Error("should error and id should zero")
	}

	defer func() {
		userRepo.Delete(ctx, userID)
	}()

	verifErr := svc.Verification(ctx, model.UserVerificationCode{
		Code:  "wrongCode",
		Email: "wrongEmail",
	})
	if verifErr == nil {
		t.Error(constants.ShouldErr)
	}
	fiberErr, ok := verifErr.(*fiber.Error)
	if ok {
		if fiberErr.Code != fiber.StatusNotFound {
			t.Error("should error 404")
		}
	}

	deleteUserErr := svc.DeleteUserByVerification(ctx, model.UserVerificationCode{
		Code:  "wrongCode",
		Email: "wrongEmail",
	})
	if deleteUserErr == nil {
		t.Error(constants.ShouldErr)
	}
	fiberErr, ok = deleteUserErr.(*fiber.Error)
	if ok {
		if fiberErr.Code != fiber.StatusNotFound {
			t.Error("should error 404")
		}
	}

	// failed login
	_, loginErr := svc.Login(ctx, model.UserLogin{
		IP: helper.RandomIPAddress(),
	})
	if loginErr == nil {
		t.Error(constants.ShouldErr)
	}

	forgetErr := svc.ForgetPassword(ctx, model.UserForgetPassword{Email: "wrong_email@gost.project"})
	if forgetErr == nil {
		t.Error(constants.ShouldErr)
	}

	verifyErr := svc.ResetPassword(ctx, model.UserResetPassword{Code: "wrong-code"})
	if verifyErr == nil {
		t.Error(constants.ShouldErr)
	}

	updatePasswdErr := svc.UpdatePassword(ctx, model.UserPasswordUpdate{ID: -1})
	if updatePasswdErr == nil {
		t.Error(constants.ShouldErr)
	}

	_, getErr := svc.MyProfile(ctx, -10)
	if getErr == nil {
		t.Error(constants.ShouldErr)
	}
}

func TestBannedIPAddress(t *testing.T) {
	defer func() {
		connector.LoadRedisCache().FlushAll()
	}()
	permSvc := permService.NewPermissionService()
	roleSvc := roleService.NewRoleService(permSvc)
	svc := NewUserService(roleSvc)
	c := helper.NewFiberCtx()
	ctx := c.Context()
	if svc == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}

	for i := 1; i <= 15; i++ {
		counter, err := svc.FailedLoginCounter(helper.RandomIPAddress(), true)
		if err != nil {
			t.Error(constants.ShouldNotErr)
		}
		if i >= 4 {
			if counter == i {
				t.Error("counter should error")
			}
		}
	}
}
