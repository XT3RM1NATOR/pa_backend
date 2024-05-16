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

func (as *APIServiceImpl) IncrementViewCount(articleId int, language string) error {
	var langType entity.Language
	switch language {
	case "en":
		langType = entity.English
	case "uz":
		langType = entity.Uzbek
	case "uz-uz":
		langType = entity.UzbekCyr
	case "ru":
		langType = entity.Russian
	}

	_, err := as.apiRepo.GetArticleByIdAndLanguage(articleId, langType)
	if errors.Is(err, mongo.ErrNoDocuments) {
		if err := as.apiRepo.CreateArticle(articleId, langType); err != nil {
			return err
		}
		return nil
	}

	if err := as.apiRepo.IncrementArticleViewCount(articleId); err != nil {
		return err
	}

	return nil
}

func (as *APIServiceImpl) GetAllArticles(language string) ([]entity.HelpDeskArticle, error) {
	var langType entity.Language
	switch language {
	case "en":
		langType = entity.English
	case "uz":
		langType = entity.Uzbek
	case "uz-uz":
		langType = entity.UzbekCyr
	case "ru":
		langType = entity.Russian
	}

	return as.apiRepo.GetAllArticlesByLanguage(langType)
}
