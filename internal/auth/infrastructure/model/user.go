package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Email        string             `bson:"email"`
	PasswordHash string             `bson:"passwordHash"`
	IsConfirmed  bool               `bson:"isConfirmed"`
	ConfirmToken string             `bson:"confirmToken"`
	ResetToken   string             `bson:"resetToken"`
	CreatedAt    primitive.DateTime `bson:"createdAt"`
}
