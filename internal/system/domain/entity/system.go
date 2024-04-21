package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

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
	WorkspaceId string                               `bson:"workspace_id"`
	CreatedAt   primitive.DateTime                   `bson:"created_at"`
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
