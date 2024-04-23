package repository

import (
	"context"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/messenger/domain/entity"
	infrastructureInterface "github.com/Point-AI/backend/internal/messenger/service/interface"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MessengerRepositoryImpl struct {
	database *mongo.Database
	config   *config.Config
}

func NewMessengerRepositoryImpl(cfg *config.Config, db *mongo.Database) infrastructureInterface.MessengerRepository {
	return &MessengerRepositoryImpl{
		database: db,
		config:   cfg,
	}
}

func (mr *MessengerRepositoryImpl) FindWorkspaceByWorkspaceId(workspaceId string) (*entity.Workspace, error) {
	var workspace entity.Workspace
	err := mr.database.Collection(mr.config.MongoDB.WorkspaceCollection).FindOne(context.Background(), bson.M{"workspace_id": workspaceId}).Decode(&workspace)
	if err != nil {
		return &workspace, err
	}

	return &workspace, nil
}

func (mr *MessengerRepositoryImpl) AddTelegramIntegration(id primitive.ObjectID, botToken string) error {
	integration := entity.TelegramIntegration{
		BotToken: botToken,
		IsActive: true,
	}

	_, err := mr.database.Collection(mr.config.MongoDB.WorkspaceCollection).UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{
		"$push": bson.M{"integrations.telegram": integration},
	})
	if err != nil {
		return err
	}

	return nil
}

func (mr *MessengerRepositoryImpl) CheckBotExists(botToken string) (bool, error) {
	count, err := mr.database.Collection(mr.config.MongoDB.WorkspaceCollection).CountDocuments(context.Background(), bson.M{
		"integrations.telegram": bson.M{"$elemMatch": bson.M{"bot_token": botToken}},
	})
	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}
