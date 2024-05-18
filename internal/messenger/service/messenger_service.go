package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/messenger/delivery/model"
	"github.com/Point-AI/backend/internal/messenger/domain/entity"
	_interface "github.com/Point-AI/backend/internal/messenger/domain/interface"
	infrastructureInterface "github.com/Point-AI/backend/internal/messenger/service/interface"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"sort"
	"time"
)

type MessengerServiceImpl struct {
	messengerRepo    infrastructureInterface.MessengerRepository
	websocketService _interface.WebsocketService
	fileService      _interface.FileService
	config           *config.Config
}

func NewMessengerServiceImpl(cfg *config.Config, messengerRepo infrastructureInterface.MessengerRepository, websocketService _interface.WebsocketService, fileService _interface.FileService) _interface.MessengerService {
	return &MessengerServiceImpl{
		messengerRepo:    messengerRepo,
		websocketService: websocketService,
		fileService:      fileService,
		config:           cfg,
	}
}

func (ms *MessengerServiceImpl) ReassignTicketToTeam(userId primitive.ObjectID, chatId string, ticketId, workspaceId, teamId string) error {
	session, err := ms.messengerRepo.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	err = session.StartTransaction()
	if err != nil {
		return err
	}

	err = mongo.WithSession(context.Background(), session, func(sc mongo.SessionContext) error {
		originalChat, err := ms.messengerRepo.FindChatByTicketId(sc, ticketId)
		if err != nil {
			return err
		}

		ticketToMove, err := ms.findTicketInChat(originalChat, ticketId)
		if err != nil {
			return err
		}

		if len(originalChat.Tickets) == 0 {
			if err := ms.messengerRepo.DeleteChat(sc, originalChat.Id); err != nil {
				return err
			}
		} else {
			if err := ms.messengerRepo.UpdateChat(sc, originalChat); err != nil {
				return err
			}
		}

		workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(sc, workspaceId)
		if err != nil {
			return err
		}

		if err = ms.ValidateUserInWorkspace(userId, workspace); err != nil {
			return err
		}

		assigneeId, err := ms.getAssigneeIdByTeam(workspace, teamId)
		if err != nil {
			return err
		}

		chat, err := ms.messengerRepo.FindChatByUserId(sc, originalChat.TgClientId, workspace.Id, assigneeId)
		if err != nil {
			return err
		} else if chat == nil {
			newChat := ms.createChat(originalChat.TgChatId, originalChat.TgClientId, originalChat.Source, *ticketToMove, workspace.Id, originalChat.UserId, originalChat.TeamId, originalChat.IsImported, originalChat.LastMessage, originalChat.Name, originalChat.Company, originalChat.ClientEmail, originalChat.ClientPhone, originalChat.Address)
			return ms.messengerRepo.InsertNewChat(sc, newChat)
		} else if chat != nil {
			chat.Tickets = append(chat.Tickets, *ticketToMove)
			return ms.messengerRepo.UpdateChat(sc, chat)
		}

		return nil
	})

	if err != nil {
		_ = session.AbortTransaction(context.Background())
		return err
	}
	return session.CommitTransaction(context.Background())
}

func (ms *MessengerServiceImpl) ReassignTicketToUser(userId primitive.ObjectID, chatId string, ticketId, workspaceId, email string) error {
	session, err := ms.messengerRepo.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	err = session.StartTransaction()
	if err != nil {
		return err
	}

	err = mongo.WithSession(context.Background(), session, func(sc mongo.SessionContext) error {
		originalChat, err := ms.messengerRepo.FindChatByTicketId(sc, ticketId)
		if err != nil {
			return err
		}

		ticketToMove, err := ms.findTicketInChat(originalChat, ticketId)
		if err != nil {
			return err
		}

		if len(originalChat.Tickets) == 0 {
			if err := ms.messengerRepo.DeleteChat(sc, originalChat.Id); err != nil {
				return err
			}
		} else {
			if err := ms.messengerRepo.UpdateChat(sc, originalChat); err != nil {
				return err
			}
		}

		workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(sc, workspaceId)
		if err != nil {
			return err
		}

		if err = ms.ValidateUserInWorkspace(userId, workspace); err != nil {
			return err
		}

		reassignUserId, err := ms.messengerRepo.FindUserByEmail(sc, email)
		if err != nil {
			return err
		}

		chat, err := ms.messengerRepo.FindChatByUserId(sc, originalChat.TgClientId, workspace.Id, reassignUserId)
		if err != nil {
			return err
		} else if chat == nil && err == nil {
			newChat := ms.createChat(originalChat.TgChatId, originalChat.TgClientId, originalChat.Source, *ticketToMove, workspace.Id, originalChat.UserId, originalChat.TeamId, originalChat.IsImported, originalChat.LastMessage, originalChat.Name, originalChat.Company, originalChat.ClientEmail, originalChat.ClientPhone, originalChat.Address)

			return ms.messengerRepo.InsertNewChat(sc, newChat)
		} else if chat != nil {
			chat.Tickets = append(chat.Tickets, *ticketToMove)
			return ms.messengerRepo.UpdateChat(sc, chat)
		}

		return nil
	})

	if err != nil {
		_ = session.AbortTransaction(context.Background())
		return err
	}
	err = session.CommitTransaction(context.Background())
	return err
}

func (ms *MessengerServiceImpl) GetAllChats(userId primitive.ObjectID, workspaceId string) ([]model.ChatResponse, error) {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(nil, workspaceId)
	if err != nil {
		return nil, err
	}

	if _, ok := workspace.Team[userId]; !ok {
		return nil, errors.New("unauthorised")
	}
	chats, err := ms.messengerRepo.FindLatestChatsByWorkspaceId(workspace.Id, 50)
	if err != nil {
		return nil, err
	}

	var responseChats []model.ChatResponse

	for _, chat := range chats {
		if chat.IsImported {
			go ms.updateWallpaper(workspaceId, chat.ChatId, chat.TgChatId)
		}
		totalTickets, averageSolutionTime, averageNumberOfMessagesPerTicket := ms.extractTicketData(chat.Tickets)

		logo, _ := ms.fileService.LoadFile("chat." + chat.ChatId)
		messageResponse := ms.createMessageResponse(nil, chat.LastMessage.CreatedAt, userId == chat.LastMessage.SenderId, chat.LastMessage.From, "", workspaceId, chat.Tickets[0].TicketId, chat.ChatId, chat.LastMessage.MessageId, chat.LastMessage.Message, string(chat.LastMessage.Type))
		responseChats = append(responseChats, *ms.createChatResponse(workspace.WorkspaceId, chat.ChatId, chat.TgClientId, chat.TgChatId, chat.Tags, *messageResponse, string(entity.SourceTelegram), chat.IsImported, chat.CreatedAt, chat.Name, logo, nil, string(chat.Language), chat.Company, chat.ClientEmail, chat.ClientPhone, chat.Address, totalTickets, averageSolutionTime, averageNumberOfMessagesPerTicket))
	}

	return responseChats, nil
}

func (ms *MessengerServiceImpl) GetAllUnassignedChats(userId primitive.ObjectID, workspaceId string) ([]model.ChatResponse, error) {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(nil, workspaceId)
	if err != nil {
		return nil, err
	}

	if _, ok := workspace.Team[userId]; !ok {
		return nil, errors.New("unauthorised")
	}
	chats, err := ms.messengerRepo.FindLatestUnassignedChatsByWorkspaceId(workspace.Id, 50)
	if err != nil {
		return nil, err
	}

	var responseChats []model.ChatResponse

	for _, chat := range chats {
		if chat.IsImported {

		}
		totalTickets, averageSolutionTime, averageNumberOfMessagesPerTicket := ms.extractTicketData(chat.Tickets)

		logo, _ := ms.fileService.LoadFile("chat." + chat.ChatId)
		messageResponse := ms.createMessageResponse(nil, chat.LastMessage.CreatedAt, false, chat.LastMessage.From, "", workspaceId, chat.Tickets[0].TicketId, chat.ChatId, chat.LastMessage.MessageId, chat.LastMessage.Message, string(chat.LastMessage.Type))
		responseChats = append(responseChats, *ms.createChatResponse(workspace.WorkspaceId, chat.ChatId, chat.TgClientId, chat.TgChatId, chat.Tags, *messageResponse, string(entity.SourceTelegram), chat.IsImported, chat.CreatedAt, chat.Name, logo, nil, string(chat.Language), chat.Company, chat.ClientEmail, chat.ClientPhone, chat.Address, totalTickets, averageSolutionTime, averageNumberOfMessagesPerTicket))
	}

	return responseChats, nil
}

func (ms *MessengerServiceImpl) GetAllPrimaryChats(userId primitive.ObjectID, workspaceId string) ([]model.ChatResponse, error) {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(nil, workspaceId)
	if err != nil {
		return nil, err
	}

	if _, ok := workspace.Team[userId]; !ok {
		return nil, errors.New("unauthorised")
	}
	chats, err := ms.messengerRepo.FindLatestChatsByWorkspaceIdAndUserId(workspace.Id, userId, 50)
	if err != nil {
		return nil, err
	}

	var responseChats []model.ChatResponse

	for _, chat := range chats {
		if chat.IsImported {

		}
		totalTickets, averageSolutionTime, averageNumberOfMessagesPerTicket := ms.extractTicketData(chat.Tickets)

		logo, _ := ms.fileService.LoadFile("chat." + chat.ChatId)
		messageResponse := ms.createMessageResponse(nil, chat.LastMessage.CreatedAt, true, chat.LastMessage.From, "", workspaceId, chat.Tickets[0].TicketId, chat.ChatId, chat.LastMessage.MessageId, chat.LastMessage.Message, string(chat.LastMessage.Type))
		responseChats = append(responseChats, *ms.createChatResponse(workspace.WorkspaceId, chat.ChatId, chat.TgClientId, chat.TgChatId, chat.Tags, *messageResponse, string(entity.SourceTelegram), chat.IsImported, chat.CreatedAt, chat.Name, logo, nil, string(chat.Language), chat.Company, chat.ClientEmail, chat.ClientPhone, chat.Address, totalTickets, averageSolutionTime, averageNumberOfMessagesPerTicket))
	}

	return responseChats, nil
}

func (ms *MessengerServiceImpl) UpdateTicketStatus(userId primitive.ObjectID, ticketId, workspaceId, status string) error {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(nil, workspaceId)
	if err != nil {
		return err
	}

	if _, ok := workspace.Team[userId]; !ok {
		return errors.New("unauthorised")
	}

	chat, err := ms.messengerRepo.FindChatByTicketId(nil, ticketId)
	if err != nil {
		return err
	}

	fmtdStatus, err := ms.validateTicketStatus(status)
	if err != nil {
		return err
	}

	found := false
	for i, ticket := range chat.Tickets {
		if ticket.TicketId == ticketId {
			chat.Tickets[i].Status = fmtdStatus
			found = true
			break
		}
	}
	if !found {
		return errors.New("ticket not found")
	}

	return ms.messengerRepo.UpdateChat(nil, chat)
}

func (ms *MessengerServiceImpl) ValidateUserInWorkspace(userId primitive.ObjectID, workspace *entity.Workspace) error {
	if _, exists := workspace.Team[userId]; exists {
		return nil
	}

	return errors.New("user does not have the permissions")
}

func (ms *MessengerServiceImpl) ImportTelegramChats(workspaceId string, chats []model.TelegramChat) error {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(nil, workspaceId)
	if err != nil {
		return err
	}

	for _, chat := range chats {
		message := ms.createMessage(primitive.ObjectID{}, chat.LastMessage.Id, chat.LastMessage.Text, chat.Title, entity.TypeText, time.Now())
		ticket := ms.createTicket([]entity.Note{}, []entity.Message{*message}, time.Now())
		newChat := ms.createChat(int(chat.Id), int(chat.LastMessage.SenderId), entity.SourceTelegram, *ticket, workspace.Id, primitive.NilObjectID, primitive.NilObjectID, true, *message, chat.Name, "", "", "", "")
		go ms.updateWallpaper(workspaceId, newChat.ChatId, int(chat.Id))

		err := ms.messengerRepo.InsertNewChat(nil, newChat)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ms *MessengerServiceImpl) ValidateUserInWorkspaceById(userId primitive.ObjectID, workspaceId string) error {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(nil, workspaceId)
	if err != nil {
		return err
	}
	if _, exists := workspace.Team[userId]; exists {
		return nil
	}

	return errors.New("user does not have the permissions")
}

func (ms *MessengerServiceImpl) HandleChatWS(userId primitive.ObjectID, workspaceId string, w http.ResponseWriter, r *http.Request) error {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(nil, workspaceId)
	if err != nil {
		return err
	}
	if _, exists := workspace.Team[userId]; !exists {
		return errors.New("user does not have the permissions")
	}

	if err = ms.ValidateUserInWorkspaceById(userId, workspaceId); err != nil {
		return err
	}

	ws, err := ms.websocketService.UpgradeConnection(w, r, workspaceId, userId)
	if err != nil {
		return err
	}

	go func() {
		defer ms.websocketService.RemoveConnection(workspaceId, userId)
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				break
			}

			var receivedMessage model.MessageRequest
			if err := json.Unmarshal(message, &receivedMessage); err != nil {
				continue
			}
			log.Print(receivedMessage)

			if err = ms.HandleMessage(userId, workspaceId, receivedMessage.TicketId, receivedMessage.ChatId, receivedMessage.Type, receivedMessage.Message); err != nil {
				log.Println("i am in the error", err)
			}
		}
	}()

	return nil
}

func (ms *MessengerServiceImpl) HandleMessage(userId primitive.ObjectID, workspaceId, ticketId, chatId, messageType, message string) error {
	chat, err := ms.messengerRepo.FindChatByChatId(chatId)
	if err != nil {
		return err
	}

	user, err := ms.messengerRepo.GetUserById(userId)
	if err != nil {
		return err
	}

	switch messageType {
	case "chat_note":
		note := ms.createNote(userId, message)
		chat.Notes = append(chat.Notes, *note)

		if err = ms.messengerRepo.UpdateChat(nil, chat); err != nil {
			return err
		}

		res, err := json.Marshal(ms.createMessageResponse(nil, note.CreatedAt, true, user.FullName, "", workspaceId, ticketId, chatId, note.NoteId, message, messageType))
		if err != nil {
			return err
		}
		ms.websocketService.SendToOne(res, workspaceId, userId)

		res, err = json.Marshal(ms.createMessageResponse(nil, note.CreatedAt, false, user.FullName, "", workspaceId, ticketId, chatId, note.NoteId, message, messageType))
		if err != nil {
			return err
		}
		ms.websocketService.SendToAllButOne(workspaceId, res, userId)

		return nil
	case "ticket_note":
		ticket, err := ms.findTicketInChat(chat, ticketId)
		if err != nil {
			return err
		}

		note := ms.createNote(userId, message)
		ticket.Notes = append(ticket.Notes, *note)
		if err = ms.messengerRepo.UpdateChat(nil, chat); err != nil {
			return err
		}

		res, err := json.Marshal(ms.createMessageResponse(nil, note.CreatedAt, true, user.FullName, "", workspaceId, ticketId, chatId, note.NoteId, message, messageType))
		if err != nil {
			return err
		}
		ms.websocketService.SendToOne(res, workspaceId, userId)

		res, err = json.Marshal(ms.createMessageResponse(nil, note.CreatedAt, false, user.FullName, "", workspaceId, ticketId, chatId, note.NoteId, message, messageType))
		if err != nil {
			return err
		}
		ms.websocketService.SendToAllButOne(workspaceId, res, userId)

		return nil
	default:
		return errors.New("unknown message type")
	}
}

func (ms *MessengerServiceImpl) UpdateChatInfo(userId primitive.ObjectID, chatId string, tags []string, workspaceId, language string, address, company, clientEmail, clientPhone string) error {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(nil, workspaceId)
	if err != nil {
		return err
	}

	if err = ms.ValidateUserInWorkspace(userId, workspace); err != nil {
		return err
	}

	chat, err := ms.messengerRepo.FindChatByWorkspaceIdAndChatId(workspace.Id, chatId)
	if err != nil {
		return err
	}
	chat.Tags = tags

	if language != "" {
		switch entity.ChatLanguage(language) {
		case entity.English:
			chat.Language = entity.English
		case entity.Russian:
			chat.Language = entity.Russian
		case entity.Uzbek:
			chat.Language = entity.Uzbek
		}
	}
	if address != "" {
		chat.Address = address
	}
	if company != "" {
		chat.Company = company
	}
	if clientEmail != "" {
		chat.ClientEmail = clientEmail
	}
	if clientPhone != "" {
		chat.ClientPhone = clientPhone
	}

	for _, tag := range tags {
		var found bool
		for _, workspaceTag := range workspace.Tags {
			if tag == workspaceTag {
				found = true
			}
		}
		if !found {
			workspace.Tags = append(workspace.Tags, tag)
		}
	}

	if err := ms.messengerRepo.UpdateWorkspace(workspace); err != nil {
		return err
	}

	return ms.messengerRepo.UpdateChat(nil, chat)
}

func (ms *MessengerServiceImpl) GetChat(userId primitive.ObjectID, workspaceId, chatId string) (model.ChatResponse, error) {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(nil, workspaceId)
	if err != nil {
		return model.ChatResponse{}, err
	}

	if err = ms.ValidateUserInWorkspace(userId, workspace); err != nil {
		return model.ChatResponse{}, err
	}

	chat, err := ms.messengerRepo.FindChatByWorkspaceIdAndChatId(workspace.Id, chatId)
	if err != nil {
		return model.ChatResponse{}, err
	}

	var notes []model.MessageResponse
	for _, note := range chat.Notes {
		user, _ := ms.messengerRepo.FindUserById(note.UserId)
		notes = append(notes, *ms.createMessageResponse(nil, note.CreatedAt, note.UserId == userId, user.FullName, "", workspaceId, "", chatId, note.NoteId, note.Text, string(entity.TypeChatNote)))
	}
	totalTickets, averageSolutionTime, averageNumberOfMessagesPerTicket := ms.extractTicketData(chat.Tickets)

	logo, _ := ms.fileService.LoadFile("chat." + chatId)
	responseMessage := ms.createMessageResponse(nil, chat.LastMessage.CreatedAt, chat.UserId == userId, chat.LastMessage.From, "", workspaceId, "", chat.ChatId, chat.LastMessage.MessageId, chat.LastMessage.Message, string(entity.TypeText))
	responseChat := ms.createChatResponse(workspaceId, chatId, chat.TgClientId, chat.TgChatId, chat.Tags, *responseMessage, string(chat.Source), chat.IsImported, chat.CreatedAt, chat.Name, logo, notes, string(chat.Language), chat.Company, chat.ClientEmail, chat.ClientPhone, chat.Address, totalTickets, averageSolutionTime, averageNumberOfMessagesPerTicket)

	return *responseChat, nil
}

func (ms *MessengerServiceImpl) GetMessages(userId primitive.ObjectID, workspaceId, chatId string, lastMessageDate time.Time) ([]model.MessageResponse, error) {
	//workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(nil, workspaceId)
	//if err != nil {
	//	return model.ChatResponse{}, err
	//}

	//if err = ms.ValidateUserInWorkspace(userId, workspace); err != nil {
	//	return model.ChatResponse{}, err
	//}
	//
	//chat, err := ms.messengerRepo.FindChatByWorkspaceIdAndChatId(workspace.Id, chatId)
	//if err != nil {
	//	return model.ChatResponse{}, err
	//}
	//
	//responseMessage := ms.createMessageResponse(nil, chat.LastMessage.CreatedAt, chat.UserId == userId, chat.LastMessage.From, "", workspaceId, "", chat.ChatId, chat.LastMessage.MessageId, chat.LastMessage.Message, string(entity.TypeText))
	//responseChat := ms.createChatResponse(workspaceId, chatId, chat.TgClientId, chat.TgChatId, chat.Tags, *responseMessage, string(chat.Source), chat.IsImported, chat.CreatedAt, chat.Name)

	return nil, nil
}

func (ms *MessengerServiceImpl) GetChatsByFolder(userId primitive.ObjectID, workspaceId, folderName string) ([]model.ChatResponse, error) {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(nil, workspaceId)
	if err != nil {
		return nil, err
	}

	if _, exists := workspace.Team[userId]; !exists {
		return nil, errors.New("unauthorised")
	}

	chats, err := ms.messengerRepo.FindLatestChatsByWorkspaceIdAndAllTags(workspace.Id, workspace.Folders[folderName], 50)
	if err != nil {
		return nil, err
	}

	var responseChats []model.ChatResponse

	for _, chat := range chats {
		if chat.IsImported {
			//
		}
		totalTickets, averageSolutionTime, averageNumberOfMessagesPerTicket := ms.extractTicketData(chat.Tickets)

		logo, _ := ms.fileService.LoadFile("chat." + chat.ChatId)
		messageResponse := ms.createMessageResponse(nil, chat.LastMessage.CreatedAt, userId == chat.LastMessage.SenderId, chat.LastMessage.From, "", workspaceId, chat.Tickets[0].TicketId, chat.ChatId, chat.LastMessage.MessageId, chat.LastMessage.Message, string(chat.LastMessage.Type))
		responseChats = append(responseChats, *ms.createChatResponse(workspace.WorkspaceId, chat.ChatId, chat.TgClientId, chat.TgChatId, chat.Tags, *messageResponse, string(entity.SourceTelegram), chat.IsImported, chat.CreatedAt, chat.Name, logo, nil, string(chat.Language), chat.Company, chat.ClientEmail, chat.ClientPhone, chat.Address, totalTickets, averageSolutionTime, averageNumberOfMessagesPerTicket))
	}

	return responseChats, nil
}

func (ms *MessengerServiceImpl) GetAllTags(userId primitive.ObjectID, workspaceId string) ([]string, error) {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(nil, workspaceId)
	if err != nil {
		return nil, err
	}

	if !ms.isAdmin(workspace.Team[userId]) && !ms.isOwner(workspace.Team[userId]) {
		return nil, errors.New("unauthorised")
	}

	return workspace.Tags, nil
}

func (ms *MessengerServiceImpl) DeleteMessage(userId primitive.ObjectID, messageType, workspaceId, ticketId, messageId, chatId string) error {
	workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(nil, workspaceId)
	if err != nil {
		return err
	}
	if _, exists := workspace.Team[userId]; !exists {
		return errors.New("unauthorized")
	}

	chat, err := ms.messengerRepo.FindChatByWorkspaceIdAndChatId(workspace.Id, chatId)
	if err != nil {
		return err
	}

	switch messageType {
	case "chat_note":
		index, err := ms.findNoteIndexInChat(chat, messageId)
		if err != nil {
			return err
		}

		chat.Notes = append(chat.Notes[:index], chat.Notes[index+1:]...)
		if err = ms.messengerRepo.UpdateChat(nil, chat); err != nil {
			return err
		}

		res, err := json.Marshal(ms.createMessageResponse(nil, time.Time{}, false, "", "delete", workspaceId, "", chatId, messageId, "", messageType))
		if err != nil {
			return err
		}

		ms.websocketService.SendToAll(workspaceId, res)
		return nil
	case "ticket_note":
		ticketIndex, noteIndex, err := ms.findTicketIdAndNoteIdByNoteId(chat, messageId)
		if err != nil {
			return err
		}

		chat.Tickets[ticketIndex].Notes = append(chat.Tickets[ticketIndex].Notes[:noteIndex], chat.Tickets[ticketIndex].Notes[noteIndex+1:]...)
		if err = ms.messengerRepo.UpdateChat(nil, chat); err != nil {
			return err
		}

		res, err := json.Marshal(ms.createMessageResponse(nil, time.Time{}, false, "", "delete", workspaceId, ticketId, chatId, messageId, "", messageType))
		if err != nil {
			return err
		}

		ms.websocketService.SendToAll(workspaceId, res)
		return nil
	default:
		return errors.New("invalid message type")
	}
}

func (ms *MessengerServiceImpl) getAssigneeIdByTeam(workspace *entity.Workspace, teamId string) (primitive.ObjectID, error) {
	team, err := ms.messengerRepo.FindTeamByWorkspaceIdAndTeamId(workspace.Id, teamId)
	if err != nil {
		return primitive.ObjectID{}, err
	}

	var leastBusyMember primitive.ObjectID
	minTickets := int(^uint(0) >> 1)

	findMember := func(status entity.UserStatus) bool {
		for memberId := range team.Members {
			user, err := ms.messengerRepo.FindUserById(memberId)
			if err != nil || user.Status != status {
				continue
			}

			activeTicketsCount, err := ms.messengerRepo.CountActiveTickets(memberId)
			if err != nil {
				continue
			}
			if activeTicketsCount < minTickets {
				minTickets = activeTicketsCount
				leastBusyMember = memberId
			}
		}
		return !leastBusyMember.IsZero()
	}

	if findMember(entity.StatusAvailable) || findMember(entity.StatusBusy) || findMember(entity.StatusOffline) {
		return leastBusyMember, nil
	}

	return primitive.NilObjectID, errors.New("no suitable team member found")
}

func (ms *MessengerServiceImpl) validateTicketStatus(status string) (entity.TicketStatus, error) {
	switch entity.TicketStatus(status) {
	case entity.StatusOpen, entity.StatusPending, entity.StatusClosed:
		return entity.TicketStatus(status), nil
	default:
		return "", fmt.Errorf("invalid ticket status: %s", status)
	}
}

func (ms *MessengerServiceImpl) findTicketInChat(chat *entity.Chat, ticketId string) (*entity.Ticket, error) {
	for _, ticket := range chat.Tickets {
		if ticket.TicketId == ticketId {
			return &ticket, nil
		}
	}
	return nil, errors.New("ticket not found")
}

func (ms *MessengerServiceImpl) findNoteIndexInChat(chat *entity.Chat, noteId string) (int, error) {
	for i, note := range chat.Notes {
		if note.NoteId == noteId {
			return i, nil
		}
	}

	return -1, errors.New("invalid noteId")
}

func (ms *MessengerServiceImpl) findTicketIdAndNoteIdByNoteId(chat *entity.Chat, noteId string) (int, int, error) {
	for i, ticket := range chat.Tickets {
		for j, note := range ticket.Notes {
			if note.NoteId == noteId {
				return i, j, nil
			}
		}
	}

	return -1, -1, errors.New("invalid noteId")
}

func (ms *MessengerServiceImpl) updateWallpaper(workspaceId, chatId string, userId int) error {
	client := resty.New()

	reqBody := map[string]interface{}{
		"workspace_id": workspaceId,
		"user_id":      userId,
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+ms.config.Auth.IntegrationsServerSecretKey).
		SetBody(reqBody).
		Post(ms.config.Website.IntegrationsServerURL + "/point_ai/telegram_wrapper/get_user_avatar")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return errors.New("an error occured")
	}

	//compressedPhoto, err := utils.ValidatePhoto(resp.Body())
	//if err == nil {
	//	ms.fileService.SaveFile("chat."+chatId, compressedPhoto)
	//}

	ms.fileService.SaveFile("chat."+chatId, resp.Body())

	return nil
}

func (ms *MessengerServiceImpl) extractTicketData(tickets []entity.Ticket) (int, time.Duration, float64) {
	totalTickets := len(tickets)
	if totalTickets == 0 {
		return 0, 0, 0
	}

	var totalSolutionTime time.Duration
	var totalMessages int

	for _, ticket := range tickets {
		if !ticket.ResolvedAt.IsZero() {
			solutionTime := ticket.ResolvedAt.Sub(ticket.CreatedAt)
			totalSolutionTime += solutionTime
		}

		totalMessages += len(ticket.Messages)
	}

	averageSolutionTime := totalSolutionTime / time.Duration(totalTickets)

	averageNumberOfMessagesPerTicket := float64(totalMessages) / float64(totalTickets)

	return totalTickets, averageSolutionTime, averageNumberOfMessagesPerTicket
}

func (ms *MessengerServiceImpl) createChat(tgChatId int, tgClientId int, source entity.ChatSource, ticket entity.Ticket, workspaceId, assigneeId, teamId primitive.ObjectID, isImported bool, lastMessage entity.Message, name, company, clientEmail, clientPhone, address string) *entity.Chat {
	return &entity.Chat{
		UserId:      assigneeId,
		WorkspaceId: workspaceId,
		TeamId:      teamId,
		ChatId:      uuid.New().String(),
		TgChatId:    tgChatId,
		TgClientId:  tgClientId,
		Tickets:     []entity.Ticket{ticket},
		Notes:       []entity.Note{},
		Tags:        []string{},
		Source:      source,
		IsImported:  isImported,
		Name:        name,
		LastMessage: lastMessage,
		Company:     company,
		ClientEmail: clientEmail,
		ClientPhone: clientPhone,
		Address:     address,
		CreatedAt:   time.Now(),
	}
}

func (ms *MessengerServiceImpl) createTicket(notes []entity.Note, messages []entity.Message, createdAt time.Time) *entity.Ticket {
	return &entity.Ticket{
		TicketId:  uuid.New().String(),
		Subject:   "",
		Notes:     notes,
		Messages:  messages,
		Status:    entity.StatusClosed,
		CreatedAt: createdAt,
	}
}

// Messages This one is the one that goes to the database
func (ms *MessengerServiceImpl) createMessage(senderId primitive.ObjectID, messageId int, message, from string, messageType entity.MessageType, createdAt time.Time) *entity.Message {
	return &entity.Message{
		SenderId:        senderId,
		MessageId:       uuid.New().String(),
		MessageIdClient: messageId,
		Message:         message,
		From:            from,
		Type:            messageType,
		CreatedAt:       createdAt,
	}
}

func (ms *MessengerServiceImpl) createNote(userId primitive.ObjectID, message string) *entity.Note {
	return &entity.Note{
		UserId:    userId,
		Text:      message,
		NoteId:    uuid.New().String(),
		CreatedAt: time.Now(),
	}
}

// This one is the one that goes to the client
func (ms *MessengerServiceImpl) createMessageResponse(content []byte, createdAt time.Time, isOwner bool, name, action, workspaceId, ticketId, chatId, messageId, message, messageType string) *model.MessageResponse {
	return &model.MessageResponse{
		WorkspaceId: workspaceId,
		TicketId:    ticketId,
		ChatId:      chatId,
		MessageId:   messageId,
		Message:     message,
		Content:     content,
		Type:        messageType,
		Action:      action,
		Name:        name,
		IsOwner:     isOwner,
		CreatedAt:   createdAt,
	}
}

func (ms *MessengerServiceImpl) createChatResponse(workspaceId, chatId string, tgClientId, tgChatId int, tags []string, lastMessage model.MessageResponse, source string, isImported bool, createdAt time.Time, name string, logo []byte, notes []model.MessageResponse, language, company, clientEmail, clientPhone, address string, totalTickets int, averageSolutionTime time.Duration, averageNumberOfMessagesPerTicket float64) *model.ChatResponse {
	return &model.ChatResponse{
		WorkspaceId:                      workspaceId,
		ChatId:                           chatId,
		TgClientId:                       tgClientId,
		TgChatId:                         tgChatId,
		Tags:                             tags,
		LastMessage:                      lastMessage,
		Source:                           source,
		IsImported:                       isImported,
		Notes:                            notes,
		Name:                             name,
		Logo:                             logo,
		Language:                         language,
		Company:                          company,
		Address:                          address,
		ClientPhone:                      clientPhone,
		ClientEmail:                      clientEmail,
		TicketNumber:                     totalTickets,
		AverageSolutionTime:              averageSolutionTime,
		AverageNumberOfMessagesPerTicket: averageNumberOfMessagesPerTicket,
		CreatedAt:                        createdAt,
	}
}

func (ms *MessengerServiceImpl) GetLatestMessage(ticket entity.Ticket) *entity.Message {
	if len(ticket.Messages) == 0 {
		return nil
	}

	sort.Slice(ticket.Messages, func(i, j int) bool {
		return ticket.Messages[i].CreatedAt.After(ticket.Messages[j].CreatedAt)
	})

	return &ticket.Messages[0]
}

func (ms *MessengerServiceImpl) isAdmin(userRole entity.WorkspaceRole) bool {
	return userRole == entity.RoleAdmin
}

func (ms *MessengerServiceImpl) isOwner(userRole entity.WorkspaceRole) bool {
	return userRole == entity.RoleOwner
}
