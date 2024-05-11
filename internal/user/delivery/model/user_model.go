package model

import (
	"time"
)

// Requests

type UserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type PasswordResetRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type RenewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type OAuth2TokenRequest struct {
	OAuth2Token string `json:"oAuth2Token"`
}

type UserUpdateProfileRequest struct {
	FullName string `json:"name"`
	Logo     []byte `json:"logo"`
}

// Responses

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"` // Optional field for some responses
}

type URLResponse struct {
	URL string `json:"url"`
}

type UserProfileResponse struct {
	Email     string    `json:"email"`
	FullName  string    `json:"name"`
	Logo      []byte    `json:"logo"`
	CreatedAt time.Time `json:"created_at"`
}
