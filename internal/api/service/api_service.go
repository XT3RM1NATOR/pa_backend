package service

import (
	"errors"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/api/domain/entity"
	_interface "github.com/Point-AI/backend/internal/api/domain/interface"
	"github.com/Point-AI/backend/internal/api/service/interface"
	"go.mongodb.org/mongo-driver/mongo"
)

type APIServiceImpl struct {
	apiRepo infrastructureInterface.APIRepository
	config  *config.Config
}

func NewAPIServiceImpl(cfg *config.Config, apiRepo infrastructureInterface.APIRepository) _interface.APIService {
	return &APIServiceImpl{
		apiRepo: apiRepo,
		config:  cfg,
	}
}

func (as *APIServiceImpl) IncrementViewCount(articleId int) error {
	_, err := as.apiRepo.GetArticleById(articleId)
	if errors.Is(err, mongo.ErrNoDocuments) {
		if err := as.apiRepo.CreateArticle(articleId); err != nil {
			return err
		}
		return nil
	}

	if err := as.apiRepo.IncrementArticleViewCount(articleId); err != nil {
		return err
	}

	return nil
}

func (as *APIServiceImpl) GetAllArticles() (*[]entity.HelpDeskArticle, error) {
	articles, err := as.apiRepo.GetAllArticles()
	if err != nil {
		return nil, err
	}
	return articles, nil
}
