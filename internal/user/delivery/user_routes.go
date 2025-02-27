package authDelivery

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/user/delivery/controller"
	"github.com/Point-AI/backend/internal/user/infrastructure/client"
	"github.com/Point-AI/backend/internal/user/infrastructure/repository"
	"github.com/Point-AI/backend/internal/user/service"
	"github.com/Point-AI/backend/middleware"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
)

func RegisterAuthRoutes(e *echo.Echo, cfg *config.Config, db *mongo.Database, str *minio.Client, repoMu *sync.RWMutex, storageMu *sync.RWMutex) {
	ur := repository.NewUserRepositoryImpl(db, cfg, repoMu)
	ec := client.NewEmailClientImpl(cfg.Email.SMTPUsername, cfg.Email.SMTPPassword, cfg.Email.SMTPHost, cfg.Email.SMTPPort)
	//sc := client.NewStorageClientImpl(str, storageMu)
	es := service.NewEmailServiceImpl(ec)
	fsi := service.NewFileServiceImpl("../../static")
	us := service.NewUserServiceImpl(ur, fsi, es, cfg)
	uc := controller.NewUserController(us, cfg)

	userGroup := e.Group("/user")
	userGroup.POST("/signup", uc.RegisterUser)
	userGroup.POST("/verify/:token", uc.ConfirmUser)
	userGroup.POST("/signin", uc.Login)
	userGroup.POST("/logout", uc.Logout, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	userGroup.POST("/recover", uc.ForgotPassword)
	userGroup.POST("/reset", uc.ResetPassword)
	userGroup.PUT("/renew", uc.RenewAccessToken)
	userGroup.PUT("/update/:status", uc.UpdateMemberStatus, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	userGroup.GET("/profile", uc.GetProfile, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	userGroup.PUT("/profile", uc.UpdateProfile, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))

	oAuth2Group := e.Group("/oauth2")
	oAuth2Group.GET("/google/callback", uc.GoogleCallback)
	oAuth2Group.GET("/google/tokens", uc.GoogleTokens)
	oAuth2Group.GET("/facebook/callback", uc.FacebookCallback)
}
