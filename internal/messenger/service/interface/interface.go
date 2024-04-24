package infrastructureInterface

import (
	"github.com/Point-AI/backend/internal/messenger/domain/entity"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TelegramClient interface {
	RegisterNewBot(botToken string) error
	DeleteWebhook(botToken string) error
	SendTextMessage(botToken string, chatID int64, messageText string) error
	//SendMessage(chatID int, botToken, text string) error
	//SendTyping(chatID int, botToken string) error
	//DeleteMessage(botToken string, chatID int, messageID int) error
}

type MessengerRepository interface {
	FindWorkspaceByWorkspaceId(workspaceId string) (*entity.Workspace, error)
	AddTelegramIntegration(id primitive.ObjectID, botToken string) error
	CheckBotExists(botToken string) (bool, error)
	UpdateWorkspace(workspace *entity.Workspace) error
	FindWorkspaceByTelegramBotToken(botToken string) (*entity.Workspace, error)
	FindUserByEmail(email string) (primitive.ObjectID, error)
}
