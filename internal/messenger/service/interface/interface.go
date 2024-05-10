package infrastructureInterface

import (
	"github.com/Point-AI/backend/internal/messenger/domain/entity"
	"github.com/celestix/gotgproto/ext"
	"go.mongodb.org/mongo-driver/mongo"

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
	SetClient(workspaceId string, client *gotgproto.Client)
	CreateClientBySession(session, phone, workspaceId string,
		messageHandler func(ctx *ext.Context, update *ext.Update) error,
	) error
}

type WhatsAppClientManager interface {
}

type MessengerRepository interface {
	FindWorkspaceByWorkspaceId(ctx mongo.SessionContext, workspaceId string) (*entity.Workspace, error)
	CheckBotExists(botToken string) (bool, error)
	UpdateWorkspace(workspace *entity.Workspace) error
	FindWorkspaceByTelegramBotToken(botToken string) (*entity.Workspace, error)
	FindUserByEmail(ctx mongo.SessionContext, email string) (primitive.ObjectID, error)
	GetAllWorkspaceRepositories() ([]*entity.Workspace, error)
	FindWorkspaceByPhoneNumber(phoneNumber string) (*entity.Workspace, error)
	FindWorkspaceByTicketId(ticketId string) (*entity.Workspace, error)
	GetUserById(id primitive.ObjectID) (*entity.User, error)
	FindChatByWorkspaceIdAndChatId(workspaceId primitive.ObjectID, chatId string) (*entity.Chat, error)
	FindChatByTicketId(ctx mongo.SessionContext, ticketId string) (*entity.Chat, error)
	DeleteChat(ctx mongo.SessionContext, chatId primitive.ObjectID) error
	UpdateChat(ctx mongo.SessionContext, chat *entity.Chat) error
	FindWorkspaceById(id primitive.ObjectID) (*entity.Workspace, error)
	FindChatByUserId(ctx mongo.SessionContext, tgClientId int, workspaceId, assigneeId primitive.ObjectID) (*entity.Chat, error)
	InsertNewChat(ctx mongo.SessionContext, chat *entity.Chat) error
	CountActiveTickets(memberId primitive.ObjectID) (int, error)
	StartSession() (mongo.Session, error)
	FindChatByChatId(chatId string) (*entity.Chat, error)
	FindChatsWithLatestTicket(ctx mongo.SessionContext, workspaceId primitive.ObjectID) ([]entity.Chat, error)
}
