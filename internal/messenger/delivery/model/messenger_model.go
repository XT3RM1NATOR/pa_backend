package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Requests

type RegisterBotRequest struct {
	WorkspaceId string `json:"workspace_id"`
	BotToken    string `json:"bot_token"`
}

type MessageRequest struct {
	TicketId  string              `json:"ticket_id"`
	Message   string              `json:"message"`
	Type      string              `json:"type"`
	Source    string              `json:"source"`
	CreatedAt *primitive.DateTime `json:"created_at,omitempty"`
}

type TelegramAuthRequest struct {
	WorkspaceId   string `json:"workspace_id"`
	PhoneNumber   string `json:"phone_number"`
	PhoneCodeHash string `json:"phone_code_hash"`
	Code          string `json:"code"`
}

// Responses

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type MessageResponse struct {
	TicketId  string             `json:"ticket_id"`
	Message   string             `json:"message"`
	Content   []byte             `json:"content"`
	Type      string             `json:"type"`
	Source    string             `json:"source"`
	Username  string             `json:"username"`
	CreatedAt primitive.DateTime `json:"created_at"`
}

type TelegramStatusResponse struct {
	Status string `json:"status"`
}
