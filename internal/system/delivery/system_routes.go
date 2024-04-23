package systemDelivery

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/system/delivery/controller"
	"github.com/Point-AI/backend/internal/system/infrastructure/client"
	"github.com/Point-AI/backend/internal/system/infrastructure/repository"
	"github.com/Point-AI/backend/internal/system/service"
	"github.com/Point-AI/backend/middleware"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterSystemRoutes(e *echo.Echo, cfg *config.Config, db *mongo.Database, str *minio.Client) {
	systemGroup := e.Group("/system")

	ec := client.NewEmailClientImpl(cfg.Email.SMTPUsername, cfg.Email.SMTPPassword, cfg.Email.SMTPHost, cfg.Email.SMTPPort)
	src := client.NewStorageClientImpl(str)
	sr := repository.NewSystemRepositoryImpl(cfg, db)
	es := service.NewEmailServiceImpl(ec)
	ss := service.NewSystemServiceImpl(cfg, src, sr, es)
	sc := controller.NewSystemController(cfg, ss)

	workspaceGroup := systemGroup.Group("/workspace")
	workspaceGroup.POST("/", sc.CreateWorkspace, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.POST("/member", sc.AddWorkspaceMembers, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.GET("/:id", sc.GetWorkspaceById, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.GET("/members/:id", sc.GetUserProfiles, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.GET("/", sc.GetAllWorkspaces, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.PUT("/update", sc.UpdateWorkspaceMember, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.PUT("/status/:id/:status", sc.UpdateWorkspacePendingStatus, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.PUT("/:id", sc.UpdateWorkspace, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.DELETE("/leave/:id", sc.LeaveWorkspace, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.DELETE("/member/:id/:email", sc.DeleteWorkspaceMember, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.DELETE("/workspace/:id", sc.DeleteWorkspaceById, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
}
