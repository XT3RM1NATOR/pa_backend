package controller

import (
	"github.com/Point-AI/backend/internal/auth/delivery/model"
	"github.com/Point-AI/backend/internal/auth/service"
	"github.com/labstack/echo/v4"
	"net/http"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
	return &UserController{
		userService: userService,
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

	token, err := uc.userService.Login(userInput.Email, userInput.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
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
