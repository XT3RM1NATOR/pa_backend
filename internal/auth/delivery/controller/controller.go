package controller

import "github.com/Point-AI/backend/internal/auth/service"

type AuthController struct {
	Service service.AuthService
}

func NewAuthController(authService service.AuthService, accessTokenSecret, refreshTokenSecret string, accessTokenExpiry, refreshTokenExpiry int) *AuthController {
	return &AuthController{Service: authService}
}
