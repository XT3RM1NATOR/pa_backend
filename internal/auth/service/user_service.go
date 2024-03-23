package service

import (
	"errors"
	"fmt"
	"github.com/Point-AI/backend/internal/auth/infrastructure/repository"
	"github.com/Point-AI/backend/utils"
)

type UserService struct {
	userRepo     *repository.UserRepository
	emailService *EmailService
	jwtSecretKey string
}

func NewUserService(userRepo *repository.UserRepository, emailService *EmailService, jwtSecretKey string) *UserService {
	return &UserService{
		userRepo:     userRepo,
		emailService: emailService,
		jwtSecretKey: jwtSecretKey,
	}
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

	accessToken, err := utils.GenerateJWTAccessToken(user.ID, us.jwtSecretKey)
	refreshToken, err := utils.GenerateJWTRefreshToken(user.ID, us.jwtSecretKey)
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
	if err := us.userRepo.CreateUser(email, passwordHash, confirmToken); err != nil {
		return err
	}

	confirmationLink := fmt.Sprintf("https://your-domain.com/confirm?token=%s", confirmToken)
	if err := us.emailService.SendConfirmationEmail(email, confirmationLink); err != nil {
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
