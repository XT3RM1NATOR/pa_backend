package apiDelivery

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/api/delivery/controller"
	"github.com/Point-AI/backend/internal/api/infrastructure/repository"
	"github.com/Point-AI/backend/internal/api/service"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterAPIRoutes(e *echo.Echo, cfg *config.Config, db *mongo.Database) {
	ar := repository.NewAPIRepositoryImpl(db, cfg)
	as := service.NewAPIServiceImpl(cfg, ar)
	ac := controller.NewAPIController(as, cfg)

	apiGroup := e.Group("/api/v1")
	apiGroup.POST("/article/:id", ac.HandlePost)
	apiGroup.POST("/articles", ac.GetAllArticles)
}
