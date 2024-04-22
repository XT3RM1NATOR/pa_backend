package _interface

import "go.mongodb.org/mongo-driver/bson/primitive"

type MessengerService interface {
	RegisterBotIntegration(userId primitive.ObjectID, botToken, workspaceId string) error
}

type WebsocketService interface {
}
