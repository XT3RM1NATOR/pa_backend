package model

import (
	"time"
)

// Requests

type RegisterBotRequest struct {
	WorkspaceId string `json:"workspace_id"`
	BotToken    string `json:"bot_token"`
}

type MessageRequest struct {
	TicketId string `json:"ticket_id"`
	ChatId   string `json:"chat_id"`
	Message  string `json:"message"`
	Type     string `json:"type"`
}

type TelegramAuthRequest struct {
	WorkspaceId   string `json:"workspace_id"`
	PhoneNumber   string `json:"phone_number"`
	PhoneCodeHash string `json:"phone_code_hash"`
	Code          string `json:"code"`
}

type ReassignTicketToTeamRequest struct {
	WorkspaceId string `json:"workspace_id"`
	TeamName    string `json:"team_name"`
	TicketId    string `json:"ticket_id"`
	ChatId      string `json:"tg_client_id"`
}

type ReassignTicketToUserRequest struct {
	WorkspaceId string `json:"workspace_id"`
	Email       string `json:"email"`
	TicketId    string `json:"ticket_id"`
	ChatId      string `json:"tg_client_id"`
}

type ChangeTicketStatusRequest struct {
	TicketId    string `json:"ticket_id"`
	WorkspaceId string `json:"workspace_id"`
	Status      string `json:"status"`
}

type UpdateChatInfoRequest struct {
	ChatId      string   `json:"tg_client_id"`
	WorkspaceId string   `json:"workspace_id"`
	Tags        []string `json:"tags"`
}

// Responses

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type MessageResponse struct {
	TicketId  string    `json:"ticket_id"`
	ChatId    string    `json:"chat_id"`
	Message   string    `json:"message"`
	Content   []byte    `json:"content"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type TelegramStatusResponse struct {
	Status string `json:"status"`
}
