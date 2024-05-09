package messengerDelivery

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/messenger/delivery/controller"
	"github.com/Point-AI/backend/internal/messenger/infrastructure/repository"
	"github.com/Point-AI/backend/internal/messenger/service"
	"github.com/Point-AI/backend/middleware"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterMessengerRoutes(e *echo.Echo, cfg *config.Config, db *mongo.Database) {
	ir := repository.NewMessengerRepositoryImpl(cfg, db)
	wss := service.NewWebSocketServiceImpl(ir)
	is := service.NewMessengerServiceImpl(cfg, ir, wss)
	ic := controller.NewMessengerController(cfg, is, wss)

	//integrationGroup := e.Group("/integrations")

	//telegramGroup := integrationGroup.Group("/telegram")
	//telegramGroup.POST("/bots", ic.RegisterBotIntegration, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	//telegramGroup.POST("/bots/webhook/:token", ic.HandleBotMessage)
	//telegramGroup.POST("/setInfo/:id", ic.HandleTelegramClientAuth, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))

	messengerGroup := e.Group("/messenger")
	messengerGroup.GET("/ws/:id", ic.WSHandler, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	messengerGroup.POST("/ticket/reassign/team", ic.ReassignTicketToTeam, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	messengerGroup.POST("/ticket/reassign/member", ic.ReassignTicketToMember, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	messengerGroup.PUT("/ticket", ic.ChangeTicketStatus, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	messengerGroup.PUT("/chat", ic.UpdateChatInfo, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))

}
