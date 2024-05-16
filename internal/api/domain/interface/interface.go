package _interface

import "github.com/Point-AI/backend/internal/api/domain/entity"

type APIService interface {
	GetAllArticles(language string) ([]entity.HelpDeskArticle, error)
	IncrementViewCount(articleId int, language string) error
}
