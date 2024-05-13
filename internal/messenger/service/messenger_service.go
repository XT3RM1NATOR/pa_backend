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
	"os"
	"sort"
	"strconv"
	"time"
)

type MessengerServiceImpl struct {
	messengerRepo    infrastructureInterface.MessengerRepository
	websocketService _interface.WebsocketService
	config           *config.Config
}

func NewMessengerServiceImpl(cfg *config.Config, messengerRepo infrastructureInterface.MessengerRepository, websocketService _interface.WebsocketService) _interface.MessengerService {
	return &MessengerServiceImpl{
		messengerRepo:    messengerRepo,
		websocketService: websocketService,
		config:           cfg,
	}
}

func (ms *MessengerServiceImpl) ReassignTicketToTeam(userId primitive.ObjectID, chatId string, ticketId, workspaceId, teamName string) error {
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

		assigneeId, err := ms.getAssigneeIdByTeam(workspace, teamName)
		if err != nil {
			return err
		}

		chat, err := ms.messengerRepo.FindChatByUserId(sc, originalChat.TgClientId, workspace.Id, assigneeId)
		if err != nil {
			return err
		} else if chat == nil {
			newChat := ms.createChat(originalChat, *ticketToMove, workspace.Id, assigneeId)
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
			newChat := ms.createChat(originalChat, *ticketToMove, workspace.Id, reassignUserId)
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
			go ms.updateWallpaper(workspaceId, chat.TgClientId)
		}
		messageResponse := ms.createMessageResponse(nil, chat.LastMessage.CreatedAt, userId == chat.LastMessage.SenderId, chat.LastMessage.From, "", workspaceId, chat.Tickets[0].TicketId, chat.ChatId, chat.LastMessage.MessageId, chat.LastMessage.Message, string(chat.LastMessage.Type))
		responseChats = append(responseChats, *ms.createChatResponse(workspace.WorkspaceId, chat.ChatId, chat.TgClientId, chat.TgChatId, chat.Tags, *messageResponse, string(entity.SourceTelegram), chat.IsImported, chat.CreatedAt, chat.Name))
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
		go ms.updateWallpaper(workspaceId, int(chat.Id))
		message := ms.createMessage(primitive.ObjectID{}, chat.LastMessage.Id, chat.LastMessage.Text, chat.Title, entity.TypeText, time.Now())
		ticket := ms.createTicket([]entity.Note{}, []entity.Message{*message}, time.Now())
		newChat := ms.createNewChat(int(chat.Id), int(chat.LastMessage.SenderId), entity.SourceTelegram, *ticket, workspace.Id, primitive.ObjectID{}, true, *message, chat.Name)

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

func (ms *MessengerServiceImpl) HandleMessage(userId primitive.ObjectID, workspaceId, ticketId, chatId, messageType, message string) error {
	if messageType == "chat_note" {
		chat, err := ms.messengerRepo.FindChatByChatId(chatId)
		if err != nil {
			return err
		}

		user, err := ms.messengerRepo.GetUserById(userId)
		if err != nil {
			return err
		}

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
	} else if messageType == "ticket_note" {
		chat, err := ms.messengerRepo.FindChatByChatId(chatId)
		if err != nil {
			return err
		}
		ticket, err := ms.findTicketInChat(chat, ticketId)
		if err != nil {
			return err
		}
		user, err := ms.messengerRepo.GetUserById(userId)
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
	}

	return errors.New("unknown message type")
}

func (ms *MessengerServiceImpl) UpdateChatInfo(userId primitive.ObjectID, chatId string, tags []string, workspaceId string) error {
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

	responseMessage := ms.createMessageResponse(nil, chat.LastMessage.CreatedAt, chat.UserId == userId, chat.LastMessage.From, "", workspaceId, "", chat.ChatId, chat.LastMessage.MessageId, chat.LastMessage.Message, string(entity.TypeText))
	responseChat := ms.createChatResponse(workspaceId, chatId, chat.TgClientId, chat.TgChatId, chat.Tags, *responseMessage, string(chat.Source), chat.IsImported, chat.CreatedAt, chat.Name)

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
			go ms.updateWallpaper(workspaceId, chat.TgClientId)
		}
		messageResponse := ms.createMessageResponse(nil, chat.LastMessage.CreatedAt, userId == chat.LastMessage.SenderId, chat.LastMessage.From, "", workspaceId, chat.Tickets[0].TicketId, chat.ChatId, chat.LastMessage.MessageId, chat.LastMessage.Message, string(chat.LastMessage.Type))
		responseChats = append(responseChats, *ms.createChatResponse(workspace.WorkspaceId, chat.ChatId, chat.TgClientId, chat.TgChatId, chat.Tags, *messageResponse, string(entity.SourceTelegram), chat.IsImported, chat.CreatedAt, chat.Name))
	}

	return responseChats, nil
}

func (ms *MessengerServiceImpl) DeleteMessage(userId primitive.ObjectID, messageType, workspaceId, ticketId, messageId, chatId string) error {
	if messageType == "chat_note" {
		workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(nil, workspaceId)
		if err != nil {
			return err
		}
		if _, exists := workspace.Team[userId]; !exists {
			return errors.New("unauthorized")
		}

		chat, err := ms.messengerRepo.FindChatByChatId(chatId)
		if err != nil {
			return err
		}
		if chat.WorkspaceId != workspace.Id {
			return errors.New("wrong chatId or workspaceId")
		}

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
	} else if messageType == "ticket_note" {
		workspace, err := ms.messengerRepo.FindWorkspaceByWorkspaceId(nil, workspaceId)
		if err != nil {
			return err
		}
		if _, exists := workspace.Team[userId]; !exists {
			return errors.New("unauthorized")
		}

		chat, err := ms.messengerRepo.FindChatByChatId(chatId)
		if err != nil {
			return err
		}
		if chat.WorkspaceId != workspace.Id {
			return errors.New("wrong chatId or workspaceId")
		}

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
	}

	return errors.New("invalid message type")
}

func (ms *MessengerServiceImpl) getAssigneeIdByTeam(workspace *entity.Workspace, teamName string) (primitive.ObjectID, error) {
	if team, exists := workspace.InternalTeams[teamName]; exists {
		return ms.findLeastBusyMember(team)
	}

	return primitive.NilObjectID, errors.New("specified team does not exist in the workspace")
}

func (ms *MessengerServiceImpl) validateTicketStatus(status string) (entity.TicketStatus, error) {
	switch entity.TicketStatus(status) {
	case entity.StatusOpen, entity.StatusPending, entity.StatusClosed:
		return entity.TicketStatus(status), nil
	default:
		return "", fmt.Errorf("invalid ticket status: %s", status)
	}
}

func (ms *MessengerServiceImpl) findLeastBusyMember(team map[primitive.ObjectID]entity.UserStatus) (primitive.ObjectID, error) {
	var leastBusyMember primitive.ObjectID
	minTickets := int(^uint(0) >> 1)

	findMember := func(status entity.UserStatus) bool {
		for memberId, userStatus := range team {
			if userStatus != status {
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

func (ms *MessengerServiceImpl) updateWallpaper(workspaceId string, userId int) error {
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

	imagePath := "../../telegram_static/" + strconv.FormatInt(int64(userId), 10) + ".jpg"
	err = os.WriteFile(imagePath, resp.Body(), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (ms *MessengerServiceImpl) createNewChat(tgChatId int, tgClientId int, source entity.ChatSource, ticket entity.Ticket, workspaceId, assigneeId primitive.ObjectID, isImported bool, lastMessage entity.Message, name string) *entity.Chat {
	return &entity.Chat{
		UserId:      assigneeId,
		WorkspaceId: workspaceId,
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
		CreatedAt:   time.Now(),
	}
}

func (ms *MessengerServiceImpl) createChat(currentChat *entity.Chat, ticket entity.Ticket, workspaceId, assigneeId primitive.ObjectID) *entity.Chat {
	return &entity.Chat{
		UserId:      assigneeId,
		WorkspaceId: workspaceId,
		ChatId:      uuid.New().String(),
		TgChatId:    currentChat.TgChatId,
		TgClientId:  currentChat.TgClientId,
		Tickets:     []entity.Ticket{ticket},
		Notes:       []entity.Note{},
		Tags:        []string{},
		Source:      currentChat.Source,
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

func (ms *MessengerServiceImpl) createChatResponse(workspaceId, chatId string, tgClientId, tgChatId int, tags []string, lastMessage model.MessageResponse, source string, isImported bool, createdAt time.Time, name string) *model.ChatResponse {
	return &model.ChatResponse{
		WorkspaceId: workspaceId,
		ChatId:      chatId,
		TgClientId:  tgClientId,
		TgChatId:    tgChatId,
		Tags:        tags,
		LastMessage: lastMessage,
		Source:      source,
		IsImported:  isImported,
		Name:        name,
		CreatedAt:   createdAt,
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
