package repository

import (
	"context"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/integration/domain/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type IntegrationRepositoryImpl struct {
	database *mongo.Database
	config   *config.Config
}

func NewIntegrationRepositoryImpl(db *mongo.Database, cfg *config.Config) *IntegrationRepositoryImpl {
	return &IntegrationRepositoryImpl{
		database: db,
		config:   cfg,
	}
}

func (ir *IntegrationRepositoryImpl) FindWorkspaceByWorkspaceId(workspaceId string) (*entity.Workspace, error) {
	var workspace entity.Workspace
	err := ir.database.Collection(ir.config.MongoDB.WorkspaceCollection).FindOne(context.Background(), bson.M{"workspace_id": workspaceId}).Decode(&workspace)
	if err != nil {
		return &workspace, err
	}

	return &workspace, nil
}
