package authDelivery

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/auth/delivery/controller"
	"github.com/Point-AI/backend/internal/auth/infrastructure/repository"
	"github.com/Point-AI/backend/internal/auth/service"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterAuthRoutes(e *echo.Echo, cfg *config.Config, db *mongo.Database) {
	userRepository := repository.NewUserRepository(db, "user")
	emailService := service.NewEmailService(cfg.Email.EmailAPIKey)
	userService := service.NewUserService(userRepository, emailService, cfg.Auth.JWTSecretKey)
	userController := controller.NewUserController(userService)

	authGroup := e.Group("/auth")

	authGroup.POST("/signup", userController.RegisterUser)
	authGroup.POST("/verify/:token", userController.ConfirmUser)
	authGroup.POST("/signin", userController.Login)
	authGroup.POST("/recover", userController.ForgotPassword)
	authGroup.POST("/reset", userController.ResetPassword)
}
