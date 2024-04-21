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
	"net/mail"
	"time"
)

type UserRepositoryImpl struct {
	database *mongo.Database
	config   *config.Config
}

func NewUserRepositoryImpl(db *mongo.Database, config *config.Config) infrastructureInterface.UserRepository {
	return &UserRepositoryImpl{
		database: db,
		config:   config,
	}
}

func (ur *UserRepositoryImpl) CreateUser(pendingInvites []string, email, passwordHash, confirmToken string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return err
	}

	user := &entity.User{
		Email:        email,
		PasswordHash: passwordHash,
		IsConfirmed:  false,
		Tokens: entity.Tokens{
			ConfirmToken: confirmToken,
		},
		PendingInvites: pendingInvites,
		CreatedAt:      primitive.NewDateTimeFromTime(time.Now()),
	}

	_, err = ur.database.Collection(ur.config.MongoDB.UserCollection).InsertOne(context.Background(), user)
	return err
}

func (ur *UserRepositoryImpl) CreateOauth2User(pendingInvites []string, email, authSource string) (string, error) {
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
		Email:          email,
		AuthSource:     authSource,
		Tokens:         entity.Tokens{OAuth2Token: oAuth2Token},
		IsConfirmed:    true,
		PendingInvites: pendingInvites,
		CreatedAt:      primitive.NewDateTimeFromTime(time.Now()),
	}

	if _, err := ur.database.Collection(ur.config.MongoDB.UserCollection).InsertOne(context.Background(), user); err != nil {
		return "", err
	}

	return oAuth2Token, nil
}

func (ur *UserRepositoryImpl) GetAllPendingInvites(email string) ([]string, error) {
	var workspaceIds []string

	cursor, err := ur.database.Collection(ur.config.MongoDB.WorkspaceCollection).Find(context.Background(), bson.M{"pending." + email: bson.M{"$exists": true}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var workspace entity.Workspace
		if err := cursor.Decode(&workspace); err != nil {
			return nil, err
		}
		workspaceIds = append(workspaceIds, workspace.WorkspaceId)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return workspaceIds, nil
}

func (ur *UserRepositoryImpl) GetUserByEmail(email string) (*entity.User, error) {
	var user entity.User
	err := ur.database.Collection(ur.config.MongoDB.UserCollection).FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepositoryImpl) GetUserById(id primitive.ObjectID) (*entity.User, error) {
	var user entity.User
	err := ur.database.Collection(ur.config.MongoDB.UserCollection).FindOne(context.Background(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepositoryImpl) GetUserByOAuth2Token(token string) (*entity.User, error) {
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
	update := bson.M{"$set": bson.M{
		"password":           password,
		"tokens.reset_token": "",
	},
	}
	_, err := ur.database.Collection(ur.config.MongoDB.UserCollection).UpdateOne(context.Background(), bson.M{"_id": id}, update)
	return err
}

func (ur *UserRepositoryImpl) ClearRefreshToken(id primitive.ObjectID) error {
	update := bson.M{"$set": bson.M{"tokens.refresh_token": ""}}
	_, err := ur.database.Collection(ur.config.MongoDB.UserCollection).UpdateOne(context.Background(), bson.M{"_id": id}, update)
	return err
}

func (ur *UserRepositoryImpl) ConfirmUser(userId primitive.ObjectID) error {
	update := bson.M{"$set": bson.M{
		"is_confirmed":         true,
		"tokens.confirm_token": "",
	}}
	_, err := ur.database.Collection(ur.config.MongoDB.UserCollection).UpdateOne(context.Background(), bson.M{"_id": userId}, update)
	return err
}

func (ur *UserRepositoryImpl) UpdateUser(user *entity.User) error {
	_, err := ur.database.Collection(ur.config.MongoDB.UserCollection).ReplaceOne(context.Background(), bson.M{"_id": user.Id}, user)
	return err
}
