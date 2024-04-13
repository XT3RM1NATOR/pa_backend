package service

import (
	"fmt"
	_interface "github.com/Point-AI/backend/internal/auth/domain/interface"
	infrastructureInterface "github.com/Point-AI/backend/internal/auth/service/interface"
)

type EmailServiceImpl struct {
	emailClient infrastructureInterface.EmailClient
}

func NewEmailServiceImpl(emailClient infrastructureInterface.EmailClient) _interface.EmailService {
	return &EmailServiceImpl{
		emailClient: emailClient,
	}
}

func (es *EmailServiceImpl) SendConfirmationEmail(recipientEmail, confirmationLink string) error {
	subject := "Confirm your email"
	body := fmt.Sprintf("Click the following link to confirm your email: %s", confirmationLink)
	return es.emailClient.SendEmail(recipientEmail, subject, body)
}

func (es *EmailServiceImpl) SendResetPasswordEmail(recipientEmail, resetLink string) error {
	subject := "Reset your password"
	body := fmt.Sprintf("Click the following link to reset your password: %s", resetLink)
	return es.emailClient.SendEmail(recipientEmail, subject, body)
}
