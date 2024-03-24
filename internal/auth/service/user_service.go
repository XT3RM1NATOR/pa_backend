package service

import (
	"errors"
	"fmt"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/auth/infrastructure/repository"
	"github.com/Point-AI/backend/utils"
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

func (us *UserService) GoogleAuthCallback(code string) (string, string, error) {
	email, fullName, err := utils.ExtractGoogleData(us.config.OAuth2.GoogleClientId, us.config.OAuth2.GoogleClientSecret, code)
	if err != nil {
		return "", "", err
	}

	userId, err := us.userRepo.CreateOauth2User(email, "google", fullName)
	if err != nil {
		return "", "", err
	}

	accessToken, err := utils.GenerateJWTAccessToken(userId, us.config.Auth.JWTSecretKey)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := utils.GenerateJWTRefreshToken(userId, us.config.Auth.JWTSecretKey)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// GoogleLogin TODO: add the actual state, which will be random
func (us *UserService) GoogleLogin() (string, error) {
	stateToken, err := utils.GenerateJWTStateToken(us.config.Auth.JWTSecretKey, us.config.OAuth2.StateText)
	if err != nil {
		return "", err
	}

	authURL := "https://accounts.google.com/o/oauth2/auth" +
		"?client_id=" + us.config.OAuth2.GoogleClientId +
		"&redirect_uri=" + us.config.OAuth2.GoogleRedirectURI +
		"&response_type=code" +
		"&scope=openid%20email%20profile" +
		"&state=" + stateToken

	return authURL, nil
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

	if !user.IsConfirmed {
		return "", "", errors.New("email not confirmed")
	}

	accessToken, err := utils.GenerateJWTAccessToken(user.ID, us.config.Auth.JWTSecretKey)
	refreshToken, err := utils.GenerateJWTRefreshToken(user.ID, us.config.Auth.JWTSecretKey)
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

	confirmationLink := fmt.Sprintf("https://your-domain.com/confirm?token=%s", confirmToken)
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

	if err := us.userRepo.ConfirmUser(user); err != nil {
		return err
	}

	return nil
}

func (us *UserService) ForgotPassword(email string) error {
	resetToken, err := utils.GenerateToken()
	if err != nil {
		return err
	}

	user, err := us.userRepo.GetUserByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	if err := us.userRepo.SetResetToken(user, resetToken); err != nil {
		return err
	}

	resetLink := fmt.Sprintf("https://your-domain.com/reset-password?token=%s", resetToken) // Adjust the URL accordingly
	if err := us.emailService.SendResetPasswordEmail(email, resetLink); err != nil {
		return err
	}

	return nil
}

func (us *UserService) ResetPassword(token, newPassword string) error {
	user, err := us.userRepo.GetUserByResetToken(token)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("invalid reset password token")
	}
	passwordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("error hashing the password")
	}

	if err := us.userRepo.ClearResetToken(user, passwordHash); err != nil {
		return err
	}

	return nil
}
