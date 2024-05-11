package repository

import (
	"context"
	"errors"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/messenger/domain/entity"
	infrastructureInterface "github.com/Point-AI/backend/internal/messenger/service/interface"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"sync"
	"time"
)

type MessengerRepositoryImpl struct {
	database *mongo.Database
	config   *config.Config
	mu       *sync.RWMutex
}

func NewMessengerRepositoryImpl(cfg *config.Config, db *mongo.Database, mu *sync.RWMutex) infrastructureInterface.MessengerRepository {
	return &MessengerRepositoryImpl{
		database: db,
		config:   cfg,
		mu:       mu,
	}
}

func (mr *MessengerRepositoryImpl) UpdateWorkspace(workspace *entity.Workspace) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	res, err := mr.database.Collection(mr.config.MongoDB.WorkspaceCollection).ReplaceOne(
		context.Background(),
		bson.M{"_id": workspace.Id},
		bson.M{"$set": workspace},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("workspace not found")
	}

	return nil
}

func (mr *MessengerRepositoryImpl) FindWorkspaceByTelegramBotToken(botToken string) (*entity.Workspace, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var workspace entity.Workspace
	err := mr.database.Collection(mr.config.MongoDB.WorkspaceCollection).FindOne(
		context.Background(),
		bson.M{
			"integrations.telegram_bot": bson.M{
				"$elemMatch": bson.M{
					"bot_token": botToken,
					"is_active": true,
				},
			},
		}).Decode(&workspace)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("workspace not found")
		}
		return nil, err
	}

	return &workspace, nil
}

func (mr *MessengerRepositoryImpl) FindChatByWorkspaceIdAndChatId(workspaceId primitive.ObjectID, chatId string) (*entity.Chat, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var chat entity.Chat
	err := mr.database.Collection(mr.config.MongoDB.ChatCollection).FindOne(
		context.Background(),
		bson.M{"workspace_id": workspaceId, "chat_id": chatId},
	).Decode(&chat)
	if err != nil {
		return nil, err
	}

	return &chat, nil
}

func (mr *MessengerRepositoryImpl) FindChatByChatId(chatId string) (*entity.Chat, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var chat entity.Chat
	err := mr.database.Collection(mr.config.MongoDB.ChatCollection).FindOne(
		context.Background(),
		bson.M{"chat_id": chatId},
	).Decode(&chat)
	if err != nil {
		return nil, err
	}

	return &chat, nil
}

func (mr *MessengerRepositoryImpl) FindChatByTicketId(ctx mongo.SessionContext, ticketId string) (*entity.Chat, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var chat entity.Chat
	err := mr.database.Collection(mr.config.MongoDB.ChatCollection).FindOne(
		ctx,
		bson.M{"tickets.ticket_id": ticketId},
	).Decode(&chat)
	if err != nil {
		return nil, err
	}

	return &chat, nil
}

func (mr *MessengerRepositoryImpl) FindWorkspaceByPhoneNumber(phoneNumber string) (*entity.Workspace, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var workspace entity.Workspace
	err := mr.database.Collection(mr.config.MongoDB.WorkspaceCollection).FindOne(
		context.Background(),
		bson.M{"integrations.telegram.phone_number": phoneNumber},
	).Decode(&workspace)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("workspace not found")
		}
		return nil, err
	}

	return &workspace, nil
}

func (mr *MessengerRepositoryImpl) FindWorkspaceByTicketId(ticketId string) (*entity.Workspace, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var workspace entity.Workspace
	err := mr.database.Collection(mr.config.MongoDB.WorkspaceCollection).FindOne(
		context.Background(),
		bson.M{"tickets.ticket_id": ticketId},
	).Decode(&workspace)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("workspace not found")
		}
		return nil, err
	}

	return &workspace, nil
}

func (mr *MessengerRepositoryImpl) GetAllWorkspaceRepositories() ([]*entity.Workspace, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var workspaces []*entity.Workspace

	cursor, err := mr.database.Collection(mr.config.MongoDB.WorkspaceCollection).Find(
		context.Background(),
		bson.M{},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var workspace entity.Workspace
		if err := cursor.Decode(&workspace); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, &workspace)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return workspaces, nil
}

func (mr *MessengerRepositoryImpl) FindWorkspaceByWorkspaceId(ctx mongo.SessionContext, workspaceId string) (*entity.Workspace, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var workspace entity.Workspace
	err := mr.database.Collection(mr.config.MongoDB.WorkspaceCollection).FindOne(
		ctx,
		bson.M{"workspace_id": workspaceId},
	).Decode(&workspace)
	if err != nil {
		return nil, err
	}

	return &workspace, nil
}

func (mr *MessengerRepositoryImpl) FindWorkspaceById(id primitive.ObjectID) (*entity.Workspace, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var workspace entity.Workspace
	err := mr.database.Collection(mr.config.MongoDB.WorkspaceCollection).FindOne(
		context.Background(),
		bson.M{"_id": id},
	).Decode(&workspace)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("workspace not found")
		}
		return nil, err
	}

	return &workspace, nil
}

func (mr *MessengerRepositoryImpl) FindChatByUserId(ctx mongo.SessionContext, tgClientId int, workspaceId, assigneeId primitive.ObjectID) (*entity.Chat, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var chat entity.Chat
	err := mr.database.Collection(mr.config.MongoDB.ChatCollection).FindOne(
		ctx,
		bson.M{"workspace_id": workspaceId, "user_id": assigneeId, "tg_client_id": tgClientId},
	).Decode(&chat)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &chat, nil
}

func (mr *MessengerRepositoryImpl) FindLatestChatsByWorkspaceId(workspaceId primitive.ObjectID, n int) ([]entity.Chat, error) {
	var chats []entity.Chat
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"workspace_id": workspaceId}
	opts := options.Find().SetSort(bson.M{"last_message.created_at": -1}).SetLimit(int64(n))

	cursor, err := mr.database.Collection(mr.config.MongoDB.ChatCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var chat entity.Chat
		if err = cursor.Decode(&chat); err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}

	if err = cursor.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}

func (mr *MessengerRepositoryImpl) InsertNewChat(ctx mongo.SessionContext, chat *entity.Chat) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	//c := context.Background()
	//
	//if ctx != nil {
	//	c = ctx
	//}
	log.Println(mr.config.MongoDB.ChatCollection)

	_, err := mr.database.Collection(mr.config.MongoDB.ChatCollection).InsertOne(
		context.Background(),
		chat,
	)

	return err
}

func (mr *MessengerRepositoryImpl) CountActiveTickets(memberId primitive.ObjectID) (int, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	filter := bson.M{
		"user_id":        memberId,
		"tickets.status": entity.StatusOpen,
	}
	pipeline := mongo.Pipeline{
		{{"$match", filter}},
		{{"$unwind", "$tickets"}},
		{{"$match", bson.M{"tickets.status": "open"}}},
		{{"$count", "activeTickets"}},
	}
	cursor, err := mr.database.Collection(mr.config.MongoDB.ChatCollection).Aggregate(
		context.Background(),
		pipeline,
	)
	if err != nil {
		return 0, err
	}

	var results []bson.M
	if err := cursor.All(context.Background(), &results); err != nil {
		return 0, err
	}
	if len(results) == 0 {
		return 0, nil
	}

	return int(results[0]["activeTickets"].(int32)), nil
}

func (mr *MessengerRepositoryImpl) GetUserById(id primitive.ObjectID) (*entity.User, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var user entity.User
	err := mr.database.Collection(mr.config.MongoDB.UserCollection).FindOne(
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

func (mr *MessengerRepositoryImpl) CheckBotExists(botToken string) (bool, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	count, err := mr.database.Collection(mr.config.MongoDB.WorkspaceCollection).CountDocuments(
		context.Background(),
		bson.M{"integrations.telegram_bot": bson.M{"$elemMatch": bson.M{"bot_token": botToken}}})
	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}

func (mr *MessengerRepositoryImpl) FindUserByEmail(ctx mongo.SessionContext, email string) (primitive.ObjectID, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var user entity.User
	err := mr.database.Collection(mr.config.MongoDB.UserCollection).FindOne(
		ctx,
		bson.M{"email": email},
	).Decode(&user)
	if err != nil {
		return primitive.ObjectID{}, err
	}

	return user.Id, nil
}

func (mr *MessengerRepositoryImpl) DeleteChat(ctx mongo.SessionContext, chatId primitive.ObjectID) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	_, err := mr.database.Collection(mr.config.MongoDB.ChatCollection).DeleteOne(
		ctx,
		bson.M{"_id": chatId},
	)
	return err
}

func (mr *MessengerRepositoryImpl) UpdateChat(ctx mongo.SessionContext, chat *entity.Chat) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	_, err := mr.database.Collection(mr.config.MongoDB.ChatCollection).UpdateOne(
		ctx,
		bson.M{"_id": chat.Id},
		bson.M{"$set": chat},
	)
	return err
}

func (mr *MessengerRepositoryImpl) StartSession() (mongo.Session, error) {
	session, err := mr.database.Client().StartSession()
	if err != nil {
		return nil, err
	}
	return session, nil
}
