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

	src := client.NewStorageClientImpl(str)
	sr := repository.NewSystemRepositoryImpl(db, cfg)
	ss := service.NewSystemServiceImpl(cfg, src, sr)
	sc := controller.NewSystemController(ss, cfg)

	projectGroup := systemGroup.Group("/project")
	projectGroup.POST("/", sc.CreateProject, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	projectGroup.POST("/member", sc.AddProjectMembers, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	projectGroup.GET("/:id", sc.GetProjectByID, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	projectGroup.GET("", sc.GetAllProjects, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	projectGroup.PUT("/update", sc.UpdateProjectMember, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	projectGroup.PUT("/:id", sc.UpdateProject, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	projectGroup.PUT("/leave/:id", sc.LeaveProject, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	projectGroup.DELETE("/member/:id/:email", sc.DeleteProjectMember, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	projectGroup.DELETE("/project/:id", sc.DeleteProjectByID, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
}
