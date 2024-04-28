package client

import (
	"context"
	"errors"
	"github.com/Point-AI/backend/config"
	infrastructureInterface "github.com/Point-AI/backend/internal/messenger/service/interface"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"strconv"
)

type TelegramClient struct {
	client *telegram.Client
	config *config.Config
}

func NewTelegramClientImpl(config *config.Config) infrastructureInterface.TelegramClient {
	clientId, err := strconv.Atoi(config.OAuth2.TelegramClientId)
	if err != nil {
		panic(err)
	}

	return &TelegramClient{
		config: config,
		client: telegram.NewClient(clientId, config.OAuth2.TelegramClientSecret, telegram.Options{}),
	}
}

func (tc *TelegramClient) Authenticate(ctx context.Context, phoneNumber string) (*tg.AuthSentCode, error) {
	var sentCode *tg.AuthSentCode
	clientId, err := strconv.Atoi(tc.config.OAuth2.TelegramClientId)
	if err != nil {
		return sentCode, err
	}

	err = tc.client.Run(ctx, func(ctx context.Context) error {
		api := tc.client.API()
		result, err := api.AuthSendCode(ctx, &tg.AuthSendCodeRequest{
			PhoneNumber: phoneNumber,
			APIID:       clientId,
			APIHash:     tc.config.OAuth2.TelegramClientSecret,
			Settings:    tg.CodeSettings{AllowFlashcall: false, AllowAppHash: true},
		})
		if err != nil {
			return err
		}

		code, ok := result.(*tg.AuthSentCode)
		if !ok {
			return errors.New("failed to assert the type")
		}

		sentCode = code
		return nil
	})
	if err != nil {
		return &tg.AuthSentCode{}, err
	}

	return sentCode, err
}

func (tc *TelegramClient) SignIn(ctx context.Context, phoneNumber, phoneCodeHash, phoneCode string) (*tg.AuthAuthorization, error) {
	var authResult *tg.AuthAuthorization
	err := tc.client.Run(ctx, func(ctx context.Context) error {
		api := tc.client.API()
		result, err := api.AuthSignIn(ctx, &tg.AuthSignInRequest{
			PhoneNumber:   phoneNumber,
			PhoneCodeHash: phoneCodeHash,
			PhoneCode:     phoneCode,
		})
		if err != nil {
			return err
		}

		authorization, ok := result.(*tg.AuthAuthorization)
		if !ok {
			return errors.New("failed to assert a type")
		}

		authResult = authorization

		return nil
	})
	return authResult, err
}

func (tc *TelegramClient) SignInFA(ctx context.Context, password string) (*tg.AuthAuthorization, error) {
	var authResult *tg.AuthAuthorization
	err := tc.client.Run(ctx, func(ctx context.Context) error {
		//api := tc.client.API()
		//api.AuthSi
		return nil
	})

	return authResult, err
}
