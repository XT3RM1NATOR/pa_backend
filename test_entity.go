package test

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Project struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Logo        string               `json:"logo"`
	Emoji       string               `json:"emoji"`
	OwnerID     string               `json:"owner_id"`
	Team        []primitive.ObjectID `json:"team"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Email        string             `bson:"email"`
	PasswordHash string             `bson:"passwordHash"`
	IsConfirmed  bool               `bson:"isConfirmed"`
	AuthSource   string             `bson:"authSource"`
	FullName     string             `bson:"name"`
	Tokens       Tokens             `bson:"tokens"`
	UserRole     UserRole           `bson:"user_role"`
	CreatedAt    primitive.DateTime `bson:"createdAt"`
}

type Tokens struct {
	ConfirmToken string `bson:"confirmToken"`
	OAuth2Token  string `bson:"oAuth2Token"`
	ResetToken   string `bson:"resetToken"`
	RefreshToken string `bson:"refreshToken"`
}

type Integrations struct {
	ID        string
	ProjectID string
	Telegram  *TelegramIntegration
	Crisp     *CrispIntegration
	Meta      *MetaIntegration
	Zendesk   *ZendeskIntegration
	API       *APIIntegration
	HelpScout *HelpScoutIntegration
	WhatsApp  *WhatsAppIntegration
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TelegramIntegration struct {
	BotToken string
	IsActive bool
}

type MetaIntegration struct {
	AuthToken string
	PageID    string
	IsActive  bool
}

type CrispIntegration struct {
	WebsiteID string
	TokenID   string
	TokenKey  string
	LunaID    string
	IsActive  bool
}

type ZendeskIntegration struct {
	Subdomain string
	Email     string
	ApiToken  string
	LunaID    string
	AppID     string
	TokenID   string
	TokenKey  string
	IsActive  bool
}

type APIIntegration struct {
	WebhookURL    string
	WebhookSecret string
	PasswordHash  string
	IsActive      bool
}

type HelpScoutIntegration struct {
	AccessToken  string
	RefreshToken string
	WebhookID    string
	MailboxID    int64
	IsActive     bool
}

type WhatsAppIntegration struct {
	InstanceId string
	IsActive   bool
}

type UserRole string

const (
	Owner  UserRole = "Owner"
	Admin  UserRole = "Admin"
	Member UserRole = "Member"
)
