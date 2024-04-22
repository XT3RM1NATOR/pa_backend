package repository

import (
	"context"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/messenger/domain/entity"
	infrastructureInterface "github.com/Point-AI/backend/internal/messenger/service/interface"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MessengerRepositoryImpl struct {
	database *mongo.Database
	config   *config.Config
}

func NewIntegrationRepositoryImpl(db *mongo.Database, cfg *config.Config) infrastructureInterface.MessengerRepository {
	return &MessengerRepositoryImpl{
		database: db,
		config:   cfg,
	}
}

func (ir *MessengerRepositoryImpl) FindWorkspaceByWorkspaceId(workspaceId string) (*entity.Workspace, error) {
	var workspace entity.Workspace
	err := ir.database.Collection(ir.config.MongoDB.WorkspaceCollection).FindOne(context.Background(), bson.M{"workspace_id": workspaceId}).Decode(&workspace)
	if err != nil {
		return &workspace, err
	}

	return &workspace, nil
}
