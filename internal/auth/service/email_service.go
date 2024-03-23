package service

import (
	"errors"
	"fmt"
	"net/smtp"
)

type EmailService struct {
	SMTPUsername string
	SMTPPassword string
	SMTPHost     string
	SMTPPort     string
	SenderEmail  string
}

func NewEmailService(username, password, host, port, senderEmail string) *EmailService {
	return &EmailService{
		SMTPUsername: username,
		SMTPPassword: password,
		SMTPHost:     host,
		SMTPPort:     port,
		SenderEmail:  senderEmail,
	}
}

func (es *EmailService) SendConfirmationEmail(recipientEmail, confirmationLink string) error {
	subject := "Confirm your email"
	body := fmt.Sprintf("Click the following link to confirm your email: %s", confirmationLink)
	return es.sendEmail(recipientEmail, subject, body)
}

func (es *EmailService) SendResetPasswordEmail(recipientEmail, resetLink string) error {
	subject := "Reset your password"
	body := fmt.Sprintf("Click the following link to reset your password: %s", resetLink)
	return es.sendEmail(recipientEmail, subject, body)
}
func (es *EmailService) sendEmail(to, subject, body string) error {
	auth := smtp.PlainAuth("", es.SMTPUsername, es.SMTPPassword, es.SMTPHost)

	message := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body)

	err := smtp.SendMail(es.SMTPHost+":"+es.SMTPPort, auth, es.SenderEmail, []string{to}, message)
	if err != nil {
		return errors.New("failed to send email: " + err.Error())
	}

	return nil
}
