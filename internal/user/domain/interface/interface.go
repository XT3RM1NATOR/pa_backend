package _interface

import (
	"github.com/Point-AI/backend/internal/user/domain/entity"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService interface {
	GoogleAuthCallback(code string) (string, error)
	GoogleTokens(token string) (string, string, error)
	Login(email, password string) (string, string, error)
	RegisterUser(email, password, workspaceId, emailHash, name string, logo []byte) error
	ConfirmUser(token string) error
	ForgotPassword(email string) error
	ResetPassword(token, newPassword string) error
	RenewAccessToken(refreshToken string) (string, error)
	Logout(userId primitive.ObjectID) error
	GetUserProfile(userId primitive.ObjectID) (*entity.User, []byte, error)
	UpdateUserProfile(userId primitive.ObjectID, logo []byte, name string) error
	FacebookAuthCallback(code, workspaceId string) error
	UpdateUserStatus(userId primitive.ObjectID, status string) error
}

type EmailService interface {
	SendConfirmationEmail(recipientEmail, confirmationLink string) error
	SendResetPasswordEmail(recipientEmail, resetLink string) error
}

type FileService interface {
	SaveFile(filename string, content []byte) error
	LoadFile(filename string) ([]byte, error)
	UpdateFileName(oldName, newName string) error
	UpdateFile(newFileBytes []byte, fileName string) error
}
