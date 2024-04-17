package service

import (
	"errors"
	"fmt"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/auth/domain/entity"
	_interface "github.com/Point-AI/backend/internal/auth/domain/interface"
	"github.com/Point-AI/backend/internal/auth/service/interface"
	"github.com/Point-AI/backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserServiceImpl struct {
	userRepo      infrastructureInterface.UserRepository
	emailService  _interface.EmailService
	storageClient infrastructureInterface.StorageClient
	config        *config.Config
}

func NewUserServiceImpl(userRepo infrastructureInterface.UserRepository, storageClient infrastructureInterface.StorageClient, emailService _interface.EmailService, cfg *config.Config) _interface.UserService {
	return &UserServiceImpl{
		userRepo:      userRepo,
		emailService:  emailService,
		storageClient: storageClient,
		config:        cfg,
	}
}

func (us *UserServiceImpl) GoogleAuthCallback(code string) (string, error) {
	email, photo, err := utils.ExtractGoogleData(us.config.OAuth2.GoogleClientId, us.config.OAuth2.GoogleClientSecret, code, us.config.OAuth2.GoogleRedirectURL)
	if err != nil {
		return "", err
	}

	oAuth2Token, err := us.userRepo.CreateOauth2User(email, "google")
	if err != nil {
		return "", err
	}

	us.storageClient.SaveFile(photo, us.config.MinIo.BucketName, email)

	return oAuth2Token, nil
}

func (us *UserServiceImpl) GoogleTokens(token string) (string, string, error) {
	user, err := us.userRepo.GetUserByOAuth2Token(token)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", errors.New("user not found")
	}

	accessToken, refreshToken, err := us.setRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (us *UserServiceImpl) Login(email, password string) (string, string, error) {
	user, err := us.userRepo.GetUserByEmail(email)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", errors.New("user not found")
	}

	if !utils.VerifyPassword(user.PasswordHash, password) {
		return "", "", errors.New("invalid password")
	}

	if user.PasswordHash == "" {
		return "", "", errors.New("creating a new password required")
	}

	if !user.IsConfirmed {
		return "", "", errors.New("email not confirmed")
	}

	accessToken, refreshToken, err := us.setRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (us *UserServiceImpl) RegisterUser(email string, password string) error {
	existingUser, err := us.userRepo.GetUserByEmail(email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return errors.New("user already exists")
	}

	confirmToken, err := utils.GenerateToken()
	if err != nil {
		return err
	}

	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	confirmationLink := fmt.Sprintf("%s/confirm?token=%s", us.config.Website.WebURL, confirmToken)
	if err := us.emailService.SendConfirmationEmail(email, confirmationLink); err != nil {
		return err
	}

	if err := us.userRepo.CreateUser(email, passwordHash, confirmToken); err != nil {
		return err
	}

	return nil
}

func (us *UserServiceImpl) ConfirmUser(token string) error {
	user, err := us.userRepo.GetUserByConfirmToken(token)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("invalid confirmation token")
	}

	if err := us.userRepo.ConfirmUser(user.ID); err != nil {
		return err
	}

	return nil
}

func (us *UserServiceImpl) ForgotPassword(email string) error {
	user, err := us.userRepo.GetUserByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	resetToken, err := utils.GenerateJWTToken("reset_token", user.ID, us.config.Auth.JWTSecretKey)
	if err != nil {
		return err
	}

	if err := us.userRepo.SetResetToken(user, resetToken); err != nil {
		return err
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", us.config.Website.WebURL, resetToken)
	if err := us.emailService.SendResetPasswordEmail(email, resetLink); err != nil {
		return err
	}

	return nil
}

func (us *UserServiceImpl) ResetPassword(token, newPassword string) error {
	userId, err := utils.ValidateJWTToken("reset_token", token, us.config.Auth.JWTSecretKey)
	if err != nil {
		return err
	}

	passwordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("error hashing the password")
	}

	if err := us.userRepo.ClearResetToken(userId, passwordHash); err != nil {
		return err
	}

	return nil
}

func (us *UserServiceImpl) RenewAccessToken(refreshToken string) (string, error) {
	userId, err := utils.ValidateJWTToken("refresh_token", refreshToken, us.config.Auth.JWTSecretKey)
	if err != nil {
		return "", err
	}

	user, err := us.userRepo.GetUserById(userId)
	if err != nil {
		return "", err
	}
	if user == nil || user.Tokens.RefreshToken != refreshToken {
		return "", errors.New("invalid refresh token")
	}

	accessToken, err := utils.GenerateJWTToken("access_token", user.ID, us.config.Auth.JWTSecretKey)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (us *UserServiceImpl) Logout(userId primitive.ObjectID) error {
	return us.userRepo.ClearRefreshToken(userId)
}

func (us *UserServiceImpl) setRefreshToken(user *entity.User) (string, string, error) {
	var refreshToken string
	var err error
	if user.Tokens.RefreshToken != "" {
		refreshToken = user.Tokens.RefreshToken
	} else {
		if refreshToken, err = utils.GenerateJWTToken("refresh_token", user.ID, us.config.Auth.JWTSecretKey); err != nil {
			return "", "", err
		}

		if err := us.userRepo.SetRefreshToken(user, refreshToken); err != nil {
			return "", "", err
		}
	}

	accessToken, err := utils.GenerateJWTToken("access_token", user.ID, us.config.Auth.JWTSecretKey)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (us *UserServiceImpl) GetUserProfile(userId primitive.ObjectID) (*entity.User, []byte, error) {
	user, err := us.userRepo.GetUserById(userId)
	if err != nil {
		return &entity.User{}, nil, err
	}

	logo, _ := us.storageClient.LoadFile(user.Email, us.config.MinIo.BucketName)

	return user, logo, nil
}
