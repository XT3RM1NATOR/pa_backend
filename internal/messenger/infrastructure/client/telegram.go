package client

import (
	"github.com/Point-AI/backend/config"
	infrastructureInterface "github.com/Point-AI/backend/internal/messenger/service/interface"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type TelegramClient struct {
	config *config.Config
}

func NewTelegramClientImpl(cfg *config.Config) infrastructureInterface.TelegramClient {
	return &TelegramClient{
		config: cfg,
	}
}

func (tc *TelegramClient) SendMessage(authToken string, chatID int64, messageText string) error {
	bot, err := tgbotapi.NewBotAPI(authToken)
	if err != nil {
		return err
	}

	message := tgbotapi.NewMessage(chatID, messageText)

	_, err = bot.Send(message)
	if err != nil {
		return err
	}

	return nil
}

// ReceiveMessages retrieves new messages from the Telegram account with the provided authToken.
func (tc *TelegramClient) ReceiveMessages(authToken string) ([]*tgbotapi.Message, error) {
	bot, err := tgbotapi.NewBotAPI(authToken)
	if err != nil {
		return nil, err
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return nil, err
	}

	var messages []*tgbotapi.Message
	for update := range updates {
		if update.Message != nil {
			messages = append(messages, update.Message)
		}
	}

	return messages, nil
}

// ReceiveVideoMessages retrieves new video messages from the Telegram account with the provided authToken.
func (tc *TelegramClient) ReceiveVideoMessages(authToken string) ([]*tgbotapi.Message, error) {
	bot, err := tgbotapi.NewBotAPI(authToken)
	if err != nil {
		return nil, err
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return nil, err
	}

	var videoMessages []*tgbotapi.Message
	for update := range updates {
		if update.Message != nil && update.Message.Video != nil {
			videoMessages = append(videoMessages, update.Message)
		}
	}

	return videoMessages, nil
}
