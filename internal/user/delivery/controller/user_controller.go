package controller

import (
	"fmt"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/user/delivery/model"
	_interface "github.com/Point-AI/backend/internal/user/domain/interface"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type UserController struct {
	userService _interface.UserService
	config      *config.Config
}

func NewUserController(userService _interface.UserService, cfg *config.Config) *UserController {
	return &UserController{
		userService: userService,
		config:      cfg,
	}
}

// RegisterUser registers a new user.
// @Description Registers a new user with provided email and password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.UserRequest true "User registration request"
// @Success 201 {object} model.SuccessResponse "User registered successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /user/signup [post]
func (uc *UserController) RegisterUser(c echo.Context) error {
	var request model.UserRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	if err := uc.userService.RegisterUser(request.Email, request.Password, request.WorkspaceId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, model.SuccessResponse{Message: "user registered successfully"})
}

// ConfirmUser confirms a user's registration.
// @Description Confirms a user's registration using the confirmation token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param token path string true "Confirmation token"
// @Success 200 {object} model.SuccessResponse "User confirmed successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /user/verify/{token} [get]
func (uc *UserController) ConfirmUser(c echo.Context) error {
	token := c.Param("token")
	if token == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "confirmation token not provided"})
	}

	if err := uc.userService.ConfirmUser(token); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "user confirmed successfully"})
}

// Login handles user login.
// @Summary User login
// @Description Logs in a user with the provided email and password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.UserRequest true "User login request"
// @Success 200 {object} model.TokenResponse "User logged in successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /user/signin [post]
func (uc *UserController) Login(c echo.Context) error {
	var request model.UserRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	accessToken, refreshToken, err := uc.userService.Login(request.Email, request.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// GoogleTokens exchanges OAuth2 tokens for Google tokens.
// @Summary Exchange OAuth2 tokens for Google tokens
// @Description Exchanges OAuth2 tokens for Google access and refresh tokens.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.OAuth2TokenRequest true "OAuth2 token request"
// @Success 200 {object} model.TokenResponse "Tokens exchanged successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /user/oauth2/google/tokens [get]
func (uc *UserController) GoogleTokens(c echo.Context) error {
	var request model.OAuth2TokenRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	accessToken, refreshToken, err := uc.userService.GoogleTokens(request.OAuth2Token)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// ForgotPassword initiates the process for resetting a user's password.
// @Summary Forgot password
// @Description Initiates the process for resetting a user's password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.ForgotPasswordRequest true "Forgot password request"
// @Success 200 {object} model.SuccessResponse "Password reset email sent successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /user/recover [post]
func (uc *UserController) ForgotPassword(c echo.Context) error {
	var request model.ForgotPasswordRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	if err := uc.userService.ForgotPassword(request.Email); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "password reset email sent successfully"})
}

// ResetPassword resets a user's password.
// @Summary Reset password
// @Description Resets a user's password using the reset token and new password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.PasswordResetRequest true "Password reset request"
// @Success 200 {object} model.SuccessResponse "Password reset successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /user/reset [post]
func (uc *UserController) ResetPassword(c echo.Context) error {
	var request model.PasswordResetRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	if err := uc.userService.ResetPassword(request.Token, request.NewPassword); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "password reset successfully"})
}

// Logout logs out a user.
// @Summary Logout
// @Description Logs out a user by invalidating the access token.
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} model.SuccessResponse "Successfully logged out"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /user/logout [post]
func (uc *UserController) Logout(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)
	if err := uc.userService.Logout(userId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "successfully logged out"})
}

// RenewAccessToken renews a user's access token using a refresh token.
// @Summary Renew access token
// @Description Renews a user's access token using a refresh token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.RenewAccessTokenRequest true "Access token renewal request"
// @Success 200 {object} model.TokenResponse "Access token renewed successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /user/renew [put]
func (uc *UserController) RenewAccessToken(c echo.Context) error {
	var request model.RenewAccessTokenRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	accessToken, err := uc.userService.RenewAccessToken(request.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.TokenResponse{AccessToken: accessToken})
}

// GoogleCallback handles the callback from Google OAuth2 login.
// @Summary Google OAuth2 callback
// @Description Handles the callback from Google OAuth2 login.
// @Tags Auth
// @Accept json
// @Produce json
// @Param code query string true "Authorization code from Google"
// @Success 302 {object} model.SuccessResponse "Redirect to website URL with OAuth2 token"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /user/oauth2/google/callback [get]
func (uc *UserController) GoogleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	oAuth2Token, err := uc.userService.GoogleAuthCallback(code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("%s/?oauth2token="+oAuth2Token, uc.config.Website.WebURL))
}

func (uc *UserController) FacebookCallback(c echo.Context) error {
	code, workspaceId := c.QueryParam("code"), c.QueryParam("id")
	if err := uc.userService.FacebookAuthCallback(code, workspaceId); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf(uc.config.Website.WebURL+"/integrations"))
}

// GetProfile returns the user profile.
// @Summary returns the user profile.
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} model.UserProfileResponse "user profile data"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /user/profile [get]
func (uc *UserController) GetProfile(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)
	user, logo, err := uc.userService.GetUserProfile(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.UserProfileResponse{
		Email:     user.Email,
		FullName:  user.FullName,
		Logo:      logo,
		CreatedAt: user.CreatedAt,
	})
}

// UpdateProfile updates the user profile.
// @Summary updates the user profile.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.UserUpdateProfileRequest true "Access token renewal request"
// @Success 200 {object} model.SuccessResponse "user profile data"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /user/profile [put]
func (uc *UserController) UpdateProfile(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)
	var request model.UserUpdateProfileRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	if err := uc.userService.UpdateUserProfile(userId, request.Logo, request.FullName); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "user updated successfully"})
}
