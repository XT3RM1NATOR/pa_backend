package controller

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/api/delivery/model"
	_interface "github.com/Point-AI/backend/internal/api/domain/interface"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type APIController struct {
	apiService _interface.APIService
	config     *config.Config
}

func NewAPIController(apiService _interface.APIService, cfg *config.Config) *APIController {
	return &APIController{
		apiService: apiService,
		config:     cfg,
	}
}

// HandlePost Handles Articles View Count.
// @Description Creates a new instance of a helpdesk article or increments its view count.
// @Tags HelpDesk
// @Accept json
// @Produce json
// @Param param article id
// @Success 201 {object} model.SuccessResponse "view count incremented successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/v1/article/{id} [post]
func (uc *APIController) HandlePost(c echo.Context) error {
	articleId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid article ID"})
	}

	if err := uc.apiService.IncrementViewCount(articleId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "view count incremented successfully"})
}

// GetAllArticles Returns all Articles and View Counts.
// @Description Returns all Articles and View Count.
// @Tags HelpDesk
// @Accept json
// @Produce json
// @Success 201 {object} []model.HelpDeskArticleResponse "User registered successfully"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/v1/articles [get]
func (uc *APIController) GetAllArticles(c echo.Context) error {
	articles, err := uc.apiService.GetAllArticles()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	var response []model.HelpDeskArticleResponse
	for _, article := range *articles {
		response = append(response, model.HelpDeskArticleResponse{
			ArticleId: article.ArticleId,
			ViewCount: article.ViewCount,
		})
	}

	return c.JSON(http.StatusOK, response)
}
