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
	Token        Token              `bson:"token"`
	CreatedAt    primitive.DateTime `bson:"createdAt"`
}

type Token struct {
	ConfirmToken string `bson:"confirmToken"`
	ResetToken   string `bson:"resetToken"`
	RefreshToken string `bson:"refreshToken"`
}
