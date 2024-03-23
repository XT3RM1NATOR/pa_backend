package repository

import (
	"context"
	"errors"
	"github.com/Point-AI/backend/internal/auth/infrastructure/model"
	"github.com/Point-AI/backend/utils"
	"go.mongodb.org/mongo-driver/bson"
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

func (ur *UserRepository) CreateUser(email string, passwordHash string, confirmToken string) error {
	user := &model.User{
		Email:        email,
		PasswordHash: passwordHash,
		IsConfirmed:  false,
		ConfirmToken: confirmToken,
		CreatedAt:    time.Now(),
	}

	_, err := ur.database.Collection(ur.collection).InsertOne(context.Background(), user)
	return err
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

func (ur *UserRepository) GetUserByResetToken(token string) (*model.User, error) {
	var user model.User
	err := ur.database.Collection(ur.collection).FindOne(context.Background(), bson.M{"resetToken": token}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepository) UpdateUser(user *model.User) error {
	_, err := ur.database.Collection(ur.collection).ReplaceOne(context.Background(), bson.M{"_id": user.ID}, user)
	return err
}

func (ur *UserRepository) SetResetToken(user *model.User, token string) error {
	user.ResetToken = token
	return ur.UpdateUser(user)
}

func (ur *UserRepository) ClearResetToken(user *model.User, password string) error {
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return errors.New("error hashing the password")
	}

	user.PasswordHash = passwordHash
	user.ResetToken = ""
	return ur.UpdateUser(user)
}

func (ur *UserRepository) UpdatePassword(user *model.User, newPassword string) error {
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hashedPassword
	return ur.UpdateUser(user)
}

func (ur *UserRepository) GetUserByConfirmToken(token string) (*model.User, error) {
	var user model.User
	err := ur.database.Collection(ur.collection).FindOne(context.Background(), bson.M{"confirmToken": token}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepository) ConfirmUser(user *model.User) error {
	update := bson.M{"$set": bson.M{"isConfirmed": true, "confirmToken": ""}}
	_, err := ur.database.Collection(ur.collection).UpdateOne(context.Background(), bson.M{"_id": user.ID}, update)
	return err
}
