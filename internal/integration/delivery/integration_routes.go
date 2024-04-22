package integrationsDelivery

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/integration/delivery/controller"
	"github.com/Point-AI/backend/internal/integration/infrastructure/repository"
	"github.com/Point-AI/backend/internal/integration/service"
	"github.com/Point-AI/backend/middleware"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterIntegrationRoutes(e *echo.Echo, cfg *config.Config, db *mongo.Database) {
	integrationGroup := e.Group("/integrations")

	ir := repository.NewIntegrationRepositoryImpl(db, cfg)
	is := service.NewIntegrationServiceImpl(cfg, ir)
	ic := controller.NewIntegrationsController(is, cfg)

	telegramGroup := integrationGroup.Group("/telegram")
	telegramGroup.POST("/bots", ic.RegisterBotIntegration, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	telegramGroup.POST("/bots/webhook/:token", ic.HandleBotMessage, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
}
