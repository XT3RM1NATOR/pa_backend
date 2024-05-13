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
	"sync"
)

func RegisterSystemRoutes(e *echo.Echo, cfg *config.Config, db *mongo.Database, str *minio.Client, repoMu *sync.RWMutex, storageMu *sync.RWMutex) {
	systemGroup := e.Group("/system")

	ec := client.NewEmailClientImpl(cfg.Email.SMTPUsername, cfg.Email.SMTPPassword, cfg.Email.SMTPHost, cfg.Email.SMTPPort)
	src := client.NewStorageClientImpl(str, storageMu)
	sr := repository.NewSystemRepositoryImpl(cfg, db, repoMu)
	es := service.NewEmailServiceImpl(ec)
	ss := service.NewSystemServiceImpl(cfg, src, sr, es)
	sc := controller.NewSystemController(cfg, ss)

	workspaceGroup := systemGroup.Group("/workspace")
	workspaceGroup.POST("", sc.CreateWorkspace, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.POST("/member", sc.AddWorkspaceMembers, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.POST("/teams", sc.AddTeamsMembers, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.POST("/folders", sc.AddFolders, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.POST("/teams/:id/:name", sc.SetFirstTeam, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.POST("/team", sc.CreateTeam, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.GET("/:id", sc.GetWorkspaceById, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.GET("/members/:id", sc.GetUserProfiles, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.GET("", sc.GetAllWorkspaces, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.GET("/team/:id", sc.GetAllTeams, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.PUT("/update", sc.UpdateWorkspaceMember, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.GET("/folders", sc.GetAllFolders, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.PUT("/update/:status/:id", sc.UpdateMemberStatus, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.PUT("/status/:id/:status", sc.UpdateWorkspacePendingStatus, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	//workspaceGroup.PUT("/team", sc.UpdateTeam, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.PUT("/:id", sc.UpdateWorkspace, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.DELETE("/leave/:id", sc.LeaveWorkspace, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.DELETE("/team/:id/:name", sc.DeleteTeam, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.DELETE("/member/:id/:email", sc.DeleteWorkspaceMember, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	workspaceGroup.DELETE("/workspace/:id", sc.DeleteWorkspaceById, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))

	integrationsAuthGroup := systemGroup.Group("/integrations")
	integrationsAuthGroup.GET("/telegram/:id", sc.RegisterTelegramIntegration, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
}
