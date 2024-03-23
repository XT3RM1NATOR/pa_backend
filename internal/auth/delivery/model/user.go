package model

type UserInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type NewPasswordInput struct {
	Email string `json:"email"`
}

type PasswordResetInput struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}
