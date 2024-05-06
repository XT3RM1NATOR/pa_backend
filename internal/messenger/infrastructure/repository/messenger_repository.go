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
	"sync"
)

type MessengerRepositoryImpl struct {
	database *mongo.Database
	config   *config.Config
	mu       sync.RWMutex
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

func (mr *MessengerRepositoryImpl) FindWorkspaceByPhoneNumber(phoneNumber string) (*entity.Workspace, error) {
	filter := bson.M{"integrations.telegram.phone_number": phoneNumber}

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

func (mr *MessengerRepositoryImpl) FindWorkspaceByTicketId(ticketId string) (*entity.Workspace, error) {
	filter := bson.M{"tickets.ticket_id": ticketId}

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

func (mr *MessengerRepositoryImpl) GetAllWorkspaceRepositories() ([]*entity.Workspace, error) {
	var workspaces []*entity.Workspace

	cursor, err := mr.database.Collection(mr.config.MongoDB.WorkspaceCollection).Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var workspace entity.Workspace
		if err := cursor.Decode(&workspace); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, &workspace)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return workspaces, nil
}

func (mr *MessengerRepositoryImpl) FindWorkspaceByWorkspaceId(workspaceId string) (*entity.Workspace, error) {
	var workspace entity.Workspace
	err := mr.database.Collection(mr.config.MongoDB.WorkspaceCollection).FindOne(context.Background(), bson.M{"workspace_id": workspaceId}).Decode(&workspace)
	if err != nil {
		return &workspace, err
	}

	return &workspace, nil
}

func (mr *MessengerRepositoryImpl) GetUserById(id primitive.ObjectID) (*entity.User, error) {
	var user entity.User
	err := mr.database.Collection(mr.config.MongoDB.UserCollection).FindOne(context.Background(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
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

func (mr *MessengerRepositoryImpl) FindUserByEmail(email string) (primitive.ObjectID, error) {
	var user entity.User
	err := mr.database.Collection(mr.config.MongoDB.UserCollection).FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		return primitive.ObjectID{}, err
	}

	return user.Id, nil
}
