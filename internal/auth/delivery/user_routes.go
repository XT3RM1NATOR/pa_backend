package authDelivery

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/auth/delivery/controller"
	"github.com/Point-AI/backend/internal/auth/infrastructure/client"
	"github.com/Point-AI/backend/internal/auth/infrastructure/repository"
	"github.com/Point-AI/backend/internal/auth/service"
	"github.com/Point-AI/backend/middleware"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterAuthRoutes(e *echo.Echo, cfg *config.Config, db *mongo.Database) {
	authGroup := e.Group("/auth")

	ur := repository.NewUserRepositoryImpl(db, cfg.MongoDB.UserCollection)
	ec := client.NewEmailClientImpl(cfg.Email.SMTPUsername, cfg.Email.SMTPPassword, cfg.Email.SMTPHost, cfg.Email.SMTPPort)
	es := service.NewEmailServiceImpl(ec)
	us := service.NewUserServiceImpl(ur, es, cfg)
	uc := controller.NewUserController(us, cfg)

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
