package service

import (
	"encoding/json"
	"errors"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/messenger/delivery/model"
	"github.com/Point-AI/backend/internal/messenger/domain/entity"
	_interface "github.com/Point-AI/backend/internal/messenger/domain/interface"
	infrastructureInterface "github.com/Point-AI/backend/internal/messenger/service/interface"
	"github.com/Point-AI/backend/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/gotd/td/tg"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
)

type MessengerServiceImpl struct {
	messengerRepo            infrastructureInterface.MessengerRepository
	telegramBotClientManager infrastructureInterface.TelegramBotClientManager
	telegramClientManager    infrastructureInterface.TelegramClientManager
	websocketService         _interface.WebsocketService
	config                   *config.Config
}

func NewMessengerServiceImpl(cfg *config.Config, messengerRepo infrastructureInterface.MessengerRepository, websocketService _interface.WebsocketService, telegramBotClientManager infrastructureInterface.TelegramBotClientManager, telegramClientManager infrastructureInterface.TelegramClientManager) _interface.MessengerService {
	return &MessengerServiceImpl{
		messengerRepo:            messengerRepo,
		telegramBotClientManager: telegramBotClientManager,
		telegramClientManager:    telegramClientManager,
		websocketService:         websocketService,
		config:                   cfg,
	}
}

func (ms *MessengerServiceImpl) ValidateUserInWorkspace(userId primitive.ObjectID, workspaceId string) error {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	if _, exists := workspace.Team[userId]; exists {
		return nil
	}

	return errors.New("user does not have the permissions")
}

func (ms *MessengerServiceImpl) findMinAssignee(assignedCount map[primitive.ObjectID]int) (primitive.ObjectID, error) {
	var minAssignments int
	var minAssignee primitive.ObjectID
	first := true

	for assignee, count := range assignedCount {
		if first || count < minAssignments {
			minAssignee = assignee
			minAssignments = count
			first = false
		}
	}

	if first {
		return primitive.NilObjectID, errors.New("no tickets are assigned")
	}
	return minAssignee, nil
}

func (ms *MessengerServiceImpl) createMessageEntitiesFromTelegramBot(botToken string, message *tgbotapi.Update) (entity.IntegrationsMessage, *model.MessageResponse) {
	messageType, fileId := utils.GetMessageTypeAndFileID(message.Message)
	content, err := ms.telegramBotClientManager.HandleFileMessage(botToken, fileId)
	if err != nil {
		log.Println(err)
	}

	botMessage := entity.IntegrationsMessage{
		MessageId: message.Message.MessageID,
		Message:   ms.extractMessageText(message),
		FileIdStr: fileId,
		Type:      messageType,
		CreatedAt: primitive.DateTime(int64(message.Message.Date)),
	}

	messageResponse := &model.MessageResponse{
		Source:    string(entity.SourceTelegramBot),
		Message:   botMessage.Message,
		Type:      string(messageType),
		Content:   content,
		CreatedAt: botMessage.CreatedAt,
	}

	return botMessage, messageResponse
}

// TODO: change the content extraction method
func (ms *MessengerServiceImpl) createMessageEntitiesFromTelegramAccount(botToken string, message *tg.Message) (entity.IntegrationsMessage, *model.MessageResponse) {
	messageType, fileId := utils.GetMessageTypeAndFileIDFromTelegramAccount(message)
	//content, err := ms.telegramBotClientManager.HandleFileMessage(botToken, fileId)
	//if err != nil {
	//	log.Println(err)
	//}

	botMessage := entity.IntegrationsMessage{
		MessageId:   message.ID,
		Message:     message.Message,
		FileIdInt64: fileId,
		Type:        messageType,
		CreatedAt:   primitive.DateTime(int64(message.Date)),
	}

	messageResponse := &model.MessageResponse{
		Source:    string(entity.SourceTelegramBot),
		Message:   botMessage.Message,
		Type:      string(messageType),
		Content:   nil,
		CreatedAt: botMessage.CreatedAt,
	}

	return botMessage, messageResponse
}

func (ms *MessengerServiceImpl) extractMessageText(message *tgbotapi.Update) string {
	if message.Message.Text == "" && message.Message.Caption != "" {
		return message.Message.Caption
	}
	return message.Message.Text
}

func (ms *MessengerServiceImpl) broadcastMessageResponse(workspaceId string, messageResponse *model.MessageResponse) error {
	jsonBytes, err := json.Marshal(messageResponse)
	if err != nil {
		return err
	}

	ms.websocketService.SendToAll(workspaceId, jsonBytes)
	return nil
}

func (ms *MessengerServiceImpl) isAdmin(userRole entity.WorkspaceRole) bool {
	return userRole == entity.RoleAdmin
}

func (ms *MessengerServiceImpl) isOwner(userRole entity.WorkspaceRole) bool {
	return userRole == entity.RoleOwner
}
