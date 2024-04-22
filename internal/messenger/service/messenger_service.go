package service

import (
	"errors"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/messenger/domain/entity"
	_interface "github.com/Point-AI/backend/internal/messenger/domain/interface"
	infrastructureInterface "github.com/Point-AI/backend/internal/messenger/service/interface"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessengerServiceImpl struct {
	messengerRepo  infrastructureInterface.MessengerRepository
	telegramClient infrastructureInterface.TelegramClient
	config         *config.Config
}

func NewIntegrationServiceImpl(cfg *config.Config, messengerRepo infrastructureInterface.MessengerRepository, telegramClient infrastructureInterface.TelegramClient) _interface.MessengerService {
	return &MessengerServiceImpl{
		messengerRepo:  messengerRepo,
		telegramClient: telegramClient,
		config:         cfg,
	}
}

func (is *MessengerServiceImpl) RegisterBotIntegration(userId primitive.ObjectID, botToken, workspaceId string) error {
	workspace, err := is.messengerRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	if is.isAdmin(workspace.Team[userId]) || is.isOwner(workspace.Team[userId]) {

	}

	return errors.New("user does not have the permissions")
}

func (is *MessengerServiceImpl) isAdmin(userRole entity.WorkspaceRole) bool {
	return userRole == entity.RoleAdmin
}

func (is *MessengerServiceImpl) isOwner(userRole entity.WorkspaceRole) bool {
	return userRole == entity.RoleOwner
}
