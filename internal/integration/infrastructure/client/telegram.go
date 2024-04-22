package client

import (
	"github.com/Point-AI/backend/config"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io"
	"net/http"
)

type TelegramClient struct {
	config *config.Config
}

func NewTelegramClientImpl(cfg *config.Config) *TelegramClient {
	return &TelegramClient{
		config: cfg,
	}
}

func (tc *TelegramClient) RegisterNewBot(botToken string) error {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return err
	}

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(tc.config.Website.BaseURL + "/bots/webhook/" + botToken))
	if err != nil {
		return err
	}

	return nil
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

//func (tc *TelegramClient) HandleFileMessage(botToken, fileId string) ([]byte, string, error) {
//	bot, err := tgbotapi.NewBotAPI(botToken)
//	if err != nil {
//		return nil, "", err
//	}
//
//	fileData, err := tc.loadMessageFile(bot, fileId)
//	if err != nil {
//		return nil, "", err
//	}
//
//	return fileData, fileType, nil
//}
