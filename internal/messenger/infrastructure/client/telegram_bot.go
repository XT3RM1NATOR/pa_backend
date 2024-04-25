package client

import (
	"github.com/Point-AI/backend/config"
	infrastructureInterface "github.com/Point-AI/backend/internal/messenger/service/interface"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io"
	"net/http"
)

type TelegramBotClient struct {
	config *config.Config
}

func NewTelegramBotClientImpl(cfg *config.Config) infrastructureInterface.TelegramBotClient {
	return &TelegramBotClient{
		config: cfg,
	}
}

func (tbc *TelegramBotClient) RegisterNewBot(botToken string) error {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return err
	}

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(tbc.config.Website.BaseURL + "integrations/telegram/bots/webhook/" + botToken))
	if err != nil {
		return err
	}

	return nil
}

func (tbc *TelegramBotClient) SendTextMessage(botToken string, chatID int64, messageText string) error {
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

func (tbc *TelegramBotClient) DeleteWebhook(botToken string) error {
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

func (tbc *TelegramBotClient) HandleFileMessage(botToken, fileId string) ([]byte, string, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, "", err
	}

	fileData, err := tbc.loadMessageFile(bot, fileId)
	if err != nil {
		return nil, "", err
	}

	return fileData, fileId, nil
}

func (tbc *TelegramBotClient) loadMessageFile(bot *tgbotapi.BotAPI, videoMessageId string) ([]byte, error) {
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
