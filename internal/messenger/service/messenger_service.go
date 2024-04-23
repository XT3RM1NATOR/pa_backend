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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessengerServiceImpl struct {
	messengerRepo    infrastructureInterface.MessengerRepository
	telegramClient   infrastructureInterface.TelegramClient
	websocketService _interface.WebsocketService
	config           *config.Config
}

func NewMessengerServiceImpl(cfg *config.Config, messengerRepo infrastructureInterface.MessengerRepository, websocketService _interface.WebsocketService, telegramClient infrastructureInterface.TelegramClient) _interface.MessengerService {
	return &MessengerServiceImpl{
		messengerRepo:    messengerRepo,
		telegramClient:   telegramClient,
		websocketService: websocketService,
		config:           cfg,
	}
}

func (ms *MessengerServiceImpl) RegisterBotIntegration(userId primitive.ObjectID, botToken, workspaceId string) error {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	if ms.isAdmin(workspace.Team[userId]) || ms.isOwner(workspace.Team[userId]) {
		exists, err := ms.messengerRepo.CheckBotExists(botToken)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("bot token already used")
		}

		if err := ms.telegramClient.RegisterNewBot(botToken); err != nil {
			return err
		}

		if err = ms.messengerRepo.AddTelegramIntegration(workspace.Id, botToken); err != nil {
			return err
		}

		return nil
	}

	return errors.New("user does not have the permissions")
}

func (ms *MessengerServiceImpl) HandleTelegramBotMessage(token string, message *tgbotapi.Update) error {
	workspace, err := ms.messengerRepo.FindWorkspaceByTelegramBotToken(token)
	if err != nil {
		return err
	}

	messageResponse := &model.MessageResponse{
		Source: string(entity.SourceTelegramBot),
	}

	if message.Message != nil && message.Message.Text != "" {
		botMessage := entity.IntegrationsMessage{
			MessageId: message.Message.MessageID,
			Message:   message.Message.Text,
			Type:      entity.TypeText,
			CreatedAt: primitive.DateTime(int64(message.Message.Date)),
		}

		messageResponse.Message = botMessage.Message
		messageResponse.Type = string(entity.TypeText)
		messageResponse.CreatedAt = botMessage.CreatedAt

		ticket, _ := ms.findTicketBySenderId(workspace, message.Message.From.ID)
		if ticket != nil {
			ticket.IntegrationMessages = append(ticket.IntegrationMessages, botMessage)
			messageResponse.TicketId = ticket.TicketId
		} else {
			ticketId, _ := utils.GenerateToken()
			newTicket := entity.Ticket{
				TicketId:            ticketId,
				BotToken:            token,
				SenderId:            message.Message.From.ID,
				ChatId:              message.Message.Chat.ID,
				IntegrationMessages: []entity.IntegrationsMessage{botMessage},
				Status:              entity.StatusOpen,
				Source:              entity.SourceTelegramBot,
				AssignedTo:          primitive.ObjectID{},
				CreatedAt:           primitive.DateTime(int64(message.Message.Date)),
			}

			workspace.Tickets = append(workspace.Tickets, newTicket)
			messageResponse.TicketId = newTicket.TicketId
		}

		if err := ms.messengerRepo.UpdateWorkspace(workspace); err != nil {
			return err
		}

		jsonBytes, err := json.Marshal(messageResponse)
		if err != nil {
			return err
		}

		ms.websocketService.SendToAll(workspace.WorkspaceId, jsonBytes)
	} else if message.CallbackQuery != nil {
		//
	} else {
		//
	}
	return nil
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

func (ms *MessengerServiceImpl) HandleTelegramPlatformMessage(userId primitive.ObjectID, workspaceId string, message model.MessageRequest) error {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	ticket, err := ms.findTicketByTicketId(workspace, message.TicketId)
	if err != nil {
		return err
	}

	if err := ms.telegramClient.SendTextMessage(ticket.BotToken, ticket.ChatId, message.Message); err != nil {
		return err
	}

	responseMessage := entity.ResponseMessage{
		SenderId:  userId,
		Message:   message.Message,
		Type:      entity.TypeText,
		CreatedAt: message.CreatedAt,
	}
	ticket.ResponseMessages = append(ticket.ResponseMessages, responseMessage)

	err = ms.messengerRepo.UpdateWorkspace(workspace)
	if err != nil {
		return err
	}

	jsonMessage, err := json.Marshal(responseMessage)
	if err != nil {
		return err
	}

	ms.websocketService.SendToAll(workspaceId, jsonMessage)

	return nil
}

func (ms *MessengerServiceImpl) findTicketByTicketId(workspace *entity.Workspace, ticketId string) (*entity.Ticket, error) {
	var ticketFound *entity.Ticket
	for _, ticket := range workspace.Tickets {
		if ticket.TicketId == ticketId {
			ticketFound = &ticket
			break
		}
	}

	if ticketFound == nil {
		return nil, errors.New("ticket not found")
	}
	return ticketFound, nil
}

func (ms *MessengerServiceImpl) findTicketBySenderId(workspace *entity.Workspace, senderId int) (*entity.Ticket, error) {
	var existingTicket *entity.Ticket
	for _, ticket := range workspace.Tickets {
		if ticket.SenderId == senderId && ticket.Status == entity.StatusOpen {
			existingTicket = &ticket
			break
		}
	}

	if existingTicket == nil {
		return nil, errors.New("ticket not found")
	}
	return existingTicket, nil
}

func (ms *MessengerServiceImpl) isAdmin(userRole entity.WorkspaceRole) bool {
	return userRole == entity.RoleAdmin
}

func (ms *MessengerServiceImpl) isOwner(userRole entity.WorkspaceRole) bool {
	return userRole == entity.RoleOwner
}
