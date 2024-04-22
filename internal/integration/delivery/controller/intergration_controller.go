package controller

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/integration/delivery/model"
	"github.com/Point-AI/backend/internal/integration/service"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type IntegrationController struct {
	integrationService *service.IntegrationServiceImpl
	config             *config.Config
}

func NewIntegrationsController(integrationService *service.IntegrationServiceImpl, cfg *config.Config) *IntegrationController {
	return &IntegrationController{
		integrationService: integrationService,
		config:             cfg,
	}
}

func (ic *IntegrationController) RegisterBotIntegration(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)
	var request model.RegisterBotRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	if err := ic.integrationService.RegisterBotIntegration(userId, request.BotToken, request.WorkspaceId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, model.SuccessResponse{Message: "bot added successfully"})
}
