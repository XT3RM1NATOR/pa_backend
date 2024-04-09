package service

import (
	"errors"
	"fmt"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/auth/infrastructure/model"
	"github.com/Point-AI/backend/internal/auth/infrastructure/repository"
	"github.com/Point-AI/backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
)

type UserService struct {
	userRepo     *repository.UserRepository
	emailService *EmailService
	config       *config.Config
}

func NewUserService(userRepo *repository.UserRepository, emailService *EmailService, cfg *config.Config) *UserService {
	return &UserService{
		userRepo:     userRepo,
		emailService: emailService,
		config:       cfg,
	}
}

func (us *UserService) GoogleAuthCallback(code string) (string, error) {
	email, err := utils.ExtractGoogleData(us.config.OAuth2.GoogleClientId, us.config.OAuth2.GoogleClientSecret, code)
	log.Println(err)
	if err != nil {
		return "", err
	}

	oAuth2Token, err := us.userRepo.CreateOauth2User(email, "google")
	log.Println(err)
	if err != nil {
		return "", err
	}

	return oAuth2Token, nil
}

func (us *UserService) GoogleTokens(token string) (string, string, error) {
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

func (us *UserService) Login(email, password string) (string, string, error) {
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

func (us *UserService) RegisterUser(email string, password string) error {
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

func (us *UserService) ConfirmUser(token string) error {
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

func (us *UserService) ForgotPassword(email string) error {
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

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", us.config.Website.WebURL, resetToken) // Adjust the URL accordingly
	if err := us.emailService.SendResetPasswordEmail(email, resetLink); err != nil {
		return err
	}

	return nil
}

func (us *UserService) ResetPassword(token, newPassword string) error {
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

func (us *UserService) RenewAccessToken(refreshToken string) (string, error) {
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

func (us *UserService) Logout(userId primitive.ObjectID) error {
	return us.userRepo.ClearRefreshToken(userId)
}

func (us *UserService) setRefreshToken(user *model.User) (string, string, error) {
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
