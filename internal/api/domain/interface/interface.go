package _interface

import "github.com/Point-AI/backend/internal/api/domain/entity"

type APIService interface {
	GetAllArticles() (*[]entity.HelpDeskArticle, error)
	IncrementViewCount(articleId int) error
}
