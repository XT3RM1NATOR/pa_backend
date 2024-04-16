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
	database *mongo.Database
	config   *config.Config
}

func NewSystemRepositoryImpl(db *mongo.Database, cfg *config.Config) infrastructureInterface.SystemRepository {
	return &SystemRepositoryImpl{
		database: db,
		config:   cfg,
	}
}

func (sr *SystemRepositoryImpl) CreateProject(team map[primitive.ObjectID]entity.ProjectRole, projectId, name string) error {
	project := &entity.Project{
		Name:      name,
		Team:      team,
		ProjectID: projectId,
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}

	if _, err := sr.database.Collection(sr.config.MongoDB.ProjectCollection).InsertOne(context.Background(), project); err != nil {
		return err
	}
	return nil
}

func (sr *SystemRepositoryImpl) ValidateTeam(team map[string]string, ownerId primitive.ObjectID) (map[primitive.ObjectID]entity.ProjectRole, error) {
	userRoles := make(map[primitive.ObjectID]entity.ProjectRole)
	userRoles[ownerId] = entity.RoleAdmin

	for email, role := range team {
		var user entity.User
		err := sr.database.Collection(sr.config.MongoDB.UserCollection).FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				continue
			}
			return nil, err
		}

		switch role {
		case string(entity.RoleAdmin), string(entity.RoleMember), string(entity.RoleObserver):
			userRoles[user.ID] = entity.ProjectRole(role)
		default:
			userRoles[user.ID] = entity.RoleMember
		}
	}

	return userRoles, nil
}

func (sr *SystemRepositoryImpl) FindProjectByProjectId(projectId string) (entity.Project, error) {
	var project entity.Project
	err := sr.database.Collection(sr.config.MongoDB.ProjectCollection).FindOne(context.Background(), bson.M{"project_id": projectId}).Decode(&project)
	if err != nil {
		return project, err
	}

	return project, nil
}

func (sr *SystemRepositoryImpl) RemoveUserFromProject(project entity.Project, userId primitive.ObjectID) error {
	filter, update := bson.M{"_id": project.ID}, bson.M{"$unset": bson.M{"team." + userId.Hex(): ""}}

	res, err := sr.database.Collection(sr.config.MongoDB.ProjectCollection).UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("user ID not found in the project team")
	}

	return nil
}

func (sr *SystemRepositoryImpl) AddUsersToProject(project entity.Project, teamRoles map[primitive.ObjectID]entity.ProjectRole) error {
	for userID, role := range teamRoles {
		if _, exists := project.Team[userID]; !exists {
			project.Team[userID] = role
		}
	}
	filter, update := bson.M{"_id": project.ID}, bson.M{"$set": bson.M{"team": project.Team}}

	res, err := sr.database.Collection(sr.config.MongoDB.ProjectCollection).UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("project not found")
	}

	return nil
}

func (sr *SystemRepositoryImpl) UpdateUsersInProject(project entity.Project, teamRoles map[primitive.ObjectID]entity.ProjectRole) error {
	for userID, role := range teamRoles {
		if _, exists := project.Team[userID]; exists {
			project.Team[userID] = role
		}
	}
	filter, update := bson.M{"_id": project.ID}, bson.M{"$set": bson.M{"team": project.Team}}

	res, err := sr.database.Collection(sr.config.MongoDB.ProjectCollection).UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("project not found")
	}

	return nil
}

func (sr *SystemRepositoryImpl) DeleteProject(id primitive.ObjectID) error {
	res, err := sr.database.Collection(sr.config.MongoDB.ProjectCollection).DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return errors.New("project not found")
	}

	return nil
}

func (sr *SystemRepositoryImpl) FindProjectsByUser(userID primitive.ObjectID) ([]entity.Project, error) {
	filter := bson.M{
		"team." + userID.Hex(): bson.M{"$exists": true},
	}

	cursor, err := sr.database.Collection(sr.config.MongoDB.ProjectCollection).Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var projects []entity.Project
	if err := cursor.All(context.Background(), &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

func (sr *SystemRepositoryImpl) FindUserById(userID primitive.ObjectID) (string, error) {
	var user entity.User
	err := sr.database.Collection(sr.config.MongoDB.UserCollection).FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return "", err
	}

	return user.Email, nil
}

func (sr *SystemRepositoryImpl) FindUserByEmail(email string) (primitive.ObjectID, error) {
	var user entity.User
	err := sr.database.Collection(sr.config.MongoDB.UserCollection).FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		return primitive.ObjectID{}, err
	}

	return user.ID, nil
}
