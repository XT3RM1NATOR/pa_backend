package infrastructureInterface

import (
	"github.com/Point-AI/backend/internal/messenger/domain/entity"
	"github.com/Point-AI/backend/internal/messenger/infrastructure/client"
	"github.com/celestix/gotgproto/ext"

	"github.com/celestix/gotgproto"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TelegramBotClientManager interface {
	RegisterNewBot(botToken string) error
	DeleteWebhook(botToken string) error
	SendTextMessage(botToken string, chatID int64, messageText string) error
	HandleFileMessage(botToken, fileId string) ([]byte, error)
	//SendTyping(chatID int, botToken string) error
	//DeleteMessage(botToken string, chatID int, messageID int) error
}

type TelegramClientManager interface {
	CreateClient(phone, workspaceId string,
		messageHandler func(ctx *ext.Context, update *ext.Update) error,
	) error
	GetClient(workspaceId string) (*gotgproto.Client, bool)
	GetAuthConversator(workspaceId string) (*client.TelegramAuthConversator, bool)
	SetClient(workspaceId string, client *gotgproto.Client)
	SetAuthConversator(workspaceId string, authConversator *client.TelegramAuthConversator)
	CreateClientBySession(session, phone, workspaceId string,
		messageHandler func(ctx *ext.Context, update *ext.Update) error,
	) error
}

type WhatsAppClientManager interface {
}

type MessengerRepository interface {
	FindWorkspaceByWorkspaceId(workspaceId string) (*entity.Workspace, error)
	CheckBotExists(botToken string) (bool, error)
	UpdateWorkspace(workspace *entity.Workspace) error
	FindWorkspaceByTelegramBotToken(botToken string) (*entity.Workspace, error)
	FindUserByEmail(email string) (primitive.ObjectID, error)
	GetAllWorkspaceRepositories() ([]*entity.Workspace, error)
	FindWorkspaceByPhoneNumber(phoneNumber string) (*entity.Workspace, error)
	FindWorkspaceByTicketId(ticketId string) (*entity.Workspace, error)
	GetUserById(id primitive.ObjectID) (*entity.User, error)
	FindChatByWorkspaceIdAndTgClientId(workspaceId primitive.ObjectID, tgClientId int) (*entity.Chat, error)
	FindChatByTicketID(ticketId string) (*entity.Chat, error)
	DeleteChat(chatId primitive.ObjectID) error
	UpdateChat(chat *entity.Chat) error
}
