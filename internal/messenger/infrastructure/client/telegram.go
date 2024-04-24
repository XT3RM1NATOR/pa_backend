package client

import (
	"github.com/Point-AI/backend/config"
	infrastructureInterface "github.com/Point-AI/backend/internal/messenger/service/interface"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io"
	"net/http"
)

type TelegramClient struct {
	config *config.Config
}

func NewTelegramClientImpl(cfg *config.Config) infrastructureInterface.TelegramClient {
	return &TelegramClient{
		config: cfg,
	}
}

func (tc *TelegramClient) RegisterNewBot(botToken string) error {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return err
	}

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(tc.config.Website.BaseURL + "integrations/telegram/bots/webhook/" + botToken))
	if err != nil {
		return err
	}

	return nil
}

func (tc *TelegramClient) SendTextMessage(botToken string, chatID int64, messageText string) error {
	bot, err := tgbotapi.NewBotAPI(botToken)
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

func (tc *TelegramClient) DeleteWebhook(botToken string) error {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return err
	}

	_, err = bot.RemoveWebhook()
	if err != nil {
		return err
	}

	return nil
}

func (tc *TelegramClient) HandleFileMessage(botToken, fileId string) ([]byte, string, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, "", err
	}

	fileData, err := tc.loadMessageFile(bot, fileId)
	if err != nil {
		return nil, "", err
	}

	return fileData, fileId, nil
}

func (tc *TelegramClient) loadMessageFile(bot *tgbotapi.BotAPI, videoMessageId string) ([]byte, error) {
	fileURL, err := bot.GetFileDirectURL(videoMessageId)
	if err != nil {
		return nil, err
	}

	response, err := http.Get(fileURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	videoMessageData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return videoMessageData, nil
}
