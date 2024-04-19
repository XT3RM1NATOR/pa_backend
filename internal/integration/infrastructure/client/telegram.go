package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Point-AI/backend/config"
	_interface "github.com/Point-AI/backend/internal/integration/service/interface"
	"github.com/go-resty/resty/v2"
	"math"
	"time"
)

type apiClient struct {
	conf       *config.Config
	httpClient *resty.Client
}

func NewApiClient(conf *config.Config, httpClient *resty.Client) (_interface.TelegramAPI, error) {
	timeout, err := time.ParseDuration(conf.Telegram.Timeout)
	if err != nil {
		return nil, fmt.Errorf("error during parse duration for telegram timeout : %w", err)
	}

	httpClient.
		SetBaseURL(conf.Telegram.BaseURL).
		SetTimeout(timeout).
		SetHeader("Content-Type", "application/json")

	return &apiClient{
		conf:       conf,
		httpClient: httpClient,
	}, nil
}

func (c *apiClient) SetWebhook(ctx context.Context, botToken, url string) error {
	body := map[string]any{
		"url":                  c.conf.Webhook.BaseURL + url,
		"drop_pending_updates": true,
	}

	httpResp, err := c.httpClient.R().SetBody(body).Post("/bot" + botToken + "/setWebhook")
	if err != nil {
		return err
	}

	if httpResp.StatusCode() != 200 {
		return errors.New("error response: " + httpResp.String())
	}

	var response TelegramResponse
	err = json.Unmarshal(httpResp.Body(), &response)
	if err != nil {
		return err
	}

	if !response.Ok {
		return errors.New("Failed to set webhook. Reason: " + response.Description)
	}

	return nil
}
func (c *apiClient) DeleteWebhook(ctx context.Context, botToken string) error {
	body := map[string]any{
		"drop_pending_updates": true,
	}

	httpResp, err := c.httpClient.R().SetBody(body).Post("/bot" + botToken + "/deleteWebhook")
	if err != nil {
		return err
	}

	if httpResp.StatusCode() != 200 {
		return errors.New("error response: " + httpResp.String())
	}

	var response TelegramResponse
	err = json.Unmarshal(httpResp.Body(), &response)
	if err != nil {
		return err
	}

	if !response.Ok {
		return errors.New("Failed to delete webhook. Reason: " + response.Description)
	}

	return nil
}
func (c *apiClient) SendMessage(ctx context.Context, chatID int, botToken, text string) error {
	body := map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}

	httpResp, err := c.httpClient.R().SetBody(body).Post("/bot" + botToken + "/sendMessage")
	if err != nil {
		return err
	}

	if httpResp.StatusCode() != 200 {
		return errors.New("error response: " + httpResp.String())
	}

	var response TelegramResponse
	err = json.Unmarshal(httpResp.Body(), &response)
	if err != nil {
		return err
	}

	if !response.Ok {
		return errors.New("Failed to send message. Reason: " + response.Description)
	}

	return nil
}
func (c *apiClient) SendTyping(ctx context.Context, chatID int, botToken string) error {
	body := map[string]any{
		"chat_id": chatID,
		"action":  "typing",
	}

	httpResp, err := c.httpClient.R().SetBody(body).Post("/bot" + botToken + "/sendChatAction")
	if err != nil {
		return err
	}

	if httpResp.StatusCode() != 200 {
		return errors.New("error response: " + httpResp.String())
	}

	var response TelegramResponse
	err = json.Unmarshal(httpResp.Body(), &response)
	if err != nil {
		return err
	}

	if !response.Ok {
		return errors.New("Failed to send typing. Reason: " + response.Description)
	}

	return nil
}

func (c *apiClient) SendLanguageInlineKeyboard(ctx context.Context, chatID int, botToken string, languages []*entity.GetLanguageResponse) error {
	var keyboard [][]map[string]any

	rowsCount := int(math.Ceil(float64(len(languages)) / 2.0))

	for i := 0; i < rowsCount; i++ {
		keyboard = append(keyboard, []map[string]any{})
	}

	for i, language := range languages {
		keyboard[i/2] = append(keyboard[i/2], map[string]any{
			"text":          language.Name + " " + language.Flag,
			"callback_data": language.Code,
		})
	}

	body := map[string]any{
		"chat_id": chatID,
		"text":    "Choose the language",
		"reply_markup": map[string]any{
			"inline_keyboard": keyboard,
		},
	}

	httpResp, err := c.httpClient.R().SetBody(body).Post("/bot" + botToken + "/sendMessage")
	if err != nil {
		return err
	}

	var response TelegramResponse
	err = json.Unmarshal(httpResp.Body(), &response)
	if err != nil {
		return err
	}

	if !response.Ok {
		return errors.New("Failed to send language options. Reason: " + response.Description)
	}

	return nil
}

func (c *apiClient) DeleteMessage(ctx context.Context, botToken string, chatID int, messageID int) error {
	body := map[string]any{
		"chat_id":    chatID,
		"message_id": messageID,
	}

	httpResp, err := c.httpClient.R().SetBody(body).Post("/bot" + botToken + "/deleteMessage")
	if err != nil {
		return err
	}

	var response TelegramResponse
	err = json.Unmarshal(httpResp.Body(), &response)
	if err != nil {
		return err
	}

	if !response.Ok {
		return errors.New("Failed to delete message. Reason: " + response.Description)
	}

	return nil
}
