package controller

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/system/delivery/model"
	_interface "github.com/Point-AI/backend/internal/system/domain/interface"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"strconv"
)

type SystemController struct {
	systemService _interface.SystemService
	config        *config.Config
}

func NewSystemController(systemService _interface.SystemService, cfg *config.Config) *SystemController {
	return &SystemController{
		systemService: systemService,
		config:        cfg,
	}
}

// CreateWorkspace creates a new Workspace.
// @Summary Creates a new Workspace.
// @Tags System
// @Accept json
// @Produce json
// @Param request body CreateWorkspaceRequest true "Workspace details"
// @Success 201 {object} model.SuccessResponse "Workspace added successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/workspace [post]
func (sc *SystemController) CreateWorkspace(c echo.Context) error {
	var request model.CreateWorkspaceRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	ownerId := c.Request().Context().Value("userId").(primitive.ObjectID)
	if err := sc.systemService.CreateWorkspace(request.Logo, request.Team, ownerId, request.WorkspaceID, request.Name); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, model.SuccessResponse{Message: "Workspace added successfully"})
}

// LeaveWorkspace removes user from a Workspace.
// @Summary Removes user from a Workspace.
// @Tags System
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 200 {object} model.SuccessResponse "Workspace left successfully"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/workspace/leave/{id} [delete]
func (sc *SystemController) LeaveWorkspace(c echo.Context) error {
	workspaceId := c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	if err := sc.systemService.LeaveWorkspace(workspaceId, userId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "workspace left successfully"})
}

// GetWorkspaceByID retrieves Workspace details by WorkspaceId.
// @Summary Retrieves Workspace details by ID.
// @Tags System
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 200 {object} model.WorkspaceResponse "Workspace details"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/workspace/{id} [get]
func (sc *SystemController) GetWorkspaceByID(c echo.Context) error {
	workspaceID := c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	workspace, err := sc.systemService.GetWorkspaceById(workspaceID, userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.WorkspaceResponse{
		Name:        workspace.Name,
		Logo:        workspace.Logo,
		WorkspaceID: workspace.WorkspaceId,
	})
}

// GetAllWorkspaces retrieves all Workspaces for a user.
// @Summary Retrieves all Workspaces for a user.
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {array} model.WorkspaceResponse "List of Workspaces"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/workspace [get]
func (sc *SystemController) GetAllWorkspaces(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	workspaces, err := sc.systemService.GetAllWorkspaces(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	var responseWorkspaces []model.WorkspaceResponse
	for _, workspace := range workspaces {
		responseWorkspace := model.WorkspaceResponse{
			Name: workspace.Name,
			Logo: workspace.Logo,
			//Team:      workspace.Team,
			WorkspaceID: workspace.WorkspaceId,
		}
		responseWorkspaces = append(responseWorkspaces, responseWorkspace)
	}

	return c.JSON(http.StatusOK, responseWorkspaces)
}

// UpdateWorkspace updates Workspace details.
// @Summary Updates Workspace details.
// @Tags System
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Param request body UpdateWorkspaceRequest true "Updated Workspace details"
// @Success 200 {object} model.SuccessResponse "Workspace updated successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/workspace/{id} [put]
func (sc *SystemController) UpdateWorkspace(c echo.Context) error {
	workspaceId := c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	var request model.UpdateWorkspaceRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	if err := sc.systemService.UpdateWorkspace(userId, request.Logo, workspaceId, request.WorkspaceID, request.Name); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "Workspace updated successfully"})
}

// AddWorkspaceMembers adds members to a Workspace.
// @Summary Adds members to a Workspace.
// @Tags System
// @Accept json
// @Produce json
// @Param request body AddWorkspaceMemberRequest true "Member details"
// @Success 200 {object} model.SuccessResponse "Users added successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/workspace/member [post]
func (sc *SystemController) AddWorkspaceMembers(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	var request model.AddWorkspaceMemberRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	if err := sc.systemService.AddWorkspaceMembers(userId, request.Team, request.WorkspaceId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "users added successfully"})
}

// UpdateWorkspaceMember updates Workspace members.
// @Summary Updates Workspace members.
// @Tags System
// @Accept json
// @Produce json
// @Param request body UpdateWorkspaceMemberRequest true "Updated member details"
// @Success 200 {object} model.SuccessResponse "Users updated successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/workspace/update [put]
func (sc *SystemController) UpdateWorkspaceMember(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	var request model.UpdateWorkspaceMemberRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	if err := sc.systemService.UpdateWorkspaceMembers(userId, request.Team, request.WorkspaceId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "users updated successfully"})
}

// DeleteWorkspaceMember removes a member from a Workspace.
// @Summary Removes a member from a Workspace.
// @Tags System
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Param email path string true "Member email"
// @Success 200 {object} model.SuccessResponse "Member removed successfully"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/workspace/member/{id}/{email} [delete]
func (sc *SystemController) DeleteWorkspaceMember(c echo.Context) error {
	memberEmail, workspaceId := c.Param("email"), c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	if err := sc.systemService.DeleteWorkspaceMember(userId, workspaceId, memberEmail); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "member removed successfully"})
}

// DeleteWorkspaceByID removes a Workspace by ID.
// @Summary Removes a Workspace by ID.
// @Tags System
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 200 {object} model.SuccessResponse "Workspace deleted successfully"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/workspace/workspace/{id} [delete]
func (sc *SystemController) DeleteWorkspaceByID(c echo.Context) error {
	workspaceId := c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	if err := sc.systemService.DeleteWorkspaceByID(workspaceId, userId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "Workspace deleted successfully"})
}

// GetUserProfiles Returns users in the Workspace.
// @Summary Returns users in the Workspace.
// @Tags System
// @Accept json
// @Produce json
// @Param id path string true "Workspace id"
// @Success 200 {object} model.SuccessResponse "Workspace deleted successfully"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/workspace/members/{id} [get]
func (sc *SystemController) GetUserProfiles(c echo.Context) error {
	workspaceId := c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	users, err := sc.systemService.GetUserProfiles(workspaceId, userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	var userResponses []model.UserResponse
	for _, user := range users {
		userResponse := model.UserResponse{
			Email:    user.Email,
			FullName: user.FullName,
			Role:     user.Role,
			Logo:     user.Logo,
		}
		userResponses = append(userResponses, userResponse)
	}

	return c.JSON(http.StatusOK, userResponses)
}

// UpdateWorkspacePendingStatus Returns users in the Workspace.
// @Summary Returns users in the Workspace.
// @Tags System
// @Accept json
// @Produce json
// @Param id path string true "Workspace id"
// @Param status path bool true "Status of invite"
// @Success 200 {object} model.SuccessResponse "Workspace deleted successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/workspace/{id}/{status} [put]
func (sc *SystemController) UpdateWorkspacePendingStatus(c echo.Context) error {
	statusStr, workspaceId := c.Param("status"), c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	status, err := strconv.ParseBool(statusStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid status"})
	}

	if err := sc.systemService.UpdateWorkspacePendingStatus(userId, workspaceId, status); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "status updated successfully"})
}
