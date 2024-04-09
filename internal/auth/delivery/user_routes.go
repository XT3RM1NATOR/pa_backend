package authDelivery

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/auth/delivery/controller"
	"github.com/Point-AI/backend/middleware"
	"github.com/labstack/echo/v4"
)

func RegisterAuthRoutes(e *echo.Echo, cfg *config.Config, uc *controller.UserController) {

	authGroup := e.Group("/auth")

	authGroup.POST("/signup", uc.RegisterUser)
	authGroup.POST("/verify/:token", uc.ConfirmUser)
	authGroup.POST("/signin", uc.Login)
	authGroup.POST("/logout", uc.Logout, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	authGroup.POST("/recover", uc.ForgotPassword)
	authGroup.POST("/reset", uc.ResetPassword)
	authGroup.PUT("/renew", uc.RenewAccessToken)

	authGroup.GET("/oauth2/gooogle/callback", uc.GoogleCallback)
	authGroup.GET("/oauth2/google/tokens", uc.GoogleTokens)
}
