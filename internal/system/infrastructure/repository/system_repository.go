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
	"sync"
	"time"
)

type SystemRepositoryImpl struct {
	database *mongo.Database
	config   *config.Config
	mu       *sync.RWMutex
}

func NewSystemRepositoryImpl(cfg *config.Config, db *mongo.Database, mu *sync.RWMutex) infrastructureInterface.SystemRepository {
	return &SystemRepositoryImpl{
		database: db,
		config:   cfg,
		mu:       mu,
	}
}

func (sr *SystemRepositoryImpl) CreateWorkspace(ownerId primitive.ObjectID, workspaceId, name string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	team := make(map[primitive.ObjectID]entity.WorkspaceRole)
	team[ownerId] = entity.RoleOwner

	workspace := &entity.Workspace{
		WorkspaceId: workspaceId,
		Name:        name,
		Team:        team,
		PendingTeam: make(map[string]entity.WorkspaceRole),
		Folders:     make(map[string][]string),
		Tags:        make([]string, 0),
		CreatedAt:   time.Now(),
	}

	_, err := sr.database.Collection(sr.config.MongoDB.WorkspaceCollection).InsertOne(
		context.Background(),
		workspace,
	)
	return err
}

func (sr *SystemRepositoryImpl) CountChatsByTeamId(teamId primitive.ObjectID) (int, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	count, err := sr.database.Collection(sr.config.MongoDB.ChatCollection).CountDocuments(
		context.Background(),
		bson.M{"team_id": teamId},
	)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (sr *SystemRepositoryImpl) ValidateTeam(team map[string]string, ownerId primitive.ObjectID) (map[primitive.ObjectID]entity.WorkspaceRole, map[string]entity.WorkspaceRole, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	userRoles := make(map[primitive.ObjectID]entity.WorkspaceRole)
	pendingUserRoles := make(map[string]entity.WorkspaceRole)

	for email, role := range team {
		var user entity.User
		err := sr.database.Collection(sr.config.MongoDB.UserCollection).FindOne(
			context.Background(),
			bson.M{"email": email},
		).Decode(&user)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				switch role {
				case string(entity.RoleAdmin), string(entity.RoleAgent), string(entity.RoleOwner):
					pendingUserRoles[email] = entity.WorkspaceRole(role)
				default:
					pendingUserRoles[email] = entity.RoleAgent
				}

				continue
			}
			return nil, nil, err
		}
		if _, exists := userRoles[user.Id]; exists {
			continue
		}

		switch role {
		case string(entity.RoleAdmin), string(entity.RoleAgent), string(entity.RoleOwner):
			userRoles[user.Id] = entity.WorkspaceRole(role)
		default:
			userRoles[user.Id] = entity.RoleAgent
		}
	}

	return userRoles, pendingUserRoles, nil
}

func (sr *SystemRepositoryImpl) ValidateNewTeamUsers(team map[string]string) (map[primitive.ObjectID]entity.WorkspaceRole, map[string]entity.WorkspaceRole, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	userRoles := make(map[primitive.ObjectID]entity.WorkspaceRole)
	pendingUserRoles := make(map[string]entity.WorkspaceRole)

	for email, role := range team {
		var user entity.User
		err := sr.database.Collection(sr.config.MongoDB.UserCollection).FindOne(
			context.Background(),
			bson.M{"email": email},
		).Decode(&user)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				switch role {
				case string(entity.RoleAdmin), string(entity.RoleAgent), string(entity.RoleOwner):
					pendingUserRoles[email] = entity.WorkspaceRole(role)
				default:
					pendingUserRoles[email] = entity.RoleAgent
				}

				continue
			}
			return nil, nil, err
		}
		if _, exists := userRoles[user.Id]; exists {
			continue
		}

		switch role {
		case string(entity.RoleAdmin), string(entity.RoleAgent), string(entity.RoleOwner):
			userRoles[user.Id] = entity.WorkspaceRole(role)
		default:
			userRoles[user.Id] = entity.RoleAgent
		}
	}

	return userRoles, pendingUserRoles, nil
}

func (sr *SystemRepositoryImpl) FormatTeam(team map[primitive.ObjectID]entity.WorkspaceRole) (map[string]string, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	userRoles := make(map[string]string)

	for id, role := range team {
		var user entity.User
		err := sr.database.Collection(sr.config.MongoDB.UserCollection).FindOne(
			context.Background(),
			bson.M{"_id": id},
		).Decode(&user)
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

func (sr *SystemRepositoryImpl) FindWorkspaceByWorkspaceId(workspaceId string) (*entity.Workspace, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	var workspace entity.Workspace
	err := sr.database.Collection(sr.config.MongoDB.WorkspaceCollection).FindOne(
		context.Background(),
		bson.M{"workspace_id": workspaceId},
	).Decode(&workspace)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return &workspace, nil
		}
		return &workspace, err
	}

	return &workspace, nil
}

func (sr *SystemRepositoryImpl) InsertNewTeam(team *entity.Team) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	_, err := sr.database.Collection(sr.config.MongoDB.TeamCollection).InsertOne(context.Background(), team)
	return err
}

func (sr *SystemRepositoryImpl) FindTeamByTeamId(teamId string) (*entity.Team, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	var team entity.Team
	err := sr.database.Collection(sr.config.MongoDB.TeamCollection).FindOne(
		context.Background(),
		bson.M{"team_id": teamId},
	).Decode(&team)
	if err != nil {
		return nil, err
	}

	return &team, nil
}

func (sr *SystemRepositoryImpl) FindTeamByTeamIdAndWorkspaceId(teamId string, workspaceId primitive.ObjectID) (*entity.Team, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	var team entity.Team
	err := sr.database.Collection("team").FindOne(
		context.Background(),
		bson.M{"team_id": teamId, "workspace_id": workspaceId},
	).Decode(&team)
	if err != nil {
		return nil, err
	}

	return &team, nil
}

func (sr *SystemRepositoryImpl) FindTeamsByWorkspaceId(workspaceId primitive.ObjectID) ([]*entity.Team, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	var teams []*entity.Team
	cursor, err := sr.database.Collection("team").Find(
		context.Background(),
		bson.M{"workspace_id": workspaceId},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var team entity.Team
		if err := cursor.Decode(&team); err != nil {
			return nil, err
		}
		teams = append(teams, &team)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	if len(teams) == 0 {
		return nil, nil
	}

	return teams, nil
}

func (sr *SystemRepositoryImpl) RemoveUserFromWorkspace(workspace *entity.Workspace, userId primitive.ObjectID) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	res, err := sr.database.Collection(sr.config.MongoDB.WorkspaceCollection).UpdateOne(
		context.Background(),
		bson.M{"_id": workspace.Id},
		bson.M{"$unset": bson.M{"team." + userId.Hex(): ""}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("user ID not found in the Workspace team")
	}

	return nil
}

// AddUsersToWorkspace TODO: delete initialization
func (sr *SystemRepositoryImpl) AddUsersToWorkspace(workspace *entity.Workspace, teamRoles map[primitive.ObjectID]entity.WorkspaceRole, pendingTeamRoles map[string]entity.WorkspaceRole) error {
	if workspace.Team == nil {
		workspace.Team = make(map[primitive.ObjectID]entity.WorkspaceRole)
	}
	if workspace.PendingTeam == nil {
		workspace.PendingTeam = make(map[string]entity.WorkspaceRole)
	}

	for userId, role := range teamRoles {
		if _, exists := workspace.Team[userId]; !exists {
			workspace.Team[userId] = role
		}
	}
	for email, role := range pendingTeamRoles {
		if _, exists := workspace.PendingTeam[email]; !exists {
			workspace.PendingTeam[email] = role
		}
	}

	return sr.UpdateWorkspace(workspace)
}

func (sr *SystemRepositoryImpl) UpdateUsersInWorkspace(workspace *entity.Workspace, teamRoles map[primitive.ObjectID]entity.WorkspaceRole) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	for userId, role := range teamRoles {
		if _, exists := workspace.Team[userId]; exists {
			workspace.Team[userId] = role
		}
	}

	res, err := sr.database.Collection(sr.config.MongoDB.WorkspaceCollection).UpdateOne(
		context.Background(),
		bson.M{"_id": workspace.Id},
		bson.M{"$set": bson.M{"team": workspace.Team}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("workspace not found")
	}

	return nil
}

func (sr *SystemRepositoryImpl) GetUserProfiles(Workspace entity.Workspace) (*[]infrastructureModel.User, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	var users []infrastructureModel.User

	for userId, role := range Workspace.Team {
		user, err := sr.findUserById(userId)
		if err == nil {
			users = append(users, infrastructureModel.User{
				Email:    user.Email,
				FullName: user.FullName,
				Role:     string(role),
			})
		}
	}

	return &users, nil
}

func (sr *SystemRepositoryImpl) findUserById(userId primitive.ObjectID) (*entity.User, error) {
	var user entity.User

	err := sr.database.Collection(sr.config.MongoDB.UserCollection).FindOne(
		context.Background(),
		bson.M{"_id": userId},
	).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return &user, errors.New("user not found")
	}

	return &user, nil
}

func (sr *SystemRepositoryImpl) DeleteWorkspace(id primitive.ObjectID) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	res, err := sr.database.Collection(sr.config.MongoDB.WorkspaceCollection).DeleteOne(
		context.Background(),
		bson.M{"_id": id},
	)
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return errors.New("workspace not found")
	}

	return nil
}

func (sr *SystemRepositoryImpl) DeleteTeam(id primitive.ObjectID) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	res, err := sr.database.Collection(sr.config.MongoDB.TeamCollection).DeleteOne(
		context.Background(),
		bson.M{"_id": id},
	)
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return errors.New("workspace not found")
	}

	return nil
}

func (sr *SystemRepositoryImpl) UpdateChatTeamIdToNil(teamId primitive.ObjectID) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	_, err := sr.database.Collection(sr.config.MongoDB.ChatCollection).UpdateMany(
		context.Background(),
		bson.M{"team_id": teamId},
		bson.M{"$set": bson.M{"team_id": primitive.NilObjectID}},
	)

	return err
}

func (sr *SystemRepositoryImpl) ClearPendingStatus(userId primitive.ObjectID, workspaceId string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	_, err := sr.database.Collection(sr.config.MongoDB.UserCollection).UpdateOne(
		context.Background(),
		bson.M{"_id": userId},
		bson.M{"$pull": bson.M{"pending_invites": workspaceId}},
	)
	if err != nil {
		return err
	}
	return nil
}

func (sr *SystemRepositoryImpl) UpdateWorkspaceUserStatus(userId primitive.ObjectID, workspaceId string, status bool) error {
	email, err := sr.FindUserEmailById(userId)
	if err != nil {
		return err
	}

	workspace, err := sr.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	if status {
		workspace.Team[userId] = workspace.PendingTeam[email]
	}
	delete(workspace.PendingTeam, email)

	if err := sr.UpdateWorkspace(workspace); err != nil {
		return err
	}

	return nil
}

func (sr *SystemRepositoryImpl) FindWorkspacesByUser(userId primitive.ObjectID) (*[]entity.Workspace, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	cursor, err := sr.database.Collection(sr.config.MongoDB.WorkspaceCollection).Find(
		context.Background(),
		bson.M{"team." + userId.Hex(): bson.M{"$exists": true}},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var workspaces []entity.Workspace
	if err := cursor.All(context.Background(), &workspaces); err != nil {
		return nil, err
	}

	return &workspaces, nil
}

func (sr *SystemRepositoryImpl) FindUserEmailById(userId primitive.ObjectID) (string, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	var user entity.User
	err := sr.database.Collection(sr.config.MongoDB.UserCollection).FindOne(
		context.Background(),
		bson.M{"_id": userId},
	).Decode(&user)
	if err != nil {
		return "", err
	}

	return user.Email, nil
}

func (sr *SystemRepositoryImpl) FindUserById(userId primitive.ObjectID) (*entity.User, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	var user entity.User
	err := sr.database.Collection(sr.config.MongoDB.UserCollection).FindOne(
		context.Background(),
		bson.M{"_id": userId},
	).Decode(&user)
	if err != nil {
		return &entity.User{}, err
	}

	return &user, nil
}

func (sr *SystemRepositoryImpl) FindUserByEmail(email string) (primitive.ObjectID, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	var user entity.User
	err := sr.database.Collection(sr.config.MongoDB.UserCollection).FindOne(
		context.Background(),
		bson.M{"email": email},
	).Decode(&user)
	if err != nil {
		return primitive.ObjectID{}, err
	}

	return user.Id, nil
}

func (sr *SystemRepositoryImpl) AddPendingInviteToUser(userId primitive.ObjectID, projectId string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	_, err := sr.database.Collection(sr.config.MongoDB.UserCollection).UpdateOne(
		context.Background(),
		bson.M{"_id": userId},
		bson.M{"$addToSet": bson.M{"pending_invites": projectId}},
	)
	return err
}

func (sr *SystemRepositoryImpl) UpdateWorkspace(workspace *entity.Workspace) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	res, err := sr.database.Collection(sr.config.MongoDB.WorkspaceCollection).ReplaceOne(
		context.Background(),
		bson.M{"_id": workspace.Id},
		workspace,
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("workspace not found")
	}

	return nil
}

func (sr *SystemRepositoryImpl) UpdateTeam(team *entity.Team) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	res, err := sr.database.Collection(sr.config.MongoDB.TeamCollection).ReplaceOne(
		context.Background(),
		bson.M{"_id": team.Id},
		team,
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("team not found")
	}

	return nil
}

func (sr *SystemRepositoryImpl) GetTeamNamesByUserId(userId primitive.ObjectID) []entity.Team {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	var teams []entity.Team
	cursor, err := sr.database.Collection(sr.config.MongoDB.TeamCollection).Find(
		context.Background(),
		bson.M{},
	)
	if err != nil {
		return nil
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var team entity.Team
		if err := cursor.Decode(&team); err != nil {
			return nil
		}
		teams = append(teams, team)
	}

	if err := cursor.Err(); err != nil {
		return nil
	}

	var teamNames []entity.Team
	for _, team := range teams {
		if _, exists := team.Members[userId]; exists {
			teamNames = append(teamNames, team)
		}
	}

	return teamNames
}
