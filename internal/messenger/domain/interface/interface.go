package _interface

import (
	"github.com/Point-AI/backend/internal/messenger/domain/entity"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type MessengerService interface {
	ReassignTicketToTeam(userId primitive.ObjectID, tgClientId int, ticketId, workspaceId, teamName string) error
	ReassignTicketToUser(userId primitive.ObjectID, tgClientId int, ticketId, workspaceId, email string) error
	ValidateUserInWorkspace(userId primitive.ObjectID, workspace *entity.Workspace) error
	UpdateTicketStatus(userId primitive.ObjectID, ticketId, workspaceId, status string) error
}

type WebsocketService interface {
	UpgradeConnection(w http.ResponseWriter, r *http.Request, workspaceId string) (*websocket.Conn, error)
	RemoveConnection(workspaceId string, conn *websocket.Conn)
	SendToAll(workspaceId string, message []byte)
}
