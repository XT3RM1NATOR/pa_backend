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

func (ms *MessengerServiceImpl) ReassignTicketToTeam(userId primitive.ObjectID, ticketId string, workspaceId primitive.ObjectID, tgClientId int, teamName string) error {
	originalChat, err := ms.messengerRepo.FindChatByTicketID(ticketId)
	if err != nil {
		return err
	}

	var ticketToMove entity.Ticket
	for i, ticket := range originalChat.Tickets {
		if ticket.TicketId == ticketId {
			ticketToMove = ticket
			originalChat.Tickets = append(originalChat.Tickets[:i], originalChat.Tickets[i+1:]...)
			break
		}
	}

	if &ticketToMove == nil {
		return errors.New("ticket not found")
	}

	if len(originalChat.Tickets) == 0 {
		if err := ms.messengerRepo.DeleteChat(originalChat.Id); err != nil {
			return err
		}
	} else {
		if err := ms.messengerRepo.UpdateChat(originalChat); err != nil {
			return err
		}
	}

	assigneeId, err := ms.getAssigneeIdByTeam(workspaceId, teamName)
	if err != nil {
		return err
	}
	newChat, err := ms.messengerRepo.FindOrCreateChatByUserId(workspaceID, assigneeId)
	if err != nil {
		return err
	}

	// Add the ticket to the new chat
	newChat.Tickets = append(newChat.Tickets, ticketToMove)
	return ms.messengerRepo.UpdateChat(newChat)
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

func (ms *MessengerServiceImpl) getAssigneeIdByTeam(workspaceId, teamName string) (primitive.ObjectID, error) {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return primitive.NilObjectID, err
	}

	if team, exists := workspace.InternalTeams[teamName]; exists {
		return ms.findLeastBusyMember(team)
	}

	return primitive.NilObjectID, errors.New("specified team does not exist in the workspace")
}

func (ms *MessengerServiceImpl) findLeastBusyMember(team map[primitive.ObjectID]entity.UserStatus) (primitive.ObjectID, error) {
	var leastBusyMember primitive.ObjectID
	minTickets := int(^uint(0) >> 1)

	for memberId := range team {
		activeTicketsCount, err := ms.messengerRepo.CountActiveTickets(memberId)
		if err != nil {
			continue
		}
		if activeTicketsCount < minTickets {
			minTickets = activeTicketsCount
			leastBusyMember = memberId
		}
	}

	if leastBusyMember.IsZero() {
		return primitive.NilObjectID, errors.New("no suitable team member found")
	}

	return leastBusyMember, nil
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
