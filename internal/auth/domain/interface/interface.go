package _interface

import (
	"github.com/Point-AI/backend/internal/auth/domain/entity"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService interface {
	GoogleAuthCallback(code string) (string, error)
	GoogleTokens(token string) (string, string, error)
	Login(email, password string) (string, string, error)
	RegisterUser(email string, password string) error
	ConfirmUser(token string) error
	ForgotPassword(email string) error
	ResetPassword(token, newPassword string) error
	RenewAccessToken(refreshToken string) (string, error)
	Logout(userId primitive.ObjectID) error
	GetUserProfile(userId primitive.ObjectID) (*entity.User, []byte, error)
}

type EmailService interface {
	SendConfirmationEmail(recipientEmail, confirmationLink string) error
	SendResetPasswordEmail(recipientEmail, resetLink string) error
}
