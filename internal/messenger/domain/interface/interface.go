package _interface

import (
	"github.com/Point-AI/backend/internal/messenger/delivery/model"
	"github.com/Point-AI/backend/internal/messenger/domain/entity"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"time"
)

type MessengerService interface {
	ReassignTicketToTeam(userId primitive.ObjectID, chatId string, ticketId, workspaceId, teamName string) error
	ReassignTicketToUser(userId primitive.ObjectID, chatId string, ticketId, workspaceId, email string) error
	ValidateUserInWorkspace(userId primitive.ObjectID, workspace *entity.Workspace) error
	UpdateTicketStatus(userId primitive.ObjectID, ticketId, workspaceId, status string) error
	ValidateUserInWorkspaceById(userId primitive.ObjectID, workspaceId string) error
	UpdateChatInfo(userId primitive.ObjectID, chatId string, tags []string, workspaceId, language string, address, company, clientEmail, clientPhone string) error
	HandleMessage(userId primitive.ObjectID, workspaceId, ticketId, chatId, messageType, message string) error
	DeleteMessage(userId primitive.ObjectID, messageType, workspaceId, ticketId, messageId, chatId string) error
	GetAllChats(userId primitive.ObjectID, workspaceId string) ([]model.ChatResponse, error)
	ImportTelegramChats(workspaceId string, chats []model.TelegramChat) error
	GetChatsByFolder(userId primitive.ObjectID, workspaceId, folderName string) ([]model.ChatResponse, error)
	GetChat(userId primitive.ObjectID, workspaceId, chatId string) (model.ChatResponse, error)
	GetMessages(userId primitive.ObjectID, workspaceId, chatId string, lastMessageDate time.Time) ([]model.MessageResponse, error)
	GetAllTags(userId primitive.ObjectID, workspaceId string) ([]string, error)
	GetAllPrimaryChats(userId primitive.ObjectID, workspaceId string) ([]model.ChatResponse, error)
	GetAllUnassignedChats(userId primitive.ObjectID, workspaceId string) ([]model.ChatResponse, error)
}

type WebsocketService interface {
	UpgradeConnection(w http.ResponseWriter, r *http.Request, workspaceId string, userId primitive.ObjectID) (*websocket.Conn, error)
	RemoveConnection(workspaceId string, userId primitive.ObjectID)
	AddConnection(workspaceId string, conn *websocket.Conn, userId primitive.ObjectID)
	SendToOne(message []byte, workspaceId string, userId primitive.ObjectID)
	SendToAll(workspaceId string, message []byte)
	GetConnections(workspaceId string) map[primitive.ObjectID]*websocket.Conn
	SendToAllButOne(workspaceId string, message []byte, userId primitive.ObjectID)
}

type FileService interface {
	SaveFile(filename string, content []byte) error
	LoadFile(filename string) ([]byte, error)
	UpdateFileName(oldName, newName string) error
	UpdateFile(newFileBytes []byte, fileName string) error
}
