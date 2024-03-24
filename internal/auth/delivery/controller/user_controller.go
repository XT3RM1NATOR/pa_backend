package controller

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/auth/delivery/model"
	"github.com/Point-AI/backend/internal/auth/service"
	"github.com/labstack/echo/v4"
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

	if err := uc.userService.RegisterUser(userInput.Email, userInput.Password); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "User registered successfully"})
}

func (uc *UserController) ConfirmUser(c echo.Context) error {
	token := c.Param("token")
	if token == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "confirmation token not provided"})
	}

	if err := uc.userService.ConfirmUser(token); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User confirmed successfully"})
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

	return c.JSON(http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (uc *UserController) ForgotPassword(c echo.Context) error {
	var newPasswordInput model.NewPasswordInput

	if err := c.Bind(&newPasswordInput); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := uc.userService.ForgotPassword(newPasswordInput.Email); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Password reset email sent successfully"})
}

func (uc *UserController) ResetPassword(c echo.Context) error {
	var passwordResetInput model.PasswordResetInput

	if err := c.Bind(&passwordResetInput); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := uc.userService.ResetPassword(passwordResetInput.Token, passwordResetInput.NewPassword); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Password reset successfully"})
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
