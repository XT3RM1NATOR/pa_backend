package client

import (
	"github.com/Point-AI/backend/config"
	infrastructureInterface "github.com/Point-AI/backend/internal/messenger/service/interface"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io"
	"net/http"
)

type TelegramBotClientManager struct {
	config *config.Config
}

func NewTelegramBotClientManagerImpl(cfg *config.Config) infrastructureInterface.TelegramBotClientManager {
	return &TelegramBotClientManager{
		config: cfg,
	}
}

func (tbcm *TelegramBotClientManager) RegisterNewBot(botToken string) error {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return err
	}

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(tbcm.config.Website.BaseURL + "integrations/telegram/bots/webhook/" + botToken))
	if err != nil {
		return err
	}

	return nil
}

func (tbcm *TelegramBotClientManager) SendTextMessage(botToken string, chatID int64, messageText string) error {
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

func (tbcm *TelegramBotClientManager) DeleteWebhook(botToken string) error {
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

func (tbcm *TelegramBotClientManager) HandleFileMessage(botToken, fileId string) ([]byte, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}

	fileData, err := tbcm.loadMessageFile(bot, fileId)
	if err != nil {
		return nil, err
	}

	return fileData, nil
}

func (tbcm *TelegramBotClientManager) loadMessageFile(bot *tgbotapi.BotAPI, videoMessageId string) ([]byte, error) {
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
