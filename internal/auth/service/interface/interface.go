package infrastructureInterface

import (
	"github.com/Point-AI/backend/internal/auth/domain/entity"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRepository interface {
	CreateUser(email, passwordHash, confirmToken string) error
	CreateOauth2User(email, authSource string) (string, error)
	GetUserByEmail(email string) (*entity.User, error)
	GetUserById(id primitive.ObjectID) (*entity.User, error)
	GetUserByOAuth2Token(token string) (*entity.User, error)
	GetUserByConfirmToken(token string) (*entity.User, error)
	SetResetToken(user *entity.User, token string) error
	SetRefreshToken(user *entity.User, token string) error
	ClearResetToken(id primitive.ObjectID, password string) error
	ClearRefreshToken(id primitive.ObjectID) error
	ConfirmUser(userId primitive.ObjectID) error
}

type EmailClient interface {
	SendEmail(to, subject, body string) error
}
