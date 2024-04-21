package service

import (
	"fmt"
	_interface "github.com/Point-AI/backend/internal/system/domain/interface"
	infrastructureInterface "github.com/Point-AI/backend/internal/system/service/interface"
)

type EmailServiceImpl struct {
	emailClient infrastructureInterface.EmailClient
}

func NewEmailServiceImpl(emailClient infrastructureInterface.EmailClient) _interface.EmailService {
	return &EmailServiceImpl{
		emailClient: emailClient,
	}
}

func (es *EmailServiceImpl) SendInvitationEmail(recipientEmail, inviteLink string) error {
	subject := "Confirm your email"
	body := fmt.Sprintf("Click the following link to confirm your email: %s", inviteLink)
	return es.emailClient.SendEmail(recipientEmail, subject, body)
}
