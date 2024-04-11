package server

import (
	"github.com/Point-AI/backend/config"
	_ "github.com/Point-AI/backend/docs"
	authDelivery "github.com/Point-AI/backend/internal/auth/delivery"
	"github.com/Point-AI/backend/internal/auth/delivery/controller"
	"github.com/Point-AI/backend/internal/auth/infrastructure/repository"
	"github.com/Point-AI/backend/internal/auth/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"os"
)

// RunHTTPServer
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
// @externalDocs.description  OpenAPI 2.0
// @BasePath /
func RunHTTPServer(cfg *config.Config, db *mongo.Database) {
	e := echo.New()

	logger := logrus.New()
	logger.Out = os.Stdout

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339_nano} [${status}] ${method} ${uri} (${latency_human})\n",
		Output: logger.Out,
	}))
	corsConfig := middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}

	e.Use(middleware.CORSWithConfig(corsConfig))

	ur := repository.NewUserRepository(db, "user")

	es := service.NewEmailService(cfg.Email.SMTPUsername, cfg.Email.SMTPPassword, cfg.Email.SMTPHost, cfg.Email.SMTPPort)
	us := service.NewUserService(ur, es, cfg)

	uc := controller.NewUserController(us, cfg)

	authDelivery.RegisterAuthRoutes(e, cfg, uc)
	//integrationsDelivery.RegisterIntegrationsRoutes(e, cfg, db)
	//messangerDelivery.RegisterMessangerAdminRoutes(e, cfg, db)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Start server
	if err := e.Start(cfg.Server.Port); err != nil {
		panic(err)
	}
}
