package server

import (
	"github.com/Point-AI/backend/config"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

func RunHTTPServer(cfg *config.Config, db *mongo.Database) {
	e := echo.New()

	routes.RegisterUserRoutes(e, cfg, db)
	routes.RegisterPostRoutes(e, cfg, db)
	routes.RegisterAdminRoutes(e, cfg, db)

	// Start server
	if err := e.Start(cfg.Server.Port); err != nil {
		panic(err)
	}
}
