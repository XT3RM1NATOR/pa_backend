package repository

import (
	"context"
	"errors"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/user/domain/entity"
	"github.com/Point-AI/backend/internal/user/service/interface"
	"github.com/Point-AI/backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
	"time"
)

type UserRepositoryImpl struct {
	database *mongo.Database
	config   *config.Config
	mu       *sync.RWMutex
}

func NewUserRepositoryImpl(db *mongo.Database, config *config.Config, mu *sync.RWMutex) infrastructureInterface.UserRepository {
	return &UserRepositoryImpl{
		database: db,
		config:   config,
		mu:       mu,
	}
}

func (ur *UserRepositoryImpl) CreateUser(email, passwordHash, confirmToken string) error {
	ur.mu.Lock()
	defer ur.mu.Unlock()

	user := &entity.User{
		Email:        email,
		PasswordHash: passwordHash,
		IsConfirmed:  false,
		Tokens: entity.Tokens{
			ConfirmToken: confirmToken,
		},
		CreatedAt: time.Now(),
	}

	_, err := ur.database.Collection(ur.config.MongoDB.UserCollection).InsertOne(
		context.Background(),
		user,
	)

	return err
}

func (ur *UserRepositoryImpl) CreateOauth2User(email, authSource string) (string, error) {
	ur.mu.Lock()
	defer ur.mu.Unlock()

	existingUser, err := ur.GetUserByEmail(email)
	if err != nil {
		return "", err
	}

	oAuth2Token, err := utils.GenerateToken()
	if err != nil {
		return "", err
	}

	if existingUser != nil {
		existingUser.Tokens.OAuth2Token = oAuth2Token
		existingUser.IsConfirmed = true

		if err = ur.UpdateUser(existingUser); err != nil {
			return "", err
		}
		return oAuth2Token, nil
	}

	user := &entity.User{
		Email:       email,
		AuthSource:  authSource,
		Tokens:      entity.Tokens{OAuth2Token: oAuth2Token},
		IsConfirmed: true,
		CreatedAt:   time.Now(),
	}

	if _, err := ur.database.Collection(ur.config.MongoDB.UserCollection).InsertOne(
		context.Background(),
		user,
	); err != nil {
		return "", err
	}

	return oAuth2Token, nil
}

func (ur *UserRepositoryImpl) UpdateAllPendingWorkspaceInvites(userId primitive.ObjectID, email string) error {
	ur.mu.Lock()
	defer ur.mu.Unlock()

	cursor, err := ur.database.Collection(ur.config.MongoDB.WorkspaceCollection).Find(
		context.Background(),
		bson.M{"pending." + email: bson.M{"$exists": true}},
	)
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var workspace entity.Workspace
		if err := cursor.Decode(&workspace); err != nil {
			continue
		}
		teamRole := workspace.PendingTeam[email]

		delete(workspace.PendingTeam, email)
		workspace.Team[userId] = teamRole
		go ur.UpdateWorkspace(&workspace)
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	return nil
}

func (ur *UserRepositoryImpl) UpdateAllPendingWorkspaceTeamInvites(userId primitive.ObjectID, email string) error {
	ur.mu.Lock()
	defer ur.mu.Unlock()

	filter := bson.M{
		"pending_internal_teams.$." + email: bson.M{"$exists": true},
	}

	cursor, err := ur.database.Collection(ur.config.MongoDB.WorkspaceCollection).Find(
		context.Background(),
		filter,
	)
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var workspace entity.Workspace
		if err := cursor.Decode(&workspace); err != nil {
			continue
		}

		for teamName, members := range workspace.PendingInternalTeams {
			if _, exists := members[email]; exists {
				delete(workspace.PendingInternalTeams[teamName], email)
				workspace.InternalTeams[teamName][userId] = entity.StatusOffline

				go ur.UpdateWorkspace(&workspace)
			}
		}
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	return nil
}

func (ur *UserRepositoryImpl) GetUserByEmail(email string) (*entity.User, error) {
	ur.mu.RLock()
	defer ur.mu.RUnlock()

	var user entity.User
	err := ur.database.Collection(ur.config.MongoDB.UserCollection).FindOne(
		context.Background(),
		bson.M{"email": email},
	).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepositoryImpl) GetUserById(id primitive.ObjectID) (*entity.User, error) {
	ur.mu.RLock()
	defer ur.mu.RUnlock()

	var user entity.User
	err := ur.database.Collection(ur.config.MongoDB.UserCollection).FindOne(
		context.Background(),
		bson.M{"_id": id},
	).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepositoryImpl) GetUserByOAuth2Token(token string) (*entity.User, error) {
	ur.mu.RLock()
	defer ur.mu.RUnlock()

	var user entity.User
	err := ur.database.Collection(ur.config.MongoDB.UserCollection).FindOne(
		context.Background(),
		bson.M{"tokens.oauth2_token": token},
	).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepositoryImpl) GetUserByConfirmToken(token string) (*entity.User, error) {
	ur.mu.RLock()
	defer ur.mu.RUnlock()

	var user entity.User
	err := ur.database.Collection(ur.config.MongoDB.UserCollection).FindOne(
		context.Background(),
		bson.M{"tokens.confirm_token": token},
	).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepositoryImpl) SetResetToken(user *entity.User, token string) error {
	user.Tokens.ResetToken = token
	return ur.UpdateUser(user)
}

func (ur *UserRepositoryImpl) SetRefreshToken(user *entity.User, token string) error {
	user.Tokens.RefreshToken = token
	user.Tokens.OAuth2Token = ""
	return ur.UpdateUser(user)
}

func (ur *UserRepositoryImpl) ClearResetToken(id primitive.ObjectID, password string) error {
	ur.mu.Lock()
	defer ur.mu.Unlock()

	_, err := ur.database.Collection(ur.config.MongoDB.UserCollection).UpdateOne(
		context.Background(),
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"password": password, "tokens.reset_token": ""}},
	)
	return err
}

func (ur *UserRepositoryImpl) ClearRefreshToken(id primitive.ObjectID) error {
	ur.mu.Lock()
	defer ur.mu.Unlock()

	_, err := ur.database.Collection(ur.config.MongoDB.UserCollection).UpdateOne(
		context.Background(),
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"tokens.refresh_token": ""}},
	)
	return err
}

func (ur *UserRepositoryImpl) ConfirmUser(userId primitive.ObjectID) error {
	ur.mu.Lock()
	defer ur.mu.Unlock()

	_, err := ur.database.Collection(ur.config.MongoDB.UserCollection).UpdateOne(
		context.Background(),
		bson.M{"_id": userId},
		bson.M{"$set": bson.M{"is_confirmed": true, "tokens.confirm_token": ""}},
	)
	return err
}

func (ur *UserRepositoryImpl) UpdateUser(user *entity.User) error {
	ur.mu.Lock()
	defer ur.mu.Unlock()

	_, err := ur.database.Collection(ur.config.MongoDB.UserCollection).ReplaceOne(
		context.Background(),
		bson.M{"_id": user.Id},
		user,
	)
	return err
}

func (ur *UserRepositoryImpl) FindWorkspaceByWorkspaceId(workspaceId string) (*entity.Workspace, error) {
	ur.mu.RLock()
	defer ur.mu.RUnlock()

	var workspace entity.Workspace
	err := ur.database.Collection(ur.config.MongoDB.WorkspaceCollection).FindOne(
		context.Background(),
		bson.M{"workspace_id": workspaceId},
	).Decode(&workspace)
	if err != nil {
		return &workspace, err
	}

	return &workspace, nil
}

func (ur *UserRepositoryImpl) UpdateWorkspace(workspace *entity.Workspace) error {
	ur.mu.Lock()
	defer ur.mu.Unlock()

	res, err := ur.database.Collection(ur.config.MongoDB.WorkspaceCollection).ReplaceOne(
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
