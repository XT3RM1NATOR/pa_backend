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
	CreatedAt      primitive.DateTime `bson:"created_at"`
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
	CreatedAt     primitive.DateTime                           `bson:"created_at"`
}

type Chat struct {
	Id          primitive.ObjectID `bson:"_id,omitempty"`
	UserId      primitive.ObjectID `bson:"user_id"`
	WorkspaceId primitive.ObjectID `bson:"workspace_id"`
	TgClientId  int                `bson:"tg_user_id"`
	Language    LanguageType       `bson:"language"`
	Tags        []string           `bson:"tags"`
	Comments    []Comment          `bson:"comments"`
	Tickets     []Ticket           `bson:"tickets"`
	Source      ChatSource         `bson:"source"`
	CreatedAt   time.Time          `bson:"created_at"`
}

type Comment struct {
	UserId    primitive.ObjectID `bson:"user_id"`
	Text      string             `bson:"text"`
	CreatedAt primitive.DateTime `bson:"created_at"`
	CommentId string             `bson:"comment_id"`
}

type Ticket struct {
	Id                  primitive.ObjectID    `bson:"_id,omitempty"`
	TicketId            string                `bson:"ticket_id"`
	Subject             string                `bson:"subject"`
	Comments            []Comment             `bson:"comments"`
	IntegrationMessages []IntegrationsMessage `bson:"integration_messages"`
	ResponseMessages    []ResponseMessage     `bson:"response_messages"`
	Status              TicketStatus          `bson:"status"`
	CreatedAt           primitive.DateTime    `bson:"created_at"`
	ResolvedAt          primitive.DateTime    `bson:"resolved_at,omitempty"`
}

type ResponseMessage struct {
	Id        primitive.ObjectID  `bson:"_id,omitempty"`
	SenderId  primitive.ObjectID  `bson:"sender_id,omitempty"`
	Message   string              `bson:"message"`
	Type      MessageType         `bson:"type"`
	CreatedAt *primitive.DateTime `bson:"created_at,omitempty"`
}

type IntegrationsMessage struct {
	Id          primitive.ObjectID `bson:"_id,omitempty"`
	MessageId   int                `bson:"message_id"`
	FileIdStr   string             `bson:"file_id_str"`
	FileIdInt64 int64              `bson:"file_id_int64"`
	Message     string             `bson:"message"`
	From        string             `bson:"from"`
	Type        MessageType        `bson:"type"`
	CreatedAt   primitive.DateTime `bson:"created_at,omitempty"`
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
type LanguageType string
type WorkspaceRole string
type ChatSource string
type TicketStatus string
type UserStatus string

const (
	TypeText                  MessageType = "text"
	TypeImage                 MessageType = "image"
	TypeAudio                 MessageType = "audio"
	TypeDocument              MessageType = "document"
	TypeSticker               MessageType = "sticker"
	TypeVideo                 MessageType = "video"
	TypeVoice                 MessageType = "voice"
	TypeVideoNote             MessageType = "video_note"
	TypeGif                   MessageType = "gif"
	TypeLocation              MessageType = "location"
	TypeContact               MessageType = "contact"
	TypeVenue                 MessageType = "venue"
	TypeNewChatMembers        MessageType = "new_chat_members"
	TypeLeftChatMember        MessageType = "left_chat_member"
	TypeNewChatTitle          MessageType = "new_chat_title"
	TypeNewChatPhoto          MessageType = "new_chat_photo"
	TypeDeleteChatPhoto       MessageType = "delete_chat_photo"
	TypeGroupChatCreated      MessageType = "group_chat_created"
	TypeSupergroupChatCreated MessageType = "supergroup_chat_created"
	TypeChannelChatCreated    MessageType = "channel_chat_created"
	TypeMigrateToChatID       MessageType = "migrate_to_chat_id"
	TypeMigrateFromChatID     MessageType = "migrate_from_chat_id"
	TypePinnedMessage         MessageType = "pinned_message"
	TypeInvoice               MessageType = "invoice"
	TypeSuccessfulPayment     MessageType = "successful_payment"
	TypeConnectedWebsite      MessageType = "connected_website"
	TypePassportData          MessageType = "passport_data"
	TypeAnimation             MessageType = "animation"
	TypeGame                  MessageType = "game"
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

const (
	English LanguageType = "english"
	Russian LanguageType = "russian"
	Uzbek   LanguageType = "uzbek"
)
