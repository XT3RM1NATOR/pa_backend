package controller

import (
	"encoding/json"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/messenger/delivery/model"
	_interface "github.com/Point-AI/backend/internal/messenger/domain/interface"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type IntegrationController struct {
	messengerService _interface.MessengerService
	config           *config.Config
}

func NewIntegrationsController(messengerService _interface.MessengerService, cfg *config.Config) *IntegrationController {
	return &IntegrationController{
		messengerService: messengerService,
		config:           cfg,
	}
}

func (ic *IntegrationController) RegisterBotIntegration(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)
	var request model.RegisterBotRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	if err := ic.messengerService.RegisterBotIntegration(userId, request.BotToken, request.WorkspaceId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, model.SuccessResponse{Message: "bot added successfully"})
}

func (ic *IntegrationController) HandleBotMessage(c echo.Context) error {
	//token := c.Param("token")
	var update tgbotapi.Update
	if err := json.NewDecoder(c.Request().Body).Decode(&update); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	return nil
}
