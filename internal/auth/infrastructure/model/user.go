package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Email        string             `bson:"email"`
	PasswordHash string             `bson:"passwordHash"`
	IsConfirmed  bool               `bson:"isConfirmed"`
	AuthSource   string             `bson:"authSource"`
	FullName     string             `bson:"name"`
	Tokens       Tokens             `bson:"tokens"`
	CreatedAt    primitive.DateTime `bson:"createdAt"`
}

type Tokens struct {
	ConfirmToken string `bson:"confirmToken"`
	OAuth2Token  string `bson:"oAuth2Token"`
	ResetToken   string `bson:"resetToken"`
	RefreshToken string `bson:"refreshToken"`
}
