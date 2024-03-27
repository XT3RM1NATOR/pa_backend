package authDelivery

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/auth/delivery/controller"
	"github.com/Point-AI/backend/internal/auth/infrastructure/repository"
	"github.com/Point-AI/backend/internal/auth/service"
	"github.com/Point-AI/backend/middleware"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterAuthRoutes(e *echo.Echo, cfg *config.Config, db *mongo.Database) {
	userRepository := repository.NewUserRepository(db, "user")
	emailService := service.NewEmailService(cfg.Email.SMTPUsername, cfg.Email.SMTPPassword, cfg.Email.SMTPHost, cfg.Email.SMTPPort)
	userService := service.NewUserService(userRepository, emailService, cfg)
	userController := controller.NewUserController(userService, cfg)

	authGroup := e.Group("/auth")

	authGroup.POST("/signup", userController.RegisterUser)
	authGroup.GET("/verify/:token", userController.ConfirmUser)
	authGroup.POST("/signin", userController.Login)
	authGroup.POST("/logout", userController.Logout, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	authGroup.POST("/recover", userController.ForgotPassword)
	authGroup.POST("/reset", userController.ResetPassword)
	authGroup.PUT("/renew", userController.RenewAccessToken)

	authGroup.GET("/oauth2/google/tokens", userController.GoogleTokens)
	authGroup.POST("/oauth2/google", userController.GoogleLogin)
	authGroup.GET("/oauth2/google/callback", userController.GoogleCallback)
}
