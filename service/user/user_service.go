package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/Lukman-e/try-gost/database/connector"
	"github.com/Lukman-e/try-gost/domain/entity"
	"github.com/Lukman-e/try-gost/domain/model"
	"github.com/Lukman-e/try-gost/internal/constants"
	"github.com/Lukman-e/try-gost/internal/env"
	"github.com/Lukman-e/try-gost/internal/hash"
	"github.com/Lukman-e/try-gost/internal/helper"
	"github.com/Lukman-e/try-gost/internal/middleware"
	repository "github.com/Lukman-e/try-gost/repository/user"
	emailService "github.com/Lukman-e/try-gost/service/email"
	roleService "github.com/Lukman-e/try-gost/service/role"
)

type UserService interface {

	// Register function register user account, than send verification-code to email
	Register(ctx context.Context, user model.UserRegister) (id int, err error)

	// Verification function activates user account with
	// verification code that has been sended to the user's email
	Verification(ctx context.Context, verifyData model.UserVerificationCode) (err error)

	// DeleteUserByVerification function deletes user data if the user account is not yet verified.
	// This implies that the email owner hasn't actually registered the email, indicating that
	// the user who registered may be making typing errors or may be a hacker attempting to get
	// the verification code.
	DeleteUserByVerification(ctx context.Context, verifyData model.UserVerificationCode) (err error)

	// FailedLoginCounter function counts failed login attempts and stores them in Redis.
	// After the N-th attempt to log in with the same IP address results in continuous failures,
	// the system will impose a 50-minute ban. During this period, login requests (refer to
	// the login function in the user controller) will not be processed.
	FailedLoginCounter(userIP string, increment bool) (counter int, err error)

	// Login func give user token/ jwt for auth header.
	Login(ctx context.Context, user model.UserLogin) (token string, err error)

	// Logout function stores the user's active token in Redis, effectively
	// blacklisting the token. This ensures that the token cannot be reused
	// for authentication (refer to the IsBlacklisted function in internal/middleware).
	Logout(c *fiber.Ctx) (err error)

	// ForgetPassword func send verification code into user's email
	ForgetPassword(ctx context.Context, user model.UserForgetPassword) (err error)

	// ResetPassword func resets password by creating
	// new password by email and verification code
	ResetPassword(ctx context.Context, user model.UserResetPassword) (err error)

	// UpdatePassword func updates user's password
	UpdatePassword(ctx context.Context, user model.UserPasswordUpdate) (err error)

	// UpdateProfile func updates user's profile data
	UpdateProfile(ctx context.Context, user model.UserProfileUpdate) (err error)

	// MyProfile func shows user's profile data
	MyProfile(ctx context.Context, id int) (profile model.UserProfile, err error)
}

type UserServiceImpl struct {
	repository   repository.UserRepository
	roleService  roleService.RoleService
	emailService emailService.EmailService
	jwtHandler   *middleware.JWTHandler
	redis        *redis.Client
}

var (
	userAuthService     *UserServiceImpl
	userAuthServiceOnce sync.Once
)

func NewUserService(roleService roleService.RoleService) UserService {
	userAuthServiceOnce.Do(func() {
		userAuthService = &UserServiceImpl{
			roleService:  roleService,
			repository:   repository.NewUserRepository(),
			emailService: emailService.NewEmailService(),
			jwtHandler:   middleware.NewJWTHandler(),
			redis:        connector.LoadRedisCache(),
		}
	})

	return userAuthService
}

func (svc *UserServiceImpl) Register(ctx context.Context, user model.UserRegister) (id int, err error) {
	// search user by email
	// if exist, return error
	userByEmail, getUserErr := svc.repository.GetByEmail(ctx, user.Email)
	if getUserErr == nil || userByEmail != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "email has been used")
	}

	// search role, if not exist return error
	roleByID, getRoleErr := svc.roleService.GetByID(ctx, user.RoleID)
	if getRoleErr != nil || roleByID == nil {
		return 0, fiber.NewError(fiber.StatusNotFound, "role not found")
	}

	// create verification code
	// for user (must unique)
	var (
		verifCode      string
		passwordHashed string
		hashErr        error
		counter        int = 0
	)
	for {
		verifCode = ""
		verifCode = helper.RandomString(7) + helper.RandomString(7) + helper.RandomString(7) // total = 21
		userGetByCode, getByCodeErr := svc.repository.GetByConditions(ctx, map[string]any{
			"verification_code =": verifCode,
		})
		if getByCodeErr != nil || userGetByCode == nil {
			break
		}
		counter++
		if counter >= 150 {
			return 0, errors.New("failed generating verification code")
		}
	}
	// generate password hashed
	counter = 0
	for {
		passwordHashed, hashErr = hash.Generate(user.Password)
		if hashErr == nil {
			break
		}
		counter++
		if counter >= 150 {
			return 0, errors.New("failed hashing user password")
		}
	}

	userEntity := entity.User{
		Name:             helper.ToTitle(user.Name),
		Email:            user.Email,
		Password:         passwordHashed,
		VerificationCode: &verifCode,
		ActivatedAt:      nil,
	}
	// set created_at and updated_at equal to now
	userEntity.SetCreateTime()
	id, err = svc.repository.Create(ctx, userEntity, user.RoleID)
	if err != nil {
		return 0, err
	}

	toEmail := []string{user.Email}
	subject := "Gost Project: Activation Account"
	message := "Hello, My name is <b>Bot001</b> from Project Gost: Golang Starter By Lukmanern."
	message += "<br/>Your account has already been created but is not yet active. To activate your account,"
	message += " you can click on the Activation Link. If you do not registering for an account or any activity"
	message += " on Project Gost, you can request data deletion too."
	message += "<br />Thank You, Best Regards <b>Bot001</b>."
	message += "<br /><br /><br /> Code : " + verifCode

	sendingErr := svc.emailService.SendMail(toEmail, subject, message)
	if sendingErr != nil {
		message := "account is created, but confimation email failed sending: "
		message += sendingErr.Error()
		return 0, errors.New(message)
	}
	return id, nil
}

func (svc *UserServiceImpl) Verification(ctx context.Context, verifyData model.UserVerificationCode) (err error) {
	// search user by code, if not exist return error
	userEntity, getByCodeErr := svc.repository.GetByConditions(ctx, map[string]any{
		"verification_code =": verifyData.Code,
		"email =":             verifyData.Email,
	})
	if getByCodeErr != nil || userEntity == nil {
		return fiber.NewError(fiber.StatusNotFound, "verification code not found")
	}
	if userEntity.ActivatedAt != nil {
		return fiber.NewError(fiber.StatusBadRequest, "your account already activated")
	}
	// set updated_at, activated_at and
	// nulling verification code
	userEntity.SetActivateAccount()
	userEntity.SetUpdateTime()
	updateErr := svc.repository.Update(ctx, *userEntity)
	if updateErr != nil {
		return updateErr
	}
	return nil
}

func (svc *UserServiceImpl) DeleteUserByVerification(ctx context.Context, verifyData model.UserVerificationCode) (err error) {
	// search user by code, if not exist return error
	userEntity, getByCodeErr := svc.repository.GetByConditions(ctx, map[string]any{
		"verification_code =": verifyData.Code,
		"email =":             verifyData.Email,
	})
	if getByCodeErr != nil || userEntity == nil {
		return fiber.NewError(fiber.StatusNotFound, "verification code not found")
	}
	// check if account active or inactive
	if userEntity.ActivatedAt != nil {
		return fiber.NewError(fiber.StatusBadRequest, "can not delete your account, your account is active")
	}
	deleteErr := svc.repository.Delete(ctx, userEntity.ID)
	if deleteErr != nil {
		return deleteErr
	}
	return nil
}

func (svc *UserServiceImpl) FailedLoginCounter(userIP string, increment bool) (counter int, err error) {
	// set key for banned counter
	key := "failed-login-" + userIP
	getStatus := svc.redis.Get(key)
	counter, _ = strconv.Atoi(getStatus.Val())
	if increment {
		counter++
		setStatus := svc.redis.Set(key, counter, 50*time.Minute)
		if setStatus.Err() != nil {
			return 0, errors.New("storing data to redis")
		}
	}
	return counter, nil
}

func (svc *UserServiceImpl) Login(ctx context.Context, user model.UserLogin) (token string, err error) {
	// search user by email
	// if not exist/found, return error
	userEntity, err := svc.repository.GetByEmail(ctx, user.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fiber.NewError(fiber.StatusNotFound, constants.NotFound)
		}
		return "", err
	}
	if userEntity == nil {
		return "", fiber.NewError(fiber.StatusNotFound, constants.NotFound)
	}

	res, verfiryErr := hash.Verify(userEntity.Password, user.Password)
	if verfiryErr != nil {
		return "", verfiryErr
	}
	if !res {
		return "", fiber.NewError(fiber.StatusBadRequest, "wrong password")
	}
	// if exist but not activated
	// return error
	if userEntity.ActivatedAt == nil {
		message := "Your account is already exist in our system, but it's still inactive, "
		message += "please check Your email inbox to activated-it"
		return "", fiber.NewError(fiber.StatusBadRequest, message)
	}

	userRole := userEntity.Roles[0]
	permIDs := make([]int, 0)
	for _, perm := range userRole.Permissions {
		permIDs = append(permIDs, perm.ID)
	}
	bitGroups := middleware.BuildBitGroups(permIDs...)
	config := env.Configuration()
	expired := time.Now().Add(config.AppAccessTokenTTL)
	token, generetaErr := svc.jwtHandler.GenerateJWT(userEntity.ID, user.Email, userRole.Name, bitGroups, expired)
	if generetaErr != nil {
		message := fmt.Sprintf("system error while generating token (%s)", generetaErr.Error())
		return "", fiber.NewError(fiber.StatusInternalServerError, message)
	}
	if len(token) > 2800 {
		message := "token is too large, more than 2800 characters (too large for http header)"
		return "", errors.New(message)
	}
	return token, nil
}

func (svc *UserServiceImpl) Logout(c *fiber.Ctx) (err error) {
	err = svc.jwtHandler.InvalidateToken(c)
	if err != nil {
		return errors.New("problem invalidating token")
	}

	return nil
}

func (svc *UserServiceImpl) ForgetPassword(ctx context.Context, user model.UserForgetPassword) (err error) {
	userEntity, err := svc.repository.GetByEmail(ctx, user.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, constants.NotFound)
		}
		return err
	}
	if userEntity == nil {
		return fiber.NewError(fiber.StatusNotFound, constants.NotFound)
	}
	if userEntity.ActivatedAt == nil {
		message := "your account has not been activated since register, please check your inbox/ spam mail."
		return fiber.NewError(fiber.StatusBadRequest, message)
	}

	var (
		verifCode string
		counter   int // max retry
	)
	for {
		verifCode = helper.RandomString(7) + helper.RandomString(7) + helper.RandomString(7) // total = 21
		userGetByCode, getByCodeErr := svc.repository.GetByConditions(ctx, map[string]any{
			"verification_code =": verifCode,
		})
		if getByCodeErr != nil || userGetByCode == nil {
			break
		}
		counter++
		if counter >= 150 {
			return errors.New("failed generating verification code")
		}
	}
	userEntity.VerificationCode = &verifCode
	userEntity.SetUpdateTime()

	err = svc.repository.Update(ctx, *userEntity)
	if err != nil {
		return err
	}

	// Todo : refactor
	toEmail := []string{user.Email}
	subject := "Gost Project: Reset Password"
	message := "Hello, My name is <b>Bot001</b> from Project Gost: Golang Starter By Lukmanern."
	message += " Your account has already been created but is not yet active. To activate your account,"
	message += " you can click on the Activation Link. If you do not registering for an account or any activity"
	message += " on Project Gost, you can request data deletion by clicking the Link Request Delete."
	message += "<br /><br />Thank You, Best Regards <b>Bot001</b>."
	message += "<br /><br />Code : " + verifCode

	sendingErr := svc.emailService.SendMail(toEmail, subject, message)
	if sendingErr != nil {
		message := "token forget password is created, but confimation email failed sending: "
		message += sendingErr.Error()
		return errors.New(message)
	}
	return nil
}

func (svc *UserServiceImpl) ResetPassword(ctx context.Context, user model.UserResetPassword) (err error) {
	userByCode, err := svc.repository.GetByConditions(ctx, map[string]any{
		"email =":             user.Email,
		"verification_code =": user.Code,
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, constants.NotFound)
		}
		return err
	}
	if userByCode == nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}
	if userByCode.ActivatedAt == nil {
		message := "Your account is already exist in our system, but it's still "
		message += "inactive, please check Your email inbox to activated-it"
		return fiber.NewError(fiber.StatusBadRequest, message)
	}

	var (
		hashErr      error
		passwdHashed string
		counter      int // max retry
	)
	for {
		passwdHashed, hashErr = hash.Generate(user.NewPassword)
		if hashErr == nil {
			break
		}
		counter++
		if counter >= 150 {
			return errors.New("failed hashing user password")
		}
	}

	userByCode.VerificationCode = nil
	userByCode.SetUpdateTime()
	updateErr := svc.repository.Update(ctx, *userByCode)
	if updateErr != nil {
		return updateErr
	}

	updatePasswdErr := svc.repository.UpdatePassword(ctx, userByCode.ID, passwdHashed)
	if updatePasswdErr != nil {
		return updatePasswdErr
	}
	return nil
}

func (svc *UserServiceImpl) UpdatePassword(ctx context.Context, user model.UserPasswordUpdate) (err error) {
	userByID, err := svc.repository.GetByID(ctx, user.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, constants.NotFound)
		}
		return err
	}
	if userByID == nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}

	res, verfiryErr := hash.Verify(userByID.Password, user.OldPassword)
	if verfiryErr != nil {
		return verfiryErr
	}
	if !res {
		return fiber.NewError(fiber.StatusBadRequest, "wrong password")
	}

	var (
		hashErr      error
		passwdHashed string
		counter      int
	)
	for {
		passwdHashed, hashErr = hash.Generate(user.NewPassword)
		if hashErr == nil {
			break
		}
		counter++
		if counter >= 150 {
			return errors.New("failed hashing user password")
		}
	}

	updatePwErr := svc.repository.UpdatePassword(ctx, userByID.ID, passwdHashed)
	if updatePwErr != nil {
		return updatePwErr
	}
	return nil
}

func (svc *UserServiceImpl) MyProfile(ctx context.Context, id int) (profile model.UserProfile, err error) {
	// search profile by ID
	user, err := svc.repository.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return profile, fiber.NewError(fiber.StatusNotFound, "user not found")
		}

		return profile, err
	}
	if user == nil {
		return profile, fiber.NewError(fiber.StatusInternalServerError, "error while checking user")
	}

	// set response
	profile = model.UserProfile{
		Name:  user.Name,
		Email: user.Email,
	}
	if len(user.Roles) > 0 {
		profile.Role = user.Roles[0]
	}
	return profile, nil
}

func (svc *UserServiceImpl) UpdateProfile(ctx context.Context, user model.UserProfileUpdate) (err error) {
	// search profile by ID
	userByID, getErr := svc.repository.GetByID(ctx, user.ID)
	if getErr != nil {
		if getErr == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, constants.NotFound)
		}
		return err
	}
	if userByID == nil {
		return fiber.NewError(fiber.StatusNotFound, constants.NotFound)
	}

	userEntity := entity.User{
		ID:               user.ID,
		Name:             helper.ToTitle(user.Name),
		VerificationCode: userByID.VerificationCode,
		ActivatedAt:      userByID.ActivatedAt,
	}
	userEntity.SetUpdateTime()

	err = svc.repository.Update(ctx, userEntity)
	if err != nil {
		return err
	}
	return nil
}
