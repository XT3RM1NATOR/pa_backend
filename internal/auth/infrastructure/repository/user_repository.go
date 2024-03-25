package repository

import (
	"context"
	"errors"
	"github.com/Point-AI/backend/internal/auth/infrastructure/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

func (ur *UserRepository) CreateUser(email string, passwordHash string, confirmToken string, fullName string) error {
	user := &model.User{
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		IsConfirmed:  false,
		Token: model.Token{
			ConfirmToken: confirmToken,
		},
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}

	_, err := ur.database.Collection(ur.collection).InsertOne(context.Background(), user)
	return err
}

func (ur *UserRepository) CreateOauth2User(email, authSource, name string) (*model.User, error) {
	existingUser, err := ur.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return existingUser, nil
	}

	user := &model.User{
		Email:       email,
		AuthSource:  authSource,
		FullName:    name,
		IsConfirmed: true,
		CreatedAt:   primitive.NewDateTimeFromTime(time.Now()),
	}

	result, err := ur.database.Collection(ur.collection).InsertOne(context.Background(), user)
	if err != nil {
		return nil, err
	}

	err = ur.database.Collection(ur.collection).FindOne(context.Background(), bson.M{"_id": result.InsertedID}).Decode(user)
	if err != nil {
		return nil, err
	}

	return user, nil
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

func (ur *UserRepository) GetUserByConfirmToken(token string) (*model.User, error) {
	var user model.User
	err := ur.database.Collection(ur.collection).FindOne(
		context.Background(),
		bson.M{"token.confirmToken": token},
	).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepository) SetResetToken(user *model.User, token string) error {
	user.Token.ResetToken = token
	return ur.updateUser(user)
}

func (ur *UserRepository) SetRefreshToken(user *model.User, token string) error {
	user.Token.RefreshToken = token
	return ur.updateUser(user)
}

func (ur *UserRepository) ClearRefreshToken(id primitive.ObjectID) error {
	update := bson.M{"$set": bson.M{"token.refreshToken": ""}}
	_, err := ur.database.Collection(ur.collection).UpdateOne(context.Background(), bson.M{"_id": id}, update)
	return err
}

func (ur *UserRepository) ClearResetToken(id primitive.ObjectID, password string) error {
	update := bson.M{"$set": bson.M{
		"password":         password,
		"token.resetToken": "",
	},
	}
	_, err := ur.database.Collection(ur.collection).UpdateOne(context.Background(), bson.M{"_id": id}, update)
	return err
}

func (ur *UserRepository) ConfirmUser(userId primitive.ObjectID) error {
	update := bson.M{"$set": bson.M{
		"isConfirmed":        true,
		"token.confirmToken": "",
	}}
	_, err := ur.database.Collection(ur.collection).UpdateOne(context.Background(), bson.M{"_id": userId}, update)
	return err
}

func (ur *UserRepository) updateUser(user *model.User) error {
	_, err := ur.database.Collection(ur.collection).ReplaceOne(context.Background(), bson.M{"_id": user.ID}, user)
	return err
}
