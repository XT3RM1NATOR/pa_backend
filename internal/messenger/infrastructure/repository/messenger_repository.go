package repository

import (
	"context"
	"errors"
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

func (mr *MessengerRepositoryImpl) UpdateWorkspace(workspace *entity.Workspace) error {
	filter, update := bson.M{"_id": workspace.Id}, bson.M{"$set": workspace}

	res, err := mr.database.Collection(mr.config.MongoDB.WorkspaceCollection).ReplaceOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("workspace not found")
	}

	return nil
}

func (mr *MessengerRepositoryImpl) FindWorkspaceByTelegramBotToken(botToken string) (*entity.Workspace, error) {
	filter := bson.M{
		"integrations.telegram_bot": bson.M{
			"$elemMatch": bson.M{
				"bot_token": botToken,
				"is_active": true,
			},
		},
	}

	var workspace entity.Workspace
	err := mr.database.Collection(mr.config.MongoDB.WorkspaceCollection).FindOne(context.Background(), filter).Decode(&workspace)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("workspace not found")
		}
		return nil, err
	}

	return &workspace, nil
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
	integration := entity.TelegramBotIntegration{
		BotToken: botToken,
		IsActive: true,
	}

	_, err := mr.database.Collection(mr.config.MongoDB.WorkspaceCollection).UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{
		"$push": bson.M{"integrations.telegram_bot": integration},
	})
	if err != nil {
		return err
	}

	return nil
}

func (mr *MessengerRepositoryImpl) CheckBotExists(botToken string) (bool, error) {
	count, err := mr.database.Collection(mr.config.MongoDB.WorkspaceCollection).CountDocuments(context.Background(), bson.M{
		"integrations.telegram_bot": bson.M{"$elemMatch": bson.M{"bot_token": botToken}},
	})
	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}
