package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Ticket struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	UserID     string             `bson:"user_id"`
	ChatID     string             `bson:"chat_id"`
	Messages   []Message          `bson:"messages"`
	Status     TicketStatus       `bson:"status"`
	AssignedTo string             `bson:"assigned_to,omitempty"`
	CreatedAt  time.Time          `bson:"created_at"`
	ResolvedAt *time.Time         `bson:"resolved_at,omitempty"`
}

type Workspace struct {
	Id           primitive.ObjectID                   `bson:"_id,omitempty"`
	Name         string                               `bson:"name"`
	Team         map[primitive.ObjectID]WorkspaceRole `bson:"team"`
	PendingTeam  map[string]WorkspaceRole             `bson:"pending"`
	Integrations Integrations                         `bson:"integrations"`
	WorkspaceId  string                               `bson:"workspace_id"`
	CreatedAt    primitive.DateTime                   `bson:"created_at"`
}

type User struct {
	Id             primitive.ObjectID `bson:"_id,omitempty"`
	Email          string             `bson:"email"`
	PasswordHash   string             `bson:"password"`
	IsConfirmed    bool               `bson:"is_confirmed"`
	AuthSource     string             `bson:"auth_source"`
	FullName       string             `bson:"name"`
	PendingInvites []string           `bson:"pending_invites"`
	Tokens         Tokens             `bson:"tokens"`
	CreatedAt      primitive.DateTime `bson:"created_at"`
}

type TelegramIntegration struct {
	BotToken string `bson:"bot_token"`
	IsActive bool   `bson:"is_active"`
}

type MetaIntegration struct {
	AuthToken string `bson:"auth_token"`
	PageID    string `bson:"page_id"`
	IsActive  bool   `bson:"is_active"`
}

type WhatsAppIntegration struct {
	InstanceId string `bson:"instance_id"`
	IsActive   bool   `bson:"is_active"`
}

type Integrations struct {
	Id        primitive.ObjectID     `bson:"_id"`
	Telegram  *[]TelegramIntegration `bson:"telegram"`
	Meta      *[]MetaIntegration     `bson:"meta"`
	WhatsApp  *[]WhatsAppIntegration `bson:"whatsapp"`
	CreatedAt time.Time              `bson:"created_at"`
}

type Tokens struct {
	ConfirmToken string `bson:"confirm_token"`
	OAuth2Token  string `bson:"oauth2_token"`
	ResetToken   string `bson:"reset_token"`
	RefreshToken string `bson:"refresh_token"`
}

type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	MessageID string             `bson:"message_id,omitempty"`
	Content   string             `bson:"content"`
	SenderID  string             `bson:"sender_id"`
	Source    MessageSource      `bson:"source"`
	Type      MessageType        `bson:"type"`
	CreatedAt time.Time          `bson:"created_at"`
}

type MessageType string

const (
	TypeText     MessageType = "text"
	TypeImage    MessageType = "image"
	TypeVideo    MessageType = "video"
	TypeAudio    MessageType = "audio"
	TypeDocument MessageType = "document"
)

type WorkspaceRole string

const (
	RoleAdmin  WorkspaceRole = "admin"
	RoleMember WorkspaceRole = "member"
	RoleOwner  WorkspaceRole = "owner"
)

type TicketStatus string

const (
	StatusOpen   TicketStatus = "open"
	StatusClosed TicketStatus = "closed"
)

type MessageSource string

const (
	SourceTelegram  MessageSource = "telegram"
	SourceWhatsApp  MessageSource = "whatsapp"
	SourceInstagram MessageSource = "instagram"
	SourceMeta      MessageSource = "meta"
)
