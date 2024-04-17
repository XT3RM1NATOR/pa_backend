package utils

import (
	"context"
	"encoding/json"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io"
)

func ExtractGoogleData(clientID, clientSecret, code, redirectURL string) (string, []byte, error) {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  redirectURL,
		Scopes:       []string{"email", "profile"},
	}

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return "", nil, err
	}

	client := config.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	var profile struct {
		Email   string `json:"email"`
		Picture string `json:"picture"`
	}
	err = json.Unmarshal(body, &profile)
	if err != nil {
		return "", nil, err
	}

	resp, err = client.Get(profile.Picture)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	pictureData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	return profile.Email, pictureData, nil
}
