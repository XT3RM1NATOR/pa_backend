package model

type UserRequest struct {
	Email    string `json:"email"`
	FullName string `json:"fullName"`
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

type AccessTokenRequest struct {
	AccessToken string `json:"access_token"`
}

type OAuth2TokenRequest struct {
	OAuth2Token string `json:"oAuth2Token"`
}
