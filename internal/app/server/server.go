package server

import (
	"github.com/Point-AI/backend/config"
	authDelivery "github.com/Point-AI/backend/internal/auth/delivery"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
)

func RunHTTPServer(cfg *config.Config, db *mongo.Database) {
	e := echo.New()
	logger := logrus.New()

	logger.Out = os.Stdout

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339_nano} [${status}] ${method} ${uri} (${latency_human})\n",
		Output: logger.Out,
	}))
	e.Use(middleware.CORS())

	authDelivery.RegisterAuthRoutes(e, cfg, db)
	//integrationsDelivery.RegisterIntegrationsRoutes(e, cfg, db)
	//messangerDelivery.RegisterMessangerAdminRoutes(e, cfg, db)

	// Start server
	if err := e.Start(cfg.Server.Port); err != nil {
		panic(err)
	}
}
