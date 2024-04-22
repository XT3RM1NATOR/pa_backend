package infrastructureInterface

import "github.com/Point-AI/backend/internal/messenger/domain/entity"

type TelegramClient interface {
	RegisterNewBot(botToken string) error
	DeleteWebhook(botToken string) error
	//SendMessage(chatID int, botToken, text string) error
	//SendTyping(chatID int, botToken string) error
	//DeleteMessage(botToken string, chatID int, messageID int) error
}

type MessengerRepository interface {
	FindWorkspaceByWorkspaceId(workspaceId string) (*entity.Workspace, error)
}
