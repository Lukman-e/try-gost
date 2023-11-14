package controller

import (
	"net"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/Lukman-e/try-gost/domain/model"
	"github.com/Lukman-e/try-gost/internal/constants"
	"github.com/Lukman-e/try-gost/internal/middleware"
	"github.com/Lukman-e/try-gost/internal/response"
	service "github.com/Lukman-e/try-gost/service/user"
)

type UserController interface {
	// Register function register user account,
	// than send verification-code to email
	Register(c *fiber.Ctx) error

	// AccountActivation function activates user account with
	// verification code that has been sended to the user's email
	AccountActivation(c *fiber.Ctx) error

	// DeleteUserByVerification function deletes user data if the
	// user account is not yet verified. This implies that the email
	// owner hasn't actually registered the email, indicating that
	// the user who registered may be making typing errors or may
	// be a hacker attempting to get the verification code.
	DeleteAccountActivation(c *fiber.Ctx) error

	// ForgetPassword function send
	// verification code into user's email
	ForgetPassword(c *fiber.Ctx) error

	// ResetPassword func resets password by creating
	// new password by email and verification code
	ResetPassword(c *fiber.Ctx) error

	// Login func gives token and access to user
	Login(c *fiber.Ctx) error

	// Logout func stores user's token into Redis
	Logout(c *fiber.Ctx) error

	// UpdatePassword func updates user's password
	UpdatePassword(c *fiber.Ctx) error

	// UpdateProfile func updates user's profile data
	UpdateProfile(c *fiber.Ctx) error

	// MyProfile func shows user's profile data
	MyProfile(c *fiber.Ctx) error
}

type UserControllerImpl struct {
	service service.UserService
}

var (
	userAuthController     *UserControllerImpl
	userAuthControllerOnce sync.Once
)

func NewUserController(service service.UserService) UserController {
	userAuthControllerOnce.Do(func() {
		userAuthController = &UserControllerImpl{
			service: service,
		}
	})

	return userAuthController
}

func (ctr *UserControllerImpl) Register(c *fiber.Ctx) error {
	var user model.UserRegister
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}
	user.Email = strings.ToLower(user.Email)
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}

	ctx := c.Context()
	id, regisErr := ctr.service.Register(ctx, user)
	if regisErr != nil {
		fiberErr, ok := regisErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, constants.ServerErr+regisErr.Error())
	}

	message := "Account success created. please check " + user.Email + " "
	message += "inbox, our system has sended verification code or link."
	data := map[string]any{
		"id": id,
	}
	return response.CreateResponse(c, fiber.StatusCreated, true, message, data)
}

func (ctr *UserControllerImpl) AccountActivation(c *fiber.Ctx) error {
	var user model.UserVerificationCode
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}
	ctx := c.Context()
	err := ctr.service.Verification(ctx, user)
	if err != nil {
		fiberErr, ok := err.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, constants.ServerErr+err.Error())
	}

	message := "Thank you for your confirmation. Your account is active now, you can login."
	return response.CreateResponse(c, fiber.StatusOK, true, message, nil)
}

func (ctr *UserControllerImpl) DeleteAccountActivation(c *fiber.Ctx) error {
	var verifyData model.UserVerificationCode
	if err := c.BodyParser(&verifyData); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(&verifyData); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}
	ctx := c.Context()
	err := ctr.service.DeleteUserByVerification(ctx, verifyData)
	if err != nil {
		fiberErr, ok := err.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, constants.ServerErr+err.Error())
	}

	message := "Your data is already deleted, thank you for your confirmation."
	return response.CreateResponse(c, fiber.StatusOK, true, message, nil)
}

func (ctr *UserControllerImpl) Login(c *fiber.Ctx) error {
	var user model.UserLogin
	// user.IP = c.IP() // Note : uncomment this line in production
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}

	userIP := net.ParseIP(user.IP)
	if userIP == nil {
		return response.BadRequest(c, constants.InvalidBody+"invalid user ip address")
	}
	counter, _ := ctr.service.FailedLoginCounter(userIP.String(), false)
	ipBlockMsg := "Your IP has been blocked by system. Please try again in 1 or 2 Hour"
	if counter >= 5 {
		return response.CreateResponse(c, fiber.StatusBadRequest, false, ipBlockMsg, nil)
	}

	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}

	ctx := c.Context()
	token, loginErr := ctr.service.Login(ctx, user)
	if loginErr != nil {
		counter, _ := ctr.service.FailedLoginCounter(userIP.String(), true)
		if counter >= 5 {
			return response.CreateResponse(c, fiber.StatusBadRequest, false, ipBlockMsg, nil)
		}
		fiberErr, ok := loginErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, constants.ServerErr+loginErr.Error())
	}

	data := map[string]any{
		"token":        token,
		"token-length": len(token),
	}
	return response.CreateResponse(c, fiber.StatusOK, true, "success login", data)
}

func (ctr *UserControllerImpl) Logout(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}
	logoutErr := ctr.service.Logout(c)
	if logoutErr != nil {
		return response.Error(c, constants.ServerErr+logoutErr.Error())
	}
	return response.CreateResponse(c, fiber.StatusOK, true, "success logout", nil)
}

func (ctr *UserControllerImpl) ForgetPassword(c *fiber.Ctx) error {
	var user model.UserForgetPassword
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}

	ctx := c.Context()
	forgetErr := ctr.service.ForgetPassword(ctx, user)
	if forgetErr != nil {
		fiberErr, ok := forgetErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, constants.ServerErr+forgetErr.Error())
	}

	message := "success sending link for reset password to email, check your email inbox"
	return response.CreateResponse(c, fiber.StatusAccepted, true, message, nil)
}

func (ctr *UserControllerImpl) ResetPassword(c *fiber.Ctx) error {
	var user model.UserResetPassword
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}
	if user.NewPassword != user.NewPasswordConfirm {
		return response.BadRequest(c, "password confirmation not match")
	}

	ctx := c.Context()
	resetErr := ctr.service.ResetPassword(ctx, user)
	if resetErr != nil {
		fiberErr, ok := resetErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, constants.ServerErr+resetErr.Error())
	}

	message := "your password already updated, you can login with your new password, thank you"
	return response.CreateResponse(c, fiber.StatusAccepted, true, message, nil)
}

func (ctr *UserControllerImpl) UpdatePassword(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}

	var user model.UserPasswordUpdate
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}
	user.ID = userClaims.ID

	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}
	if user.NewPassword != user.NewPasswordConfirm {
		return response.BadRequest(c, "new password confirmation is wrong")
	}
	if user.NewPassword == user.OldPassword {
		return response.BadRequest(c, "no new password, try another new password")
	}

	ctx := c.Context()
	updateErr := ctr.service.UpdatePassword(ctx, user)
	if updateErr != nil {
		fiberErr, ok := updateErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, constants.ServerErr+updateErr.Error())
	}

	return response.SuccessNoContent(c)
}

func (ctr *UserControllerImpl) UpdateProfile(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}

	var user model.UserProfileUpdate
	if err := c.BodyParser(&user); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}
	user.ID = userClaims.ID
	validate := validator.New()
	if err := validate.Struct(&user); err != nil {
		return response.BadRequest(c, constants.InvalidBody+err.Error())
	}

	ctx := c.Context()
	updateErr := ctr.service.UpdateProfile(ctx, user)
	if updateErr != nil {
		fiberErr, ok := updateErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, constants.ServerErr+updateErr.Error())
	}

	return response.SuccessNoContent(c)
}

func (ctr *UserControllerImpl) MyProfile(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("claims").(*middleware.Claims)
	if !ok || userClaims == nil {
		return response.Unauthorized(c)
	}

	ctx := c.Context()
	userProfile, getErr := ctr.service.MyProfile(ctx, userClaims.ID)
	if getErr != nil {
		fiberErr, ok := getErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, constants.ServerErr+getErr.Error())
	}
	return response.SuccessLoaded(c, userProfile)
}
