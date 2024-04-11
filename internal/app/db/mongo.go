package db

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Point-AI/backend/config"
)

func ConnectToDB(cfg *config.Config) *mongo.Database {
	//uri := fmt.Sprintf("mongodb+srv://pointai:%s@pointai.lglqz3w.mongodb.net/?retryWrites=true&w=majority&appName=pointai", cfg.MongoDB.Password)

	opts := options.Client().ApplyURI("mongodb+srv://pointai:pointai@pointai.lglqz3w.mongodb.net/?retryWrites=true&w=majority&appName=pointai")
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
		panic(err)
	}

	db := client.Database(cfg.MongoDB.Database)

	return db
}
