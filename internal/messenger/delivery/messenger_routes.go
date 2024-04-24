package messengerDelivery

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/messenger/delivery/controller"
	"github.com/Point-AI/backend/internal/messenger/infrastructure/client"
	"github.com/Point-AI/backend/internal/messenger/infrastructure/repository"
	"github.com/Point-AI/backend/internal/messenger/service"
	"github.com/Point-AI/backend/middleware"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterMessengerRoutes(e *echo.Echo, cfg *config.Config, db *mongo.Database) {
	tc := client.NewTelegramClientImpl(cfg)
	ir := repository.NewMessengerRepositoryImpl(cfg, db)
	wss := service.NewWebSocketServiceImpl(ir)
	is := service.NewMessengerServiceImpl(cfg, ir, wss, tc)
	ic := controller.NewMessengerController(cfg, is, wss)

	integrationGroup := e.Group("/integrations")

	telegramGroup := integrationGroup.Group("/telegram")
	telegramGroup.POST("/bots", ic.RegisterBotIntegration, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	telegramGroup.POST("/bots/webhook/:token", ic.HandleBotMessage)

	messengerGroup := e.Group("/messenger")
	messengerGroup.GET("/ws/:id", ic.WSHandler, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	messengerGroup.POST("/ticket/reassign/team/:ticket_id/:id/:name", ic.ReassignTicketToTeam, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	messengerGroup.POST("/ticket/reassign/member/:ticket_id/:id/:email", ic.ReassignTicketToMember, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	messengerGroup.PUT("/ticket/:status/:id/:ticket_id", ic.CloseTicket, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
}
