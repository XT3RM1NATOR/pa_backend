package controller

import (
	"encoding/json"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/messenger/delivery/model"
	_interface "github.com/Point-AI/backend/internal/messenger/domain/interface"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
)

type MessengerController struct {
	messengerService _interface.MessengerService
	websocketService _interface.WebsocketService
	config           *config.Config
}

func NewMessengerController(cfg *config.Config, messengerService _interface.MessengerService, websocketService _interface.WebsocketService) *MessengerController {
	return &MessengerController{
		messengerService: messengerService,
		config:           cfg,
	}
}

// RegisterBotIntegration registers a new bot integration for a workspace.
// @Summary Registers a new bot integration.
// @Tags Messenger
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param request body model.RegisterBotRequest true "Bot registration details"
// @Success 201 {object} model.SuccessResponse "Bot added successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request, unable to parse the request body"
// @Failure 500 {object} model.ErrorResponse "Internal server error, failed to register the bot"
// @Router /integrations/telegram/bots [post]
func (mc *MessengerController) RegisterBotIntegration(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)
	var request model.RegisterBotRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	if err := mc.messengerService.RegisterBotIntegration(userId, request.BotToken, request.WorkspaceId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, model.SuccessResponse{Message: "bot added successfully"})
}

func (mc *MessengerController) HandleBotMessage(c echo.Context) error {
	token := c.Param("token")
	var update *tgbotapi.Update
	if err := json.NewDecoder(c.Request().Body).Decode(&update); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	if err := mc.messengerService.HandleTelegramBotMessage(token, update); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return nil
}

func (mc *MessengerController) HandleTelegramClientAuth(c echo.Context) error {
	workspaceId, action := c.Param("id"), c.QueryParam("set")
	value, userId := c.QueryParam(action), c.Request().Context().Value("userId").(primitive.ObjectID)

	status, err := mc.messengerService.HandleTelegramClientAuth(userId, workspaceId, action, value)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, model.TelegramStatusResponse{Status: string(status)})
}

// WSHandler handles WebSocket connections for real-time messaging.
// @Summary Handles WebSocket connections.
// @Tags Messenger
// @Produce json
// @Param id path string true "Workspace ID"
// @Param userId path string true "User ID"
// @Success 200 {object} model.SuccessResponse "Connection upgraded successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request, user not valid in workspace"
// @Failure 500 {object} model.ErrorResponse "Internal server error, failed to upgrade connection"
// @Router /messenger/ws/{id} [get]
func (mc *MessengerController) WSHandler(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)
	workspaceId := c.Param("id")

	err := mc.messengerService.ValidateUserInWorkspace(userId, workspaceId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	ws, err := mc.websocketService.UpgradeConnection(c.Response(), c.Request(), workspaceId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}
	//mc.websocketService.

	go func() {
		defer mc.websocketService.RemoveConnection(workspaceId, ws)
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				break
			}

			var receivedMessage model.MessageRequest
			if err := json.Unmarshal(message, &receivedMessage); err != nil {
				continue
			}

			if receivedMessage.Source == "telegramBot" {
				err = mc.messengerService.HandleTelegramPlatformMessageToBot(receivedMessage, workspaceId, userId)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "connection upgraded successfully"})
}

// ReassignTicketToMember reassigns a support ticket to a different member.
// @Summary Reassigns a support ticket to a member.
// @Tags Messenger
// @Accept json
// @Produce json
// @Param ticket_id path string true "Ticket ID"
// @Param id path string true "Workspace ID"
// @Param email path string true "Member email"
// @Param userId path string true "User ID"
// @Success 200 {object} model.SuccessResponse "Ticket successfully reassigned to member"
// @Failure 500 {object} model.ErrorResponse "Internal server error, failed to reassign ticket"
// @Router /messenger/ticket/reassign/member/{ticket_id}/{id}/{email} [post]
func (mc *MessengerController) ReassignTicketToMember(c echo.Context) error {
	ticketId, workspaceId, userEmail := c.Param("ticket_id"), c.Param("id"), c.Param("email")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	if err := mc.messengerService.ReassignTicketToMember(userId, ticketId, workspaceId, userEmail); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "ticket successfully reassigned to " + userEmail})
}

// ReassignTicketToTeam reassigns a support ticket to a different team.
// @Summary Reassigns a support ticket to a team.
// @Tags Messenger
// @Accept json
// @Produce json
// @Param ticket_id path string true "Ticket ID"
// @Param id path string true "Workspace ID"
// @Param name path string true "Team name"
// @Param userId path string true "User ID"
// @Success 200 {object} model.SuccessResponse "Ticket successfully reassigned to team"
// @Failure 500 {object} model.ErrorResponse "Internal server error, failed to reassign ticket"
// @Router /messenger/ticket/reassign/team/{ticket_id}/{id}/{name} [post]
func (mc *MessengerController) ReassignTicketToTeam(c echo.Context) error {
	ticketId, workspaceId, teamName := c.Param("ticket_id"), c.Param("id"), c.Param("name")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	if err := mc.messengerService.ReassignTicketToMember(userId, ticketId, workspaceId, teamName); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "ticket successfully reassigned to " + teamName})
}

// CloseTicket updates the status of a support ticket.
// @Summary Updates the status of a support ticket.
// @Tags Messenger
// @Accept json
// @Produce json
// @Param status path string true "New status of the ticket"
// @Param id path string true "Workspace ID"
// @Param ticket_id path string true "Ticket ID"
// @Param userId path string true "User ID"
// @Success 200 {object} model.SuccessResponse "Ticket status updated successfully"
// @Failure 500 {object} model.ErrorResponse "Internal server error, failed to update ticket status"
// @Router /messenger/ticket/{status}/{id}/{ticket_id} [put]
func (mc *MessengerController) CloseTicket(c echo.Context) error {
	status, ticketId, workspaceId := c.Param("status"), c.Param("ticket_id"), c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	if err := mc.messengerService.UpdateTicketStatus(userId, ticketId, workspaceId, status); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "ticket status updated successfully"})
}

func (mc *MessengerController) SetUpTelegramClients() error {
	if err := mc.messengerService.SetUpTelegramClients(); err != nil {
		return err
	}
	return nil
}
