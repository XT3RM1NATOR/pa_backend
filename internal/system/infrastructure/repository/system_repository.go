package repository

import (
	"context"
	"errors"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/system/domain/entity"
	"github.com/Point-AI/backend/internal/system/infrastructure/model"
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

func (sr *SystemRepositoryImpl) CreateWorkspace(ownerId primitive.ObjectID, pendingTeam map[string]entity.WorkspaceRole, workspaceId, name string) error {
	team := make(map[primitive.ObjectID]entity.WorkspaceRole)
	team[ownerId] = entity.RoleAdmin

	workspace := &entity.Workspace{
		Name:        name,
		Team:        team,
		PendingTeam: pendingTeam,
		WorkspaceId: workspaceId,
		CreatedAt:   primitive.NewDateTimeFromTime(time.Now()),
	}

	if _, err := sr.database.Collection(sr.config.MongoDB.WorkspaceCollection).InsertOne(context.Background(), workspace); err != nil {
		return err
	}
	return nil
}

func (sr *SystemRepositoryImpl) ValidateTeam(team map[string]string, ownerId primitive.ObjectID) (map[primitive.ObjectID]entity.WorkspaceRole, error) {
	userRoles := make(map[primitive.ObjectID]entity.WorkspaceRole)
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
		if _, exists := userRoles[user.Id]; exists {
			continue
		}

		switch role {
		case string(entity.RoleAdmin), string(entity.RoleMember), string(entity.RoleOwner):
			userRoles[user.Id] = entity.WorkspaceRole(role)
		default:
			userRoles[user.Id] = entity.RoleMember
		}
	}

	return userRoles, nil
}

func (sr *SystemRepositoryImpl) FormatTeam(team map[primitive.ObjectID]entity.WorkspaceRole) (map[string]string, error) {
	userRoles := make(map[string]string)

	for id, role := range team {
		var user entity.User
		err := sr.database.Collection(sr.config.MongoDB.UserCollection).FindOne(context.Background(), bson.M{"_id": id}).Decode(&user)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				continue
			}
			return nil, err
		}

		userRoles[user.Email] = string(role)
	}

	return userRoles, nil
}

func (sr *SystemRepositoryImpl) FindWorkspaceByWorkspaceId(WorkspaceId string) (entity.Workspace, error) {
	var workspace entity.Workspace
	err := sr.database.Collection(sr.config.MongoDB.WorkspaceCollection).FindOne(context.Background(), bson.M{"workspace_id": WorkspaceId}).Decode(&workspace)
	if err != nil {
		return workspace, err
	}

	return workspace, nil
}

func (sr *SystemRepositoryImpl) RemoveUserFromWorkspace(Workspace entity.Workspace, userId primitive.ObjectID) error {
	filter, update := bson.M{"_id": Workspace.Id}, bson.M{"$unset": bson.M{"team." + userId.Hex(): ""}}

	res, err := sr.database.Collection(sr.config.MongoDB.WorkspaceCollection).UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("user ID not found in the Workspace team")
	}

	return nil
}

func (sr *SystemRepositoryImpl) AddUsersToWorkspace(Workspace entity.Workspace, teamRoles map[primitive.ObjectID]entity.WorkspaceRole) error {
	for userId, role := range teamRoles {
		if _, exists := Workspace.Team[userId]; !exists {
			Workspace.Team[userId] = role
		}
	}
	filter, update := bson.M{"_id": Workspace.Id}, bson.M{"$set": bson.M{"team": Workspace.Team}}

	res, err := sr.database.Collection(sr.config.MongoDB.WorkspaceCollection).UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("workspace not found")
	}

	return nil
}

func (sr *SystemRepositoryImpl) UpdateUsersInWorkspace(Workspace entity.Workspace, teamRoles map[primitive.ObjectID]entity.WorkspaceRole) error {
	for userId, role := range teamRoles {
		if _, exists := Workspace.Team[userId]; exists {
			Workspace.Team[userId] = role
		}
	}
	filter, update := bson.M{"_id": Workspace.Id}, bson.M{"$set": bson.M{"team": Workspace.Team}}

	res, err := sr.database.Collection(sr.config.MongoDB.WorkspaceCollection).UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("workspace not found")
	}

	return nil
}

func (sr *SystemRepositoryImpl) GetUserProfiles(Workspace entity.Workspace) ([]model.User, error) {
	var users []model.User

	for userId, role := range Workspace.Team {
		user, err := sr.findUserById(userId)
		if err == nil {
			users = append(users, model.User{
				Email:    user.Email,
				FullName: user.FullName,
				Role:     string(role),
			})
		}
	}

	return users, nil
}

func (sr *SystemRepositoryImpl) findUserById(userId primitive.ObjectID) (entity.User, error) {
	var user entity.User

	err := sr.database.Collection(sr.config.MongoDB.UserCollection).FindOne(context.Background(), bson.M{"_id": userId}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return user, errors.New("user not found")
	}

	return user, nil
}

func (sr *SystemRepositoryImpl) DeleteWorkspace(id primitive.ObjectID) error {
	res, err := sr.database.Collection(sr.config.MongoDB.WorkspaceCollection).DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return errors.New("workspace not found")
	}

	return nil
}

func (sr *SystemRepositoryImpl) FindWorkspacesByUser(userID primitive.ObjectID) ([]entity.Workspace, error) {
	filter := bson.M{
		"team." + userID.Hex(): bson.M{"$exists": true},
	}

	cursor, err := sr.database.Collection(sr.config.MongoDB.WorkspaceCollection).Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var Workspaces []entity.Workspace
	if err := cursor.All(context.Background(), &Workspaces); err != nil {
		return nil, err
	}

	return Workspaces, nil
}

func (sr *SystemRepositoryImpl) FindUserById(userId primitive.ObjectID) (string, error) {
	var user entity.User
	err := sr.database.Collection(sr.config.MongoDB.UserCollection).FindOne(context.Background(), bson.M{"_id": userId}).Decode(&user)
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

	return user.Id, nil
}

func (sr *SystemRepositoryImpl) AddPendingInviteToUser(userId primitive.ObjectID, projectId string) error {
	filter := bson.M{"_id": userId}
	update := bson.M{"$addToSet": bson.M{"pending_invites": projectId}}
	_, err := sr.database.Collection(sr.config.MongoDB.UserCollection).UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (sr *SystemRepositoryImpl) UpdateWorkspace(Workspace entity.Workspace) error {
	filter, update := bson.M{"_id": Workspace.Id}, bson.M{"$set": Workspace}

	res, err := sr.database.Collection(sr.config.MongoDB.WorkspaceCollection).ReplaceOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("workspace not found")
	}

	return nil
}
