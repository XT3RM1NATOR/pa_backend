package server

import (
	"github.com/Point-AI/backend/config"
	authDelivery "github.com/Point-AI/backend/internal/auth/delivery"
	integrationsDelivery "github.com/Point-AI/backend/internal/integrations/delivery"
	messangerDelivery "github.com/Point-AI/backend/internal/messanger/delivery"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

func RunHTTPServer(cfg *config.Config, db *mongo.Database) {
	e := echo.New()

	authDelivery.RegisterAuthRoutes(e, cfg, db)
	integrationsDelivery.RegisterIntegrationsRoutes(e, cfg, db)
	messangerDelivery.RegisterMessangerAdminRoutes(e, cfg, db)

	// Start server
	if err := e.Start(cfg.Server.Port); err != nil {
		panic(err)
	}
}
