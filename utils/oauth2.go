package utils

import (
	"context"
	"encoding/json"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io"
)

func ExtractGoogleData(clientID, clientSecret, code string) (string, error) {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "https://31c8-195-158-30-66.ngrok-free.app/auth/oauth2/gooogle/callback",
		Scopes:       []string{"email"},
	}

	token, err := config.Exchange(context.TODO(), code)
	if err != nil {
		return "", err
	}

	client := config.Client(context.TODO(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var profile struct {
		Email string `json:"email"`
	}
	err = json.Unmarshal(body, &profile)
	if err != nil {
		return "", err
	}

	return profile.Email, nil
}
