package repository

import (
	"context"
	"errors"
	"github.com/Point-AI/backend/internal/auth/infrastructure/model"
	"github.com/Point-AI/backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/mail"
	"time"
)

type UserRepository struct {
	database   *mongo.Database
	collection string
}

func NewUserRepository(db *mongo.Database, collection string) *UserRepository {
	return &UserRepository{
		database:   db,
		collection: collection,
	}
}

func (ur *UserRepository) CreateUser(email, passwordHash, confirmToken, fullName string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return err
	}

	user := &model.User{
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		IsConfirmed:  false,
		Tokens: model.Tokens{
			ConfirmToken: confirmToken,
		},
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}

	_, err = ur.database.Collection(ur.collection).InsertOne(context.Background(), user)
	return err
}

func (ur *UserRepository) CreateOauth2User(email, authSource, name string) (string, error) {
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

		if err = ur.updateUser(existingUser); err != nil {
			return "", err
		}
		return oAuth2Token, nil
	}

	user := &model.User{
		Email:       email,
		AuthSource:  authSource,
		FullName:    name,
		Tokens:      model.Tokens{OAuth2Token: oAuth2Token},
		IsConfirmed: true,
		CreatedAt:   primitive.NewDateTimeFromTime(time.Now()),
	}

	if _, err := ur.database.Collection(ur.collection).InsertOne(context.Background(), user); err != nil {
		return "", err
	}

	return oAuth2Token, nil
}

func (ur *UserRepository) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := ur.database.Collection(ur.collection).FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepository) GetUserById(id primitive.ObjectID) (*model.User, error) {
	var user model.User
	err := ur.database.Collection(ur.collection).FindOne(context.Background(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepository) GetUserByOAuth2Token(token string) (*model.User, error) {
	var user model.User
	err := ur.database.Collection(ur.collection).FindOne(
		context.Background(),
		bson.M{"tokens.oAuth2Token": token},
	).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepository) GetUserByConfirmToken(token string) (*model.User, error) {
	var user model.User
	err := ur.database.Collection(ur.collection).FindOne(
		context.Background(),
		bson.M{"tokens.confirmToken": token},
	).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepository) SetResetToken(user *model.User, token string) error {
	user.Tokens.ResetToken = token
	return ur.updateUser(user)
}

func (ur *UserRepository) SetRefreshToken(user *model.User, token string) error {
	user.Tokens.RefreshToken = token
	user.Tokens.OAuth2Token = ""
	return ur.updateUser(user)
}

func (ur *UserRepository) ClearResetToken(id primitive.ObjectID, password string) error {
	update := bson.M{"$set": bson.M{
		"password":          password,
		"tokens.resetToken": "",
	},
	}
	_, err := ur.database.Collection(ur.collection).UpdateOne(context.Background(), bson.M{"_id": id}, update)
	return err
}

func (ur *UserRepository) ClearRefreshToken(id primitive.ObjectID) error {
	update := bson.M{"$set": bson.M{"tokens.refreshToken": ""}}
	_, err := ur.database.Collection(ur.collection).UpdateOne(context.Background(), bson.M{"_id": id}, update)
	return err
}

func (ur *UserRepository) ConfirmUser(userId primitive.ObjectID) error {
	update := bson.M{"$set": bson.M{
		"isConfirmed":         true,
		"tokens.confirmToken": "",
	}}
	_, err := ur.database.Collection(ur.collection).UpdateOne(context.Background(), bson.M{"_id": userId}, update)
	return err
}

func (ur *UserRepository) updateUser(user *model.User) error {
	_, err := ur.database.Collection(ur.collection).ReplaceOne(context.Background(), bson.M{"_id": user.ID}, user)
	return err
}
