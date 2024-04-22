package infrastructureInterface

import "github.com/Point-AI/backend/internal/api/domain/entity"

type APIRepository interface {
	GetArticleById(articleId int) (*entity.HelpDeskArticle, error)
	CreateArticle(articleId int) error
	IncrementArticleViewCount(articleId int) error
	GetAllArticles() (*[]entity.HelpDeskArticle, error)
}
