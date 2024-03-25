package controller

import (
	"fmt"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/auth/delivery/model"
	"github.com/Point-AI/backend/internal/auth/service"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type UserController struct {
	userService *service.UserService
	config      *config.Config
}

func NewUserController(userService *service.UserService, cfg *config.Config) *UserController {
	return &UserController{
		userService: userService,
		config:      cfg,
	}
}

func (uc *UserController) RegisterUser(c echo.Context) error {
	var userInput model.UserInput
	if err := c.Bind(&userInput); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	fmt.Println(userInput)

	if err := uc.userService.RegisterUser(userInput.Email, userInput.Password, userInput.FullName); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "user registered successfully"})
}

func (uc *UserController) ConfirmUser(c echo.Context) error {
	token := c.Param("token")
	if token == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "confirmation token not provided"})
	}

	if err := uc.userService.ConfirmUser(token); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "user confirmed successfully"})
}

func (uc *UserController) Login(c echo.Context) error {
	var userInput model.UserInput
	if err := c.Bind(&userInput); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	accessToken, refreshToken, err := uc.userService.Login(userInput.Email, userInput.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	fmt.Println(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})

	return c.JSON(http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (uc *UserController) ForgotPassword(c echo.Context) error {
	var forgotPasswordInput model.ForgotPasswordInput
	if err := c.Bind(&forgotPasswordInput); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := uc.userService.ForgotPassword(forgotPasswordInput.Email); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "password reset email sent successfully"})
}

func (uc *UserController) ResetPassword(c echo.Context) error {
	var passwordResetInput model.PasswordResetInput
	if err := c.Bind(&passwordResetInput); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := uc.userService.ResetPassword(passwordResetInput.Token, passwordResetInput.NewPassword); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "password reset successfully"})
}

func (uc *UserController) Logout(c echo.Context) error {
	userId := c.Request().Context().Value("userID").(primitive.ObjectID)
	err := uc.userService.Logout(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "successfully logged out"})
}

func (uc *UserController) RenewAccessToken(c echo.Context) error {
	var renewAccessTokenInput model.RenewAccessTokenInput
	if err := c.Bind(&renewAccessTokenInput); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	accessToken, err := uc.userService.RenewAccessToken(renewAccessTokenInput.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"access_token": accessToken})
}

func (uc *UserController) GoogleLogin(c echo.Context) error {
	authURL, err := uc.userService.GoogleLogin()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"url": authURL})
}

// GoogleCallback TODO: change the passing of refresh and access tokens to some other solution like sessions
func (uc *UserController) GoogleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	accessToken, refreshToken, err := uc.userService.GoogleAuthCallback(code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.Redirect(http.StatusFound, "https://...com/?tokens="+accessToken+"#"+refreshToken)
}
