package client

import (
	"fmt"
)

type AuthStatus string

const (
	StatusPhone         AuthStatus = "phone"
	StatusCode          AuthStatus = "code"
	StatusPassword      AuthStatus = "password"
	StatusPasswordRetry AuthStatus = "password_retry"
)

type TelegramAuthConversator struct {
	phoneChan  chan string
	codeChan   chan string
	passwdChan chan string
	Status     AuthStatus
}

func newTelegramAuthConversator() *TelegramAuthConversator {
	return &TelegramAuthConversator{
		phoneChan:  make(chan string),
		codeChan:   make(chan string),
		passwdChan: make(chan string),
		Status:     "",
	}
}

func (tac *TelegramAuthConversator) AskPhoneNumber() (string, error) {
	tac.Status = StatusPhone
	fmt.Println("waiting for phone number...")
	phone := <-tac.phoneChan
	return phone, nil
}

func (tac *TelegramAuthConversator) AskCode() (string, error) {
	tac.Status = StatusCode
	fmt.Println("waiting for OTP code...")
	code := <-tac.codeChan
	return code, nil
}

func (tac *TelegramAuthConversator) AskPassword() (string, error) {
	tac.Status = StatusPassword
	fmt.Println("waiting for 2FA password...")
	passwd := <-tac.passwdChan
	return passwd, nil
}

func (tac *TelegramAuthConversator) RetryPassword(attemptsLeft int) (string, error) {
	tac.Status = StatusPasswordRetry
	fmt.Printf("Incorrect password. %d attempts left. Please try again...\n", attemptsLeft)
	passwd := <-tac.passwdChan
	return passwd, nil
}

func (tac *TelegramAuthConversator) ReceivePhone(phone string) {
	tac.phoneChan <- phone
}

func (tac *TelegramAuthConversator) ReceiveCode(code string) {
	tac.codeChan <- code
}

func (tac *TelegramAuthConversator) ReceivePasswd(passwd string) {
	tac.passwdChan <- passwd
}
