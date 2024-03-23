package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Point-AI/backend/config"
)

func ConnectToDB(cfg *config.Config) *mongo.Database {
	//uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s",
	//	cfg.MongoDB.User,
	//	cfg.MongoDB.Password,
	//	cfg.MongoDB.Host,
	//	cfg.MongoDB.Port,
	//	cfg.MongoDB.Database,
	//)
	uri := fmt.Sprintf("mongodb://%s/%s",
		cfg.MongoDB.Host,
		cfg.MongoDB.Database,
	)

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		panic(err)
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		panic(err)
	}

	db := client.Database(cfg.MongoDB.Database)

	return db
}
