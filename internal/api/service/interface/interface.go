package infrastructureInterface

import "github.com/Point-AI/backend/internal/api/domain/entity"

type APIRepository interface {
	GetArticleByIdAndLanguage(articleId int, lang entity.Language) (*entity.HelpDeskArticle, error)
	CreateArticle(articleId int, lang entity.Language) error
	IncrementArticleViewCount(articleId int) error
	GetAllArticles() (*[]entity.HelpDeskArticle, error)
	GetAllArticlesByLanguage(lang entity.Language) ([]entity.HelpDeskArticle, error)
}
