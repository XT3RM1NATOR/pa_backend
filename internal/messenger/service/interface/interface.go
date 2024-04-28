package infrastructureInterface

import (
	"context"
	"github.com/Point-AI/backend/internal/messenger/domain/entity"
	"github.com/gotd/td/tg"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TelegramBotClient interface {
	RegisterNewBot(botToken string) error
	DeleteWebhook(botToken string) error
	SendTextMessage(botToken string, chatID int64, messageText string) error
	HandleFileMessage(botToken, fileId string) ([]byte, error)
	//SendMessage(chatID int, botToken, text string) error
	//SendTyping(chatID int, botToken string) error
	//DeleteMessage(botToken string, chatID int, messageID int) error
}

type TelegramClient interface {
	Authenticate(ctx context.Context, phoneNumber string) (*tg.AuthSentCode, error)
	SignIn(ctx context.Context, phoneNumber, phoneCodeHash, phoneCode string) (*tg.AuthAuthorization, error)
	SignInFA(ctx context.Context, password string) (*tg.AuthAuthorization, error)
}

type WhatsAppClient interface {
}

type MessengerRepository interface {
	FindWorkspaceByWorkspaceId(workspaceId string) (*entity.Workspace, error)
	CheckBotExists(botToken string) (bool, error)
	UpdateWorkspace(workspace *entity.Workspace) error
	FindWorkspaceByTelegramBotToken(botToken string) (*entity.Workspace, error)
	FindUserByEmail(email string) (primitive.ObjectID, error)
}
