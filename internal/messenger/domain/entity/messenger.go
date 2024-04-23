package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Workspace struct {
	Id           primitive.ObjectID                           `bson:"_id,omitempty"`
	Name         string                                       `bson:"name"`
	Team         map[primitive.ObjectID]WorkspaceRole         `bson:"team"`
	PendingTeam  map[string]WorkspaceRole                     `bson:"pending"`
	Teams        map[string]map[primitive.ObjectID]UserStatus `bson:"teams"`
	FirstTeam    string                                       `bson:"first_team"`
	Integrations Integrations                                 `bson:"integrations"`
	Tickets      []Ticket                                     `bson:"tickets"`
	WorkspaceId  string                                       `bson:"workspace_id"`
	CreatedAt    primitive.DateTime                           `bson:"created_at"`
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

type Tokens struct {
	ConfirmToken string `bson:"confirm_token"`
	OAuth2Token  string `bson:"oauth2_token"`
	ResetToken   string `bson:"reset_token"`
	RefreshToken string `bson:"refresh_token"`
}

type Ticket struct {
	Id                  primitive.ObjectID    `bson:"_id,omitempty"`
	TicketId            string                `bson:"ticket_id,omitempty"`
	BotToken            string                `bson:"bot_token"`
	SenderId            int                   `bson:"user_id"`
	ChatId              int64                 `bson:"chat_id"`
	IntegrationMessages []IntegrationsMessage `bson:"integration_messages"`
	ResponseMessages    []ResponseMessage     `bson:"response_messages"`
	Status              TicketStatus          `bson:"status"`
	Source              TicketSource          `bson:"source"`
	AssignedTo          primitive.ObjectID    `bson:"assigned_to,omitempty"`
	CreatedAt           primitive.DateTime    `bson:"created_at"`
	ResolvedAt          *primitive.DateTime   `bson:"resolved_at,omitempty"`
}

type ResponseMessage struct {
	Id        primitive.ObjectID  `bson:"_id,omitempty"`
	SenderId  primitive.ObjectID  `bson:"sender_id,omitempty"`
	Message   string              `bson:"message"`
	Type      MessageType         `bson:"type"`
	CreatedAt *primitive.DateTime `bson:"created_at,omitempty"`
}

type IntegrationsMessage struct {
	Id        primitive.ObjectID `bson:"_id,omitempty"`
	MessageId int                `bson:"message_id"`
	Message   string             `bson:"message"`
	Type      MessageType        `bson:"type"`
	CreatedAt primitive.DateTime `bson:"created_at,omitempty"`
}

type Integrations struct {
	Id          primitive.ObjectID        `bson:"_id"`
	TelegramBot *[]TelegramBotIntegration `bson:"telegram_bot"`
	Meta        *[]MetaIntegration        `bson:"meta"`
	WhatsApp    *[]WhatsAppIntegration    `bson:"whatsapp"`
	CreatedAt   time.Time                 `bson:"created_at"`
}

type TelegramBotIntegration struct {
	BotToken string `bson:"bot_token"`
	IsActive bool   `bson:"is_active"`
}

type MetaIntegration struct {
	AuthToken string `bson:"auth_token"`
	PageId    string `bson:"page_id"`
	IsActive  bool   `bson:"is_active"`
}

type WhatsAppIntegration struct {
	InstanceId string `bson:"instance_id"`
	IsActive   bool   `bson:"is_active"`
}

type MessageType string
type WorkspaceRole string
type TicketSource string
type TicketStatus string
type UserStatus string

const (
	TypeText     MessageType = "text"
	TypeImage    MessageType = "image"
	TypeVideo    MessageType = "video"
	TypeAudio    MessageType = "audio"
	TypeDocument MessageType = "document"
)

const (
	RoleAdmin  WorkspaceRole = "admin"
	RoleMember WorkspaceRole = "member"
	RoleOwner  WorkspaceRole = "owner"
)

const (
	StatusOpen    TicketStatus = "open"
	StatusPending TicketStatus = "pending"
	StatusClosed  TicketStatus = "closed"
)

const (
	SourceTelegram    TicketSource = "telegram"
	SourceTelegramBot TicketSource = "telegram_bot"
	SourceWhatsApp    TicketSource = "whatsapp"
	SourceInstagram   TicketSource = "instagram"
	SourceMeta        TicketSource = "meta"
)

const (
	StatusAvailable UserStatus = "available"
	StatusBusy      UserStatus = "busy"
	StatusOnBreak   UserStatus = "on break"
	StatusOffline   UserStatus = "offline"
)
