package repository

import (
	"context"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/api/domain/entity"
	infrastructureInterface "github.com/Point-AI/backend/internal/api/service/interface"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type APIRepositoryImpl struct {
	database *mongo.Database
	config   *config.Config
}

func NewAPIRepositoryImpl(db *mongo.Database, cfg *config.Config) infrastructureInterface.APIRepository {
	return &APIRepositoryImpl{
		database: db,
		config:   cfg,
	}
}

func (ar *APIRepositoryImpl) CreateArticle(articleId int, lang entity.Language) error {
	article := entity.HelpDeskArticle{ArticleId: articleId, ViewCount: 1, Language: lang}
	if _, err := ar.database.Collection(ar.config.MongoDB.HelpDeskCollection).InsertOne(context.Background(), article); err != nil {
		return err
	}

	return nil
}

func (ar *APIRepositoryImpl) IncrementArticleViewCount(articleId int) error {
	_, err := ar.database.Collection(ar.config.MongoDB.HelpDeskCollection).UpdateOne(
		context.Background(),
		bson.M{"_id": articleId},
		bson.M{"$inc": bson.M{"view_count": 1}},
	)
	if err != nil {
		return err
	}

	return nil
}

func (ar *APIRepositoryImpl) GetArticleByIdAndLanguage(articleId int, lang entity.Language) (*entity.HelpDeskArticle, error) {
	var article *entity.HelpDeskArticle
	if err := ar.database.Collection(ar.config.MongoDB.HelpDeskCollection).FindOne(
		context.Background(),
		bson.M{"_id": articleId, "language": lang},
	).Decode(&article); err != nil {

		return article, err
	}

	return article, nil
}

func (ar *APIRepositoryImpl) GetAllArticlesByLanguage(lang entity.Language) ([]entity.HelpDeskArticle, error) {
	var articles []entity.HelpDeskArticle

	cursor, err := ar.database.Collection(ar.config.MongoDB.HelpDeskCollection).Find(
		context.Background(),
		bson.M{"language": lang},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var article entity.HelpDeskArticle
		if err := cursor.Decode(&article); err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

func (ar *APIRepositoryImpl) GetAllArticles() (*[]entity.HelpDeskArticle, error) {
	var articles []entity.HelpDeskArticle

	cursor, err := ar.database.Collection(ar.config.MongoDB.HelpDeskCollection).Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var article entity.HelpDeskArticle
		if err := cursor.Decode(&article); err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return &articles, nil
}
