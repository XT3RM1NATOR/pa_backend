package model

type UserInput struct {
	Email    string `json:"email"`
	FullName string `json:"fullName"`
	Password string `json:"password"`
}

type ForgotPasswordInput struct {
	Email string `json:"email"`
}

type PasswordResetInput struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type RenewAccessTokenInput struct {
	RefreshToken string `json:"refresh_token"`
}

type AccessTokenInput struct {
	AccessToken string `json:"access_token"`
}
