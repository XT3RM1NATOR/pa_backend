package repository

import (
	"context"
	"errors"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/system/domain/entity"
	"github.com/Point-AI/backend/internal/system/service/interface"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type SystemRepositoryImpl struct {
	database   *mongo.Database
	config     *config.Config
	collection string
}

func NewSystemRepositoryImpl(db *mongo.Database, cfg *config.Config, collection string) infrastructureInterface.SystemRepository {
	return &SystemRepositoryImpl{
		database:   db,
		config:     cfg,
		collection: collection,
	}
}

func (sr *SystemRepositoryImpl) CreateProject(team []primitive.ObjectID, projectId, name string, ownerId primitive.ObjectID) error {
	project := &entity.Project{
		Name:      name,
		Team:      team,
		OwnerID:   ownerId,
		ProjectID: projectId,
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}

	err := sr.database.Collection(sr.config.MongoDB.ProjectCollection).FindOne(context.Background(), bson.M{"project_id": projectId}).Decode(&project)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}

	_, err = sr.database.Collection(sr.collection).InsertOne(context.Background(), project)
	if err != nil {
		return err
	}

	return nil
}

func (sr *SystemRepositoryImpl) ValidateTeam(team []string) ([]primitive.ObjectID, error) {
	userIds := make([]primitive.ObjectID, 0, len(team))

	for _, email := range team {
		var user entity.User
		err := sr.database.Collection(sr.config.MongoDB.UserCollection).FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, errors.New("user not found for email: " + email)
			}
			return nil, err
		}

		userIds = append(userIds, user.ID)
	}

	return userIds, nil
}

func (sr *SystemRepositoryImpl) FindProjectById(projectId string) (entity.Project, error) {
	var project entity.Project
	err := sr.database.Collection(sr.config.MongoDB.ProjectCollection).FindOne(context.Background(), bson.M{"project_id": projectId}).Decode(&project)
	if err != nil {
		return project, err
	}

	return project, nil
}

func (sr *SystemRepositoryImpl) RemoveUserFromProject(project entity.Project, userId primitive.ObjectID) error {
	filter := bson.M{"_id": project.ID, "team": userId}
	update := bson.M{"$pull": bson.M{"team": userId}}

	res, err := sr.database.Collection(sr.config.MongoDB.ProjectCollection).UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("user ID not found in the project team")
	}

	return nil
}
