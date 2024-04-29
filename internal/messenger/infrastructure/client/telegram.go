package client

import (
	"errors"
	"github.com/Point-AI/backend/config"
	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/sessionMaker"
	"log"
	"strconv"
	"sync"
)

type TelegramClientManager struct {
	clients          map[string]*gotgproto.Client
	authConversators map[string]*TelegramAuthConversator
	config           *config.Config
	mu               sync.RWMutex
}

func NewTelegramClientManagerImpl(cfg *config.Config) *TelegramClientManager {
	return &TelegramClientManager{
		config:  cfg,
		clients: make(map[string]*gotgproto.Client),
	}
}

func (tcm *TelegramClientManager) CreateClient(phone, workspaceId string) error {
	if _, exists := tcm.clients[workspaceId]; exists {
		return errors.New("the client already exists")
	}

	clientId, err := strconv.Atoi(tcm.config.OAuth2.TelegramClientId)
	if err != nil {
		return err
	}

	authConversator := newTelegramAuthConversator()

	client, err := gotgproto.NewClient(
		clientId,
		tcm.config.OAuth2.TelegramClientSecret,
		gotgproto.ClientType{Phone: phone},
		&gotgproto.ClientOpts{
			AuthConversator: authConversator,
			Session:         sessionMaker.SimpleSession(),
		},
	)

	if err != nil {
		log.Fatalf("failed to create client for phone %s: %v", phone, err)
	}

	go func() {
		client.Dispatcher.AddHandler(handlers.NewMessage(filters.Message.All, echo))
		client.Idle()
	}()

	tcm.SetClient(workspaceId, client)
	tcm.SetAuthConversator(workspaceId, authConversator)
	return nil
}

func (tcm *TelegramClientManager) CreateClientBySession(session, phone, workspaceId string) error {
	clientId, err := strconv.Atoi(tcm.config.OAuth2.TelegramClientId)
	if err != nil {
		return err
	}

	authConversator := newTelegramAuthConversator()

	client, err := gotgproto.NewClient(
		clientId,
		tcm.config.OAuth2.TelegramClientSecret,
		gotgproto.ClientType{Phone: phone},
		&gotgproto.ClientOpts{
			AuthConversator: authConversator,
			Session:         sessionMaker.StringSession(session),
		},
	)

	if err != nil {
		log.Fatalf("failed to create client for phone %s: %v", phone, err)
	}

	go func() {
		client.Dispatcher.AddHandler(handlers.NewMessage(filters.Message.All, echo))
		client.Idle()
	}()

	tcm.SetClient(workspaceId, client)
	tcm.SetAuthConversator(workspaceId, authConversator)
	return nil
}

func (tcm *TelegramClientManager) GetClient(workspaceId string) (*gotgproto.Client, bool) {
	tcm.mu.RLock()
	defer tcm.mu.RUnlock()

	client, exists := tcm.clients[workspaceId]
	return client, exists
}

func (tcm *TelegramClientManager) GetAuthConversator(workspaceId string) (*TelegramAuthConversator, bool) {
	tcm.mu.RLock()
	defer tcm.mu.RUnlock()

	authConversator, exists := tcm.authConversators[workspaceId]
	return authConversator, exists
}

func (tcm *TelegramClientManager) SetClient(workspaceId string, client *gotgproto.Client) {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()

	tcm.clients[workspaceId] = client
}

func (tcm *TelegramClientManager) SetAuthConversator(workspaceId string, authConversator *TelegramAuthConversator) {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()

	tcm.authConversators[workspaceId] = authConversator
}

func echo(ctx *ext.Context, update *ext.Update) error {
	msg := update.EffectiveMessage
	log.Printf("Received message from %v: %s", msg.FromID, msg.Text)
	_, err := ctx.Reply(update, msg.Text, nil)
	return err
}
