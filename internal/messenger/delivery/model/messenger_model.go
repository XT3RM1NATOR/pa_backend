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
	TicketId    string `json:"ticket_id"`
	ChatId      string `json:"chat_id"`
	WorkspaceId string `json:"workspace_id"`
	MessageId   string `json:"message_id"`
	Message     string `json:"message"`
	Type        string `json:"type"`
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

type TelegramChat struct {
	EntityType  string          `json:"entity_type"`
	Id          int64           `json:"id"`
	Title       string          `json:"title"`
	UnreadCount int             `json:"unread_count"`
	LastMessage TelegramMessage `json:"last_message"`
}

type TelegramMessage struct {
	Date     string `json:"date"`
	Id       int    `json:"id"`
	SenderId int64  `json:"sender_id"`
	Text     string `json:"text"`
}

type TelegramChatsRequest struct {
	Chats []TelegramChat `json:"chats"`
}

// Responses

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type MessageResponse struct {
	TicketId    string    `json:"ticket_id"`
	ChatId      string    `json:"chat_id"`
	WorkspaceId string    `json:"workspace_id"`
	MessageId   string    `json:"message_id"`
	Message     string    `json:"message"`
	Content     []byte    `json:"content"`
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	IsOwner     bool      `json:"is_owner"`
	Action      string    `json:"action"`
	CreatedAt   time.Time `json:"created_at"`
}

type DeleteMessageResponse struct {
	Type        string `json:"type"`
	WorkspaceId string `json:"workspace_id"`
	ChatId      string `json:"chat_id"`
	MessageId   string `json:"message_id"`
}

type TelegramStatusResponse struct {
	Status string `json:"status"`
}
