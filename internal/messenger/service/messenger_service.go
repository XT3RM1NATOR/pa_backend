package service

import (
	"encoding/json"
	"errors"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/messenger/delivery/model"
	"github.com/Point-AI/backend/internal/messenger/domain/entity"
	_interface "github.com/Point-AI/backend/internal/messenger/domain/interface"
	"github.com/Point-AI/backend/internal/messenger/infrastructure/client"
	infrastructureInterface "github.com/Point-AI/backend/internal/messenger/service/interface"
	"github.com/Point-AI/backend/utils"
	"github.com/celestix/gotgproto/ext"
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

		if err := ms.telegramBotClientManager.RegisterNewBot(botToken); err != nil {
			return err
		}

		telegramBotIntegration := &entity.TelegramBotIntegration{
			BotToken: botToken,
			IsActive: true,
		}
		workspace.Integrations.TelegramBot = telegramBotIntegration

		if err = ms.messengerRepo.UpdateWorkspace(workspace); err != nil {
			return err
		}

		return nil
	}

	return errors.New("user does not have the permissions")
}

func (ms *MessengerServiceImpl) ReassignTicketToTeam(userId primitive.ObjectID, ticketId, workspaceId, teamName string) error {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	if _, exists := workspace.Team[userId]; exists {
		for _, ticket := range workspace.Tickets {
			if ticket.TicketId == ticketId && ticket.Status != entity.StatusClosed {
				assigneeId, err := ms.getAssigneeId(workspace, teamName)
				if err != nil {
					return err
				}

				ticket.AssignedTo = assigneeId
				if err := ms.messengerRepo.UpdateWorkspace(workspace); err != nil {
					return err
				}
				return nil
			}
		}
		return errors.New("no open or pending tickets with this id")
	}

	return errors.New("user is not from the workspace")
}

func (ms *MessengerServiceImpl) ReassignTicketToMember(userId primitive.ObjectID, ticketId, workspaceId, userEmail string) error {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	if _, exists := workspace.Team[userId]; exists {
		user, err := ms.messengerRepo.FindUserByEmail(userEmail)
		if err != nil {
			return err
		}

		for _, ticket := range workspace.Tickets {
			if ticket.TicketId == ticketId && ticket.Status != entity.StatusClosed {
				ticket.AssignedTo = user
				if err := ms.messengerRepo.UpdateWorkspace(workspace); err != nil {
					return err
				}
				return nil
			}
		}
		return errors.New("no open or pending tickets with this id")
	}

	return errors.New("user is not from the workspace")
}

func (ms *MessengerServiceImpl) HandleTelegramAccountMessage(ctx *ext.Context, update *ext.Update) error {
	workspace, err := ms.messengerRepo.FindWorkspaceByPhoneNumber(ctx.Self.Phone)
	if err != nil {
		return err
	}

	tgMessage, messageResponse := ms.createMessageEntitiesFromTelegramAccount("", update.EffectiveMessage.Message)

	if err := ms.processTicketHandlingTelegramAccount(workspace, update.EffectiveMessage.Message, tgMessage, messageResponse); err != nil {
		return err
	}

	if err := ms.messengerRepo.UpdateWorkspace(workspace); err != nil {
		return err
	}

	return ms.broadcastMessageResponse(workspace.WorkspaceId, messageResponse)
}

func (ms *MessengerServiceImpl) HandleTelegramPlatformMessageToBot(request model.MessageRequest, workspaceId string, userId primitive.ObjectID) error {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	user, err := ms.messengerRepo.GetUserById(userId)
	if err != nil {
		return err
	}

	if _, ok := workspace.Team[userId]; !ok {
		return errors.New("user does not belong to the workspace")
	}

	if workspace.Integrations.TelegramBot == nil || !workspace.Integrations.TelegramBot.IsActive {
		return errors.New("no active telegram bot integration found")
	}

	chatID := extractChatIDFromTicket(workspace, request.TicketId)
	if chatID == 0 {
		return errors.New("invalid ticket ID or chat ID not found")
	}
	err = ms.telegramBotClientManager.SendTextMessage(workspace.Integrations.TelegramBot.BotToken, chatID, request.Message)
	if err != nil {
		return err
	}

	messageResponse := createMessageResponseFromRequest(request, user.Email)

	return ms.broadcastMessageResponse(workspace.WorkspaceId, messageResponse)
}

func extractChatIDFromTicket(workspace *entity.Workspace, ticketId string) int64 {
	for _, ticket := range workspace.Tickets {
		if ticket.TicketId == ticketId {
			return ticket.ChatId
		}
	}
	return 0
}

func createMessageResponseFromRequest(request model.MessageRequest, email string) *model.MessageResponse {
	return &model.MessageResponse{
		TicketId:  request.TicketId,
		Message:   request.Message,
		Type:      request.Type,
		Source:    "platform",
		Username:  email,
		CreatedAt: *request.CreatedAt,
	}
}

func (ms *MessengerServiceImpl) HandleTelegramBotMessage(token string, message *tgbotapi.Update) error {
	workspace, err := ms.messengerRepo.FindWorkspaceByTelegramBotToken(token)
	if err != nil {
		return err
	}

	botMessage, messageResponse := ms.createMessageEntitiesFromTelegramBot(token, message)

	if err := ms.processTicketHandling(workspace, token, message, botMessage, messageResponse); err != nil {
		return err
	}

	if err := ms.messengerRepo.UpdateWorkspace(workspace); err != nil {
		return err
	}

	return ms.broadcastMessageResponse(workspace.WorkspaceId, messageResponse)
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

func (ms *MessengerServiceImpl) HandleTelegramClientAuth(userId primitive.ObjectID, workspaceId, action, value string) (client.AuthStatus, error) {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return "", err
	}
	if workspace.Integrations.Telegram != nil {
		return "", errors.New("telegram integration already exists")
	}

	if ms.isAdmin(workspace.Team[userId]) || ms.isOwner(workspace.Team[userId]) {
		switch action {
		case "phone":
			err := ms.telegramClientManager.CreateClient(value, workspaceId, ms.HandleTelegramAccountMessage)
			if err != nil {
				return "", err
			}

			if authConversator, ok := ms.telegramClientManager.GetAuthConversator(workspaceId); ok {
				authConversator.ReceivePhone(value)
				return authConversator.Status, nil
			}
			return "", errors.New("error adding the phone number")
		case "code":
			if authConversator, ok := ms.telegramClientManager.GetAuthConversator(workspaceId); ok {
				authConversator.ReceiveCode(value)
				if authConversator.Status != client.StatusPassword {
					if telegramClient, ok := ms.telegramClientManager.GetClient(workspaceId); ok {
						session, err := telegramClient.ExportStringSession()
						if err != nil {
							return "", err
						}

						workspace.Integrations.Telegram.Session = session
						workspace.Integrations.Telegram.PhoneNumber = telegramClient.Self.Phone
						workspace.Integrations.Telegram.IsActive = true
						err = ms.messengerRepo.UpdateWorkspace(workspace)
						if err != nil {
							return "", err
						}
					}
				}
				return authConversator.Status, nil
			}
			return "", errors.New("error validating the code")
		case "passwd":
			if authConversator, ok := ms.telegramClientManager.GetAuthConversator(workspaceId); ok {
				authConversator.ReceivePasswd(value)
				if telegramClient, ok := ms.telegramClientManager.GetClient(workspaceId); ok {
					session, err := telegramClient.ExportStringSession()
					if err != nil {
						return "", err
					}

					workspace.Integrations.Telegram.Session = session
					workspace.Integrations.Telegram.PhoneNumber = telegramClient.Self.Phone
					workspace.Integrations.Telegram.IsActive = true
					err = ms.messengerRepo.UpdateWorkspace(workspace)
					if err != nil {
						return "", err
					}
				}
				return authConversator.Status, nil
			}
			return "", errors.New("error validating the password")
		}
	}

	return "", errors.New("user does not have the permissions")
}

func (ms *MessengerServiceImpl) UpdateTicketStatus(userId primitive.ObjectID, ticketId, workspaceId, status string) error {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	var ticketStatus entity.TicketStatus
	switch entity.TicketStatus(status) {
	case entity.StatusOpen, entity.StatusPending:
		ticketStatus = entity.TicketStatus(status)
	default:
		return errors.New("invalid status")
	}

	if _, exists := workspace.Team[userId]; exists {
		for _, ticket := range workspace.Tickets {
			if ticket.TicketId == ticketId && ticket.Status != entity.StatusClosed {
				ticket.Status = ticketStatus
				if err := ms.messengerRepo.UpdateWorkspace(workspace); err != nil {
					return err
				}
				return nil
			}
		}
		return errors.New("no open or pending tickets with this id")
	}

	return errors.New("user does not have the permissions")
}

func (ms *MessengerServiceImpl) SetUpTelegramClients() error {
	workspaces, err := ms.messengerRepo.GetAllWorkspaceRepositories()
	if err != nil {
		return err
	}
	for _, workspace := range workspaces {
		if workspace.Integrations.Telegram.Session != "" {
			if err := ms.telegramClientManager.CreateClientBySession(workspace.Integrations.Telegram.Session, workspace.Integrations.Telegram.PhoneNumber, workspace.WorkspaceId, ms.HandleTelegramAccountMessage); err != nil {
				return err
			}
		}
	}
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

func (ms *MessengerServiceImpl) addNewTicketToWorkspace(token string, message *tgbotapi.Update, workspace *entity.Workspace, botMessage *entity.IntegrationsMessage) error {
	ticketId, _ := utils.GenerateToken()
	newTicket := entity.Ticket{
		TicketId:            ticketId,
		BotToken:            token,
		SenderId:            message.Message.From.ID,
		ChatId:              message.Message.Chat.ID,
		IntegrationMessages: []entity.IntegrationsMessage{*botMessage},
		Status:              entity.StatusPending,
		Source:              entity.SourceTelegramBot,
		SenderUsername:      message.Message.From.UserName,
		AssignedTo:          primitive.NilObjectID,
		CreatedAt:           primitive.DateTime(int64(message.Message.Date)),
	}

	assigneeId, err := ms.getAssigneeId(workspace, workspace.FirstTeam)
	if err == nil {
		newTicket.AssignedTo = assigneeId
	}

	workspace.Tickets = append(workspace.Tickets, newTicket)

	return nil
}

func (ms *MessengerServiceImpl) addNewTicketToWorkspaceTelegramAccount(workspace *entity.Workspace, tgMessage *entity.IntegrationsMessage, senderId int, chatId int64) error {
	ticketId, _ := utils.GenerateToken()
	newTicket := entity.Ticket{
		TicketId:            ticketId,
		SenderId:            senderId,
		ChatId:              chatId,
		IntegrationMessages: []entity.IntegrationsMessage{*tgMessage},
		Status:              entity.StatusPending,
		Source:              entity.SourceTelegramBot,
		SenderUsername:      "Michael Bay",
		AssignedTo:          primitive.NilObjectID,
		CreatedAt:           tgMessage.CreatedAt,
	}

	assigneeId, err := ms.getAssigneeId(workspace, workspace.FirstTeam)
	if err == nil {
		newTicket.AssignedTo = assigneeId
	}

	workspace.Tickets = append(workspace.Tickets, newTicket)

	return nil
}

func (ms *MessengerServiceImpl) getAssigneeId(workspace *entity.Workspace, teamName string) (primitive.ObjectID, error) {
	statusPriority := []entity.UserStatus{entity.StatusAvailable, entity.StatusBusy, entity.StatusOffline}

	var assignedCount map[primitive.ObjectID]int
	found := false

	for _, status := range statusPriority {
		assignedCount = make(map[primitive.ObjectID]int)
		for _, ticket := range workspace.Tickets {
			if ticket.AssignedTo != primitive.NilObjectID && workspace.InternalTeams[teamName][ticket.AssignedTo] == status {
				assignedCount[ticket.AssignedTo]++
				found = true
			}
		}
		if found {
			break
		}
	}

	if !found {
		log.Println("no chat members yet")
		return primitive.NilObjectID, errors.New("no tickets are assigned")
	}

	return ms.findMinAssignee(assignedCount)
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

func (ms *MessengerServiceImpl) processTicketHandling(workspace *entity.Workspace, botToken string, message *tgbotapi.Update, botMessage entity.IntegrationsMessage, messageResponse *model.MessageResponse) error {
	ticket, err := ms.findTicketBySenderId(workspace, message.Message.From.ID)
	if err != nil && !errors.Is(err, errors.New("ticket not found")) {
		return err
	}

	if ticket != nil && ticket.Status != entity.StatusClosed {
		ticket.IntegrationMessages = append(ticket.IntegrationMessages, botMessage)
		messageResponse.TicketId = ticket.TicketId

		if message.Message != nil && message.Message.From != nil {
			messageResponse.Username = message.Message.From.UserName
		}

	} else {
		if err := ms.addNewTicketToWorkspace(botToken, message, workspace, &botMessage); err != nil {
			return err
		}
	}
	return nil
}

func (ms *MessengerServiceImpl) processTicketHandlingTelegramAccount(workspace *entity.Workspace, message *tg.Message, tgMessage entity.IntegrationsMessage, messageResponse *model.MessageResponse) error {
	switch v := message.FromID.(type) {
	case *tg.PeerUser:
		ticket, err := ms.findTicketBySenderId(workspace, int(v.UserID))
		if err != nil && !errors.Is(err, errors.New("ticket not found")) {
			return err
		}

		if ticket != nil && ticket.Status != entity.StatusClosed {
			ticket.IntegrationMessages = append(ticket.IntegrationMessages, tgMessage)
			messageResponse.TicketId = ticket.TicketId

			if message.Message != "" {
				messageResponse.Username = "Michael Bay"
			}

		} else {
			if err := ms.addNewTicketToWorkspaceTelegramAccount(workspace, &tgMessage, int(v.UserID), -1); err != nil {
				return err
			}
		}
		return nil
	}
	return errors.New("wrong message resource")
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
