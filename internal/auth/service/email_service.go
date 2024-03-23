package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type EmailService struct {
	APIKey string
}

func NewEmailService(apiKey string) *EmailService {
	return &EmailService{
		APIKey: apiKey,
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
	type MailerSendRequest struct {
		From struct {
			Email string `json:"email"`
		} `json:"from"`
		To []struct {
			Email string `json:"email"`
		} `json:"to"`
		Subject string `json:"subject"`
		Text    string `json:"text"`
		HTML    string `json:"html"`
	}

	requestBody, err := json.Marshal(MailerSendRequest{
		From: struct {
			Email string `json:"email"`
		}{
			Email: "info@your-domain.com", // Adjust this to your sender email
		},
		To: []struct {
			Email string `json:"email"`
		}{
			{
				Email: to,
			},
		},
		Subject: subject,
		Text:    body,
		HTML:    body,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.mailersend.com/v1/email", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+es.APIKey)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Handle non-200 status code
		return errors.New("failed to send email: " + resp.Status)
	}

	return nil
}
