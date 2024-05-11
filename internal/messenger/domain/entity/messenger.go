package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	Id             primitive.ObjectID `bson:"_id,omitempty"`
	Email          string             `bson:"email"`
	PasswordHash   string             `bson:"password"`
	IsConfirmed    bool               `bson:"is_confirmed"`
	AuthSource     string             `bson:"auth_source"`
	FullName       string             `bson:"name"`
	PendingInvites []string           `bson:"pending_invites"`
	Tokens         Tokens             `bson:"tokens"`
	CreatedAt      time.Time          `bson:"created_at"`
}

type Tokens struct {
	ConfirmToken string `bson:"confirm_token"`
	OAuth2Token  string `bson:"oauth2_token"`
	ResetToken   string `bson:"reset_token"`
	RefreshToken string `bson:"refresh_token"`
}

type Workspace struct {
	Id            primitive.ObjectID                           `bson:"_id,omitempty"`
	WorkspaceId   string                                       `bson:"workspace_id"`
	Name          string                                       `bson:"name"`
	Team          map[primitive.ObjectID]WorkspaceRole         `bson:"team"`
	PendingTeam   map[string]WorkspaceRole                     `bson:"pending"`
	InternalTeams map[string]map[primitive.ObjectID]UserStatus `bson:"teams"`
	FirstTeam     string                                       `bson:"first_team"`
	Integrations  Integrations                                 `bson:"integrations"`
	Folders       map[string][]string                          `bson:"folders"`
	CreatedAt     time.Time                                    `bson:"created_at"`
}

type Chat struct {
	Id          primitive.ObjectID `bson:"_id,omitempty"`
	UserId      primitive.ObjectID `bson:"user_id"`
	WorkspaceId primitive.ObjectID `bson:"workspace_id"`
	ChatId      string             `bson:"chat_id"`
	TgClientId  int                `bson:"tg_user_id"`
	TgChatId    int                `bson:"tg_chat_id"`
	Tags        []string           `bson:"tags"`
	Notes       []Note             `bson:"notes"`
	Tickets     []Ticket           `bson:"tickets"`
	LastMessage Message            `bson:"last_message"`
	Source      ChatSource         `bson:"source"`
	IsImported  bool               `bson:"is_imported"`
	CreatedAt   time.Time          `bson:"created_at"`
}

type Note struct {
	UserId    primitive.ObjectID `bson:"user_id"`
	Text      string             `bson:"text"`
	CreatedAt time.Time          `bson:"created_at"`
	NoteId    string             `bson:"note_id"`
}

type Ticket struct {
	Id         primitive.ObjectID `bson:"_id,omitempty"`
	TicketId   string             `bson:"ticket_id"`
	Subject    string             `bson:"subject"`
	Notes      []Note             `bson:"notes"`
	Messages   []Message          `bson:"messages"`
	Status     TicketStatus       `bson:"status"`
	CreatedAt  time.Time          `bson:"created_at"`
	ResolvedAt time.Time          `bson:"resolved_at"`
}

type Message struct {
	Id              primitive.ObjectID `bson:"_id,omitempty"`
	SenderId        primitive.ObjectID `bson:"sender_id"`
	MessageId       string             `bson:"message_id"`
	MessageIdClient int                `bson:"message_id_client"`
	Message         string             `bson:"message"`
	From            string             `bson:"from"`
	Type            MessageType        `bson:"type"`
	CreatedAt       time.Time          `bson:"created_at"`
}

type Integrations struct {
	Id          primitive.ObjectID          `bson:"_id"`
	TelegramBot *TelegramBotIntegration     `bson:"telegram_bot"`
	Telegram    *TelegramAccountIntegration `bson:"telegram"`
	Meta        *MetaIntegration            `bson:"meta"`
	Instagram   *InstagramIntegration       `bson:"instagram"`
	WhatsApp    *WhatsAppIntegration        `bson:"whatsapp"`
	CreatedAt   time.Time                   `bson:"created_at"`
}

type TelegramBotIntegration struct {
	BotToken string `bson:"bot_token"`
	IsActive bool   `bson:"is_active"`
}

type TelegramAccountIntegration struct {
	Session     string `bson:"session"`
	PhoneNumber string `bson:"phone_number"`
	IsActive    bool   `bson:"is_active"`
}

type MetaIntegration struct {
	AccessToken  string `bson:"access_token"`
	RefreshToken string `bson:"refresh_token"`
	PageId       string `bson:"page_id"`
	IsActive     bool   `bson:"is_active"`
}

type InstagramIntegration struct {
	AccessToken  string `bson:"access_token"`
	RefreshToken string `bson:"refresh_token"`
	PageId       string `bson:"page_id"`
	IsActive     bool   `bson:"is_active"`
}

type WhatsAppIntegration struct {
	InstanceId string `bson:"instance_id"`
	IsActive   bool   `bson:"is_active"`
}

type MessageType string
type WorkspaceRole string
type ChatSource string
type TicketStatus string
type UserStatus string

const (
	TypeChatNote   MessageType = "chat_note"
	TypeTicketNote MessageType = "ticket_note"
	TypeText       MessageType = "text"
	TypeImage      MessageType = "image"
	TypeAudio      MessageType = "audio"
	TypeDocument   MessageType = "document"
	TypeSticker    MessageType = "sticker"
	TypeVideo      MessageType = "video"
	TypeVoice      MessageType = "voice"
	TypeVideoNote  MessageType = "video_note"
	TypeGif        MessageType = "gif"
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
	SourceTelegram    ChatSource = "telegram"
	SourceTelegramBot ChatSource = "telegram_bot"
	SourceWhatsApp    ChatSource = "whatsapp"
	SourceInstagram   ChatSource = "instagram"
	SourceMeta        ChatSource = "meta"
)

const (
	StatusAvailable UserStatus = "available"
	StatusBusy      UserStatus = "busy"
	StatusOffline   UserStatus = "offline"
)
