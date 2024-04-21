package client

import (
	"errors"
	"github.com/Point-AI/backend/internal/user/service/interface"
	"net/smtp"
)

type EmailClientImpl struct {
	SMTPUsername string
	SMTPPassword string
	SMTPHost     string
	SMTPPort     string
}

func NewEmailClientImpl(SMTPUsername string, SMTPPassword string, SMTPHost string, SMTPPort string) infrastructureInterface.EmailClient {
	return &EmailClientImpl{
		SMTPUsername: SMTPUsername,
		SMTPPassword: SMTPPassword,
		SMTPHost:     SMTPHost,
		SMTPPort:     SMTPPort,
	}
}

func (c *EmailClientImpl) SendEmail(to, subject, body string) error {
	auth := smtp.PlainAuth("", c.SMTPUsername, c.SMTPPassword, c.SMTPHost)

	message := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body)

	err := smtp.SendMail(c.SMTPHost+":"+c.SMTPPort, auth, c.SMTPUsername, []string{to}, message)
	if err != nil {
		return errors.New("failed to send email: " + err.Error())
	}

	return nil
}
