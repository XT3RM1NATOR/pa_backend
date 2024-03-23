package model

import "time"

type User struct {
	ID           int       `bson:"_id,omitempty"`
	Email        string    `bson:"email"`
	PasswordHash string    `bson:"passwordHash"`
	IsConfirmed  bool      `bson:"isConfirmed"`
	ConfirmToken string    `bson:"confirmToken"`
	ResetToken   string    `bson:"resetToken"`
	CreatedAt    time.Time `bson:"createdAt"`
}
