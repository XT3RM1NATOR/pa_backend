package main

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/app/db"
	"github.com/Point-AI/backend/internal/app/server"
	_ "github.com/swaggo/echo-swagger/example/docs"
)

// @title PointAI
// @version 1.0
// @description This is the backend server for PointAI.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name .
// @license.url .
// @host petstore.swagger.io
// @externalDocs.description  OpenAPI
// @BasePath /
func main() {
	cfg := config.Load()
	mongodb := db.ConnectToDB(cfg)

	server.RunHTTPServer(cfg, mongodb)
}
