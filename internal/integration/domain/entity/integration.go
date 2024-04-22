package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type WorkspaceRole string

const (
	RoleAdmin  WorkspaceRole = "admin"
	RoleMember WorkspaceRole = "member"
	RoleOwner  WorkspaceRole = "owner"
)

type Workspace struct {
	Id          primitive.ObjectID                   `bson:"_id,omitempty"`
	Name        string                               `bson:"name"`
	Team        map[primitive.ObjectID]WorkspaceRole `bson:"team"`
	PendingTeam map[string]WorkspaceRole             `bson:"pending"`
	Integration integrations                         `bson:"integrations"`
	WorkspaceId string                               `bson:"workspace_id"`
	CreatedAt   primitive.DateTime                   `bson:"created_at"`
}

type telegramIntegration struct {
	BotToken string `bson:"bot_token"`
	IsActive bool   `bson:"is_active"`
}

type metaIntegration struct {
	AuthToken string `bson:"auth_token"`
	PageID    string `bson:"page_id"`
	IsActive  bool   `bson:"is_active"`
}

type whatsAppIntegration struct {
	InstanceId string `bson:"instance_id"`
	IsActive   bool   `bson:"is_active"`
}

type integrations struct {
	Id        primitive.ObjectID   `bson:"_id"`
	Telegram  *telegramIntegration `bson:"telegram"`
	Meta      *metaIntegration     `bson:"meta"`
	WhatsApp  *whatsAppIntegration `bson:"whatsapp"`
	CreatedAt time.Time            `bson:"created_at"`
}

type User struct {
	Id             primitive.ObjectID `bson:"_id,omitempty"`
	Email          string             `bson:"email"`
	PasswordHash   string             `bson:"password"`
	IsConfirmed    bool               `bson:"is_confirmed"`
	AuthSource     string             `bson:"auth_source"`
	FullName       string             `bson:"name"`
	PendingInvites []string           `bson:"pending_invites"`
	Tokens         tokens             `bson:"tokens"`
	CreatedAt      primitive.DateTime `bson:"created_at"`
}

type tokens struct {
	ConfirmToken string `bson:"confirm_token"`
	OAuth2Token  string `bson:"oauth2_token"`
	ResetToken   string `bson:"reset_token"`
	RefreshToken string `bson:"refresh_token"`
}
