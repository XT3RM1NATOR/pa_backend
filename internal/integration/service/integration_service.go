package service

import (
	"errors"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/integration/domain/entity"
	"github.com/Point-AI/backend/internal/integration/infrastructure/client"
	"github.com/Point-AI/backend/internal/integration/infrastructure/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IntegrationServiceImpl struct {
	integrationRepo *repository.IntegrationRepositoryImpl
	telegramClient  *client.TelegramClient
	config          *config.Config
}

func NewIntegrationServiceImpl(cfg *config.Config, integrationsRepo *repository.IntegrationRepositoryImpl, telegramClient *client.TelegramClient) *IntegrationServiceImpl {
	return &IntegrationServiceImpl{
		integrationRepo: integrationsRepo,
		telegramClient:  telegramClient,
		config:          cfg,
	}
}

func (is *IntegrationServiceImpl) RegisterBotIntegration(userId primitive.ObjectID, botToken, workspaceId string) error {
	workspace, err := is.integrationRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	if is.isAdmin(workspace.Team[userId]) || is.isOwner(workspace.Team[userId]) {

	}

	return errors.New("user does not have the permissions")
}

func (is *IntegrationServiceImpl) isAdmin(userRole entity.WorkspaceRole) bool {
	return userRole == entity.RoleAdmin
}

func (is *IntegrationServiceImpl) isOwner(userRole entity.WorkspaceRole) bool {
	return userRole == entity.RoleOwner
}
