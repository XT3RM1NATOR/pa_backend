package _interface

import (
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type MessengerService interface {
	ValidateUserInWorkspace(userId primitive.ObjectID, workspaceId string) error
}

type WebsocketService interface {
	UpgradeConnection(w http.ResponseWriter, r *http.Request, workspaceId string) (*websocket.Conn, error)
	RemoveConnection(workspaceId string, conn *websocket.Conn)
	SendToAll(workspaceId string, message []byte)
}
