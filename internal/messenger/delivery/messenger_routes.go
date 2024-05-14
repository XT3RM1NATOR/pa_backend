package messengerDelivery

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/messenger/delivery/controller"
	"github.com/Point-AI/backend/internal/messenger/infrastructure/repository"
	"github.com/Point-AI/backend/internal/messenger/service"
	"github.com/Point-AI/backend/middleware"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
)

func RegisterMessengerRoutes(e *echo.Echo, cfg *config.Config, db *mongo.Database, mu *sync.RWMutex) {
	ir := repository.NewMessengerRepositoryImpl(cfg, db, mu)
	fsi := service.NewFileServiceImpl("../../static")
	wss := service.NewWebSocketServiceImpl(ir)
	is := service.NewMessengerServiceImpl(cfg, ir, wss, fsi)
	ic := controller.NewMessengerController(cfg, is, wss)

	messengerGroup := e.Group("/messenger")
	messengerGroup.GET("/poop", ic.SendOk)
	messengerGroup.GET("/ws/:id", ic.WSHandler, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	messengerGroup.POST("/ticket/reassign/team", ic.ReassignTicketToTeam, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	messengerGroup.POST("/ticket/reassign/member", ic.ReassignTicketToMember, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	messengerGroup.PUT("/ticket", ic.ChangeTicketStatus, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	messengerGroup.PUT("/chat", ic.UpdateChatInfo, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	messengerGroup.GET("/chats/:id", ic.GetAllChats, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	messengerGroup.GET("/chats/message/:id/:chat_id", ic.GetMessages, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	messengerGroup.GET("/chats/folder/:id/:name", ic.GetChatsByFolder, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	messengerGroup.GET("/chats/tags/:id", ic.GetAllTags, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	messengerGroup.GET("/chats/chat/:id/:chat_id", ic.GetChat, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	messengerGroup.DELETE("/message", ic.DeleteMessage, middleware.ValidateAccessTokenMiddleware(cfg.Auth.JWTSecretKey))
	//messengerGroup.GET("/messages/:id/")

	telegramGroup := e.Group("/telegram")
	telegramGroup.POST("/import/:id", ic.ImportTelegramChats, middleware.ValidateServerMiddleware(cfg.Auth.IntegrationsServerSecretKey))
	//telegramGroup.POST("/messages/webhook/:id", ic.HandleTelegramMessage, middleware.ValidateServerMiddleware(cfg.Auth.IntegrationsServerSecretKey))
}
