package _interface

import "context"

type TelegramAPI interface {
	SetWebhook(ctx context.Context, botToken, url string) error
	DeleteWebhook(ctx context.Context, botToken string) error
	SendMessage(ctx context.Context, chatID int, botToken, text string) error
	SendTyping(ctx context.Context, chatID int, botToken string) error
	SendLanguageInlineKeyboard(ctx context.Context, chatID int, botToken string, languages []*entity.GetLanguageResponse) error
	DeleteMessage(ctx context.Context, botToken string, chatID int, messageID int) error
}
