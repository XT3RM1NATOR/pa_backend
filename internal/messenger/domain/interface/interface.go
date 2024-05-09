package _interface

import (
	"github.com/Point-AI/backend/internal/messenger/domain/entity"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type MessengerService interface {
	ReassignTicketToTeam(userId primitive.ObjectID, chatId string, ticketId, workspaceId, teamName string) error
	ReassignTicketToUser(userId primitive.ObjectID, chatId string, ticketId, workspaceId, email string) error
	ValidateUserInWorkspace(userId primitive.ObjectID, workspace *entity.Workspace) error
	UpdateTicketStatus(userId primitive.ObjectID, ticketId, workspaceId, status string) error
	ValidateUserInWorkspaceById(userId primitive.ObjectID, workspaceId string) error
	UpdateChatInfo(userId primitive.ObjectID, chatId string, tags []string, workspaceId string) error
	HandleMessage(userId primitive.ObjectID, workspaceId, ticketId, chatId, messageType, message string) error
	DeleteMessage(userId primitive.ObjectID, messageType, workspaceId, ticketId, messageId, chatId string) error
}

type WebsocketService interface {
	UpgradeConnection(w http.ResponseWriter, r *http.Request, workspaceId string) (*websocket.Conn, error)
	RemoveConnection(workspaceId string, conn *websocket.Conn)
	SendToAll(workspaceId string, message []byte)
}
