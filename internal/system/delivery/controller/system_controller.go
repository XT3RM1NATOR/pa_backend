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

func NewSystemController(cfg *config.Config, systemService _interface.SystemService) *SystemController {
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
// @Param request body model.CreateWorkspaceRequest true "Workspace details"
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
	if err := sc.systemService.CreateWorkspace(request.Logo, ownerId, request.WorkspaceId, request.Name); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, model.SuccessResponse{Message: "workspace added successfully"})
}

// AddTeamsMembers adds a new member to a team.
// @Summary Adds a new member to a team.
// @Tags System
// @Accept json
// @Produce json
// @Param request body model.AddTeamMembersRequest true "Team member details"
// @Param userId path string true "User ID"
// @Success 201 {object} model.SuccessResponse "User added to the team successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request, unable to parse the request body"
// @Failure 500 {object} model.ErrorResponse "Internal server error, failed to add the team member"
// @Router /system/teams [post]
func (sc *SystemController) AddTeamsMembers(c echo.Context) error {
	var request model.AddTeamMembersRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	userId := c.Request().Context().Value("userId").(primitive.ObjectID)
	if err := sc.systemService.AddTeamsMembers(userId, request.Members, request.TeamId, request.WorkspaceId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, model.SuccessResponse{Message: "user added to the team"})
}

// SetFirstTeam sets the first team for a user in a workspace.
// @Summary Sets the first team for a user in a specific workspace.
// @Tags System
// @Accept json
// @Produce json
// @Param name path string true "Team name"
// @Param id path string true "Workspace ID"
// @Param userId path string true "User ID"
// @Success 201 {object} model.SuccessResponse "First team set successfully"
// @Failure 500 {object} model.ErrorResponse "Internal server error, failed to set the first team"
// @Router /system/teams/{id}/{name} [post]
func (sc *SystemController) SetFirstTeam(c echo.Context) error {
	teamId, workspaceId := c.Param("team_id"), c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	if err := sc.systemService.SetFirstTeam(userId, teamId, workspaceId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, model.SuccessResponse{Message: "first team is set"})
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

// GetWorkspaceById retrieves Workspace details by WorkspaceId.
// @Summary Retrieves Workspace details by ID.
// @Tags System
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 200 {object} model.WorkspaceResponse "Workspace details"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/workspace/{id} [get]
func (sc *SystemController) GetWorkspaceById(c echo.Context) error {
	workspaceID := c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	workspace, err := sc.systemService.GetWorkspaceById(workspaceID, userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.WorkspaceResponse{
		Name:        workspace.Name,
		Logo:        workspace.Logo,
		WorkspaceId: workspace.WorkspaceId,
	})
}

// GetAllWorkspaces retrieves all Workspaces for a user.
// @Summary Retrieves all Workspaces for a user.
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {array} model.WorkspaceResponse "List of Workspaces"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/workspace [put]
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
			WorkspaceId: workspace.WorkspaceId,
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
// @Param id path string true "Workspace id"
// @Param request body model.UpdateWorkspaceRequest true "Updated Workspace details"
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

	if err := sc.systemService.UpdateWorkspace(userId, request.Logo, workspaceId, request.WorkspaceId, request.Name); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "Workspace updated successfully"})
}

// RegisterTelegramIntegration handles the registration of Telegram integration for authentication.
// @Summary Registers or progresses Telegram integration for authentication.
// @Description This endpoint handles the various stages of Telegram authentication, including phone number submission, verification code checking, and two-factor authentication password verification.
// @Tags System
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Param set query string true "Stage of the Telegram authentication process ('phone', 'code', or 'password')"
// @Param value query string true "Value required for the current stage (phone number, verification code, or password)"
// @Success 200 {object} model.SuccessResponse "Telegram authentication processed successfully; stage complete."
// @Success 202 {object} model.SuccessResponse "Verification code accepted but password is required for two-factor authentication."
// @Failure 500 {object} model.ErrorResponse "Internal server error occurred while processing the Telegram integration."
// @Router /system/integrations/telegram/{id} [get]
func (sc *SystemController) RegisterTelegramIntegration(c echo.Context) error {
	workspaceId, stage, value := c.Param("id"), c.QueryParam("set"), c.QueryParam("value")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	statusCode, err := sc.systemService.RegisterTelegramIntegration(userId, workspaceId, stage, value)
	if err != nil {
		return c.JSON(statusCode, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(statusCode, model.SuccessResponse{Message: "telegram auth status updated successfully"})
}

// AddWorkspaceMembers adds members to a Workspace.
// @Summary Adds members to a Workspace.
// @Tags System
// @Accept json
// @Produce json
// @Param request body model.AddWorkspaceMemberRequest true "Member details"
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
// @Param request body model.UpdateWorkspaceMemberRequest true "Updated member details"
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

// DeleteWorkspaceById removes a Workspace by id.
// @Summary Removes a Workspace by id.
// @Tags System
// @Accept json
// @Produce json
// @Param id path string true "Workspace id"
// @Success 200 {object} model.SuccessResponse "Workspace deleted successfully"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/workspace/workspace/{id} [delete]
func (sc *SystemController) DeleteWorkspaceById(c echo.Context) error {
	workspaceId := c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	if err := sc.systemService.DeleteWorkspaceById(workspaceId, userId); err != nil {
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

// AddFolders Returns users in the Workspace.
// @Summary Returns users in the Workspace.
// @Tags System
// @Accept json
// @Produce json
// @Param request body model.EditFoldersRequest true "folders"
// @Success 200 {object} model.SuccessResponse "folders updated successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/workspace/folders [post]
func (sc *SystemController) AddFolders(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)
	var request model.EditFoldersRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	if err := sc.systemService.EditFolders(userId, request.WorkspaceId, request.Folders); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "folders updated successfully"})
}

func (sc *SystemController) CreateTeam(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)
	var request model.CreateTeamRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	if err := sc.systemService.CreateTeam(userId, request.WorkspaceId, request.TeamName, request.Members, request.Logo); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "new team created successfully"})
}

func (sc *SystemController) UpdateTeam(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)
	var request model.UpdateTeamRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	if err := sc.systemService.UpdateTeam(userId, request.Logo, request.WorkspaceId, request.NewTeamName, request.TeamId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "team updated successfully"})
}

func (sc *SystemController) DeleteTeam(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)
	workspaceId, teamId := c.Param("id"), c.Param("team_id")

	if err := sc.systemService.DeleteTeam(userId, workspaceId, teamId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "new team created successfully"})
}

func (sc *SystemController) GetAllTeams(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)
	workspaceId := c.Param("id")

	teams, code, err := sc.systemService.GetAllTeams(userId, workspaceId)
	if err != nil {
		return c.JSON(code, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(code, teams)
}

func (sc *SystemController) GetAllFolders(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)
	workspaceId := c.Param("id")

	folders, err, code := sc.systemService.GetAllFolders(userId, workspaceId)
	if err != nil {
		return c.JSON(code, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(code, model.FoldersResponse{Folders: folders})
}

func (sc *SystemController) GetAllUsers(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)
	workspaceId, teamId := c.Param("id"), c.QueryParam("team_id")

	users, err := sc.systemService.GetAllUsers(userId, workspaceId, teamId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, users)
}
