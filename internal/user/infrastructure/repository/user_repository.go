package repository

import (
	"context"
	"errors"
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
	database   *mongo.Database
	collection string
}

func NewUserRepositoryImpl(db *mongo.Database, collection string) infrastructureInterface.UserRepository {
	return &UserRepositoryImpl{
		database:   db,
		collection: collection,
	}
}

func (ur *UserRepositoryImpl) CreateUser(email, passwordHash, confirmToken string) error {
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
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}

	_, err = ur.database.Collection(ur.collection).InsertOne(context.Background(), user)
	return err
}

func (ur *UserRepositoryImpl) CreateOauth2User(email, authSource string) (string, error) {
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
		CreatedAt:   primitive.NewDateTimeFromTime(time.Now()),
	}

	if _, err := ur.database.Collection(ur.collection).InsertOne(context.Background(), user); err != nil {
		return "", err
	}

	return oAuth2Token, nil
}

func (ur *UserRepositoryImpl) GetUserByEmail(email string) (*entity.User, error) {
	var user entity.User
	err := ur.database.Collection(ur.collection).FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
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
	err := ur.database.Collection(ur.collection).FindOne(context.Background(), bson.M{"_id": id}).Decode(&user)
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
	err := ur.database.Collection(ur.collection).FindOne(
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
	err := ur.database.Collection(ur.collection).FindOne(
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
	_, err := ur.database.Collection(ur.collection).UpdateOne(context.Background(), bson.M{"_id": id}, update)
	return err
}

func (ur *UserRepositoryImpl) ClearRefreshToken(id primitive.ObjectID) error {
	update := bson.M{"$set": bson.M{"tokens.refresh_token": ""}}
	_, err := ur.database.Collection(ur.collection).UpdateOne(context.Background(), bson.M{"_id": id}, update)
	return err
}

func (ur *UserRepositoryImpl) ConfirmUser(userId primitive.ObjectID) error {
	update := bson.M{"$set": bson.M{
		"is_confirmed":         true,
		"tokens.confirm_token": "",
	}}
	_, err := ur.database.Collection(ur.collection).UpdateOne(context.Background(), bson.M{"_id": userId}, update)
	return err
}

func (ur *UserRepositoryImpl) UpdateUser(user *entity.User) error {
	_, err := ur.database.Collection(ur.collection).ReplaceOne(context.Background(), bson.M{"_id": user.ID}, user)
	return err
}
