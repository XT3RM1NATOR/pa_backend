package test

import (
	"time"
)

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
