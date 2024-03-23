package model

import "time"

type User struct {
	ID         string    `json:"id" bson:"_id"`
	Email      string    `json:"email" validate:"required" bson:"email"`
	Password   string    `json:"password" validate:"required" bson:"password"`
	Username   string    `json:"username" bson:"username"`
	TokenHash  string    `json:"tokenhash" bson:"tokenhash"`
	IsVerified bool      `json:"isverified" bson:"isverified"`
	CreatedAt  time.Time `json:"createdat" bson:"createdat"`
	UpdatedAt  time.Time `json:"updatedat" bson:"updatedat"`
}
