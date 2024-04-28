package infrastructureInterface

import (
	"github.com/Point-AI/backend/internal/messenger/domain/entity"
	"github.com/Point-AI/backend/internal/messenger/infrastructure/client"
	"github.com/celestix/gotgproto"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TelegramBotClientManager interface {
	RegisterNewBot(botToken string) error
	DeleteWebhook(botToken string) error
	SendTextMessage(botToken string, chatID int64, messageText string) error
	HandleFileMessage(botToken, fileId string) ([]byte, error)
	//SendMessage(chatID int, botToken, text string) error
	//SendTyping(chatID int, botToken string) error
	//DeleteMessage(botToken string, chatID int, messageID int) error
}

type TelegramClientManager interface {
	CreateClient(phone, workspaceId string) error
	GetClient(workspaceId string) (*gotgproto.Client, bool)
	GetAuthConversator(workspaceId string) (*client.TelegramAuthConversator, bool)
	SetClient(workspaceId string, client *gotgproto.Client)
	SetAuthConversator(workspaceId string, authConversator *client.TelegramAuthConversator)
}

type WhatsAppClientManager interface {
}

type MessengerRepository interface {
	FindWorkspaceByWorkspaceId(workspaceId string) (*entity.Workspace, error)
	CheckBotExists(botToken string) (bool, error)
	UpdateWorkspace(workspace *entity.Workspace) error
	FindWorkspaceByTelegramBotToken(botToken string) (*entity.Workspace, error)
	FindUserByEmail(email string) (primitive.ObjectID, error)
}
