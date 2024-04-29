package _interface

import (
	"github.com/Point-AI/backend/internal/messenger/delivery/model"
	"github.com/Point-AI/backend/internal/messenger/infrastructure/client"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type MessengerService interface {
	RegisterBotIntegration(userId primitive.ObjectID, botToken, workspaceId string) error
	ValidateUserInWorkspace(userId primitive.ObjectID, workspaceId string) error
	HandleTelegramPlatformMessage(userId primitive.ObjectID, workspaceId string, message model.MessageRequest) error
	HandleTelegramBotMessage(token string, message *tgbotapi.Update) error
	ReassignTicketToMember(userId primitive.ObjectID, ticketId, workspaceId, userEmail string) error
	UpdateTicketStatus(userId primitive.ObjectID, ticketId, workspaceId, status string) error
	HandleTelegramClientAuth(userId primitive.ObjectID, workspaceId, action, value string) (client.AuthStatus, error)
}

type WebsocketService interface {
	UpgradeConnection(w http.ResponseWriter, r *http.Request, workspaceId string) (*websocket.Conn, error)
	RemoveConnection(workspaceId string, conn *websocket.Conn)
	SendToAll(workspaceId string, message []byte)
}
