package controller

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/system/delivery/model"
	_interface "github.com/Point-AI/backend/internal/system/domain/interface"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
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

// CreateProject creates a new project.
// @Summary Creates a new project.
// @Tags System
// @Accept json
// @Produce json
// @Param request body CreateProjectRequest true "Project details"
// @Success 201 {object} model.SuccessResponse "Project added successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/project [post]
func (sc *SystemController) CreateProject(c echo.Context) error {
	var request model.CreateProjectRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	ownerId := c.Request().Context().Value("userId").(primitive.ObjectID)
	if err := sc.systemService.CreateProject(request.Logo, request.Team, ownerId, request.ProjectID, request.Name); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, model.SuccessResponse{Message: "project added successfully"})
}

// LeaveProject removes user from a project.
// @Summary Removes user from a project.
// @Tags System
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} model.SuccessResponse "Project left successfully"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/project/leave/{id} [delete]
func (sc *SystemController) LeaveProject(c echo.Context) error {
	projectID := c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	if err := sc.systemService.LeaveProject(projectID, userId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "project left successfully"})
}

// GetProjectByID retrieves project details by ProjectId.
// @Summary Retrieves project details by ID.
// @Tags System
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} model.ProjectResponse "Project details"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/project/{id} [get]
func (sc *SystemController) GetProjectByID(c echo.Context) error {
	projectID := c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	project, err := sc.systemService.GetProjectById(projectID, userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.ProjectResponse{
		Name:      project.Name,
		Logo:      project.Logo,
		ProjectID: project.ProjectID,
	})
}

// GetAllProjects retrieves all projects for a user.
// @Summary Retrieves all projects for a user.
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {array} model.ProjectResponse "List of projects"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/project [get]
func (sc *SystemController) GetAllProjects(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	projects, err := sc.systemService.GetAllProjects(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	var responseProjects []model.ProjectResponse
	for _, project := range projects {
		responseProject := model.ProjectResponse{
			Name: project.Name,
			Logo: project.Logo,
			//Team:      project.Team,
			ProjectID: project.ProjectID,
		}
		responseProjects = append(responseProjects, responseProject)
	}

	return c.JSON(http.StatusOK, responseProjects)
}

// UpdateProject updates project details.
// @Summary Updates project details.
// @Tags System
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body UpdateProjectRequest true "Updated project details"
// @Success 200 {object} model.SuccessResponse "Project updated successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/project/{id} [put]
func (sc *SystemController) UpdateProject(c echo.Context) error {
	projectID := c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	var request model.UpdateProjectRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	if err := sc.systemService.UpdateProject(userId, request.Logo, projectID, request.ProjectID, request.Name); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "project updated successfully"})
}

// AddProjectMembers adds members to a project.
// @Summary Adds members to a project.
// @Tags System
// @Accept json
// @Produce json
// @Param request body AddProjectMemberRequest true "Member details"
// @Success 200 {object} model.SuccessResponse "Users added successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/project/member [post]
func (sc *SystemController) AddProjectMembers(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	var request model.AddProjectMemberRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	if err := sc.systemService.AddProjectMembers(userId, request.Team, request.ProjectId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "users added successfully"})
}

// UpdateProjectMember updates project members.
// @Summary Updates project members.
// @Tags System
// @Accept json
// @Produce json
// @Param request body UpdateProjectMemberRequest true "Updated member details"
// @Success 200 {object} model.SuccessResponse "Users updated successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/project/update [put]
func (sc *SystemController) UpdateProjectMember(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	var request model.UpdateProjectMemberRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	if err := sc.systemService.UpdateProjectMembers(userId, request.Team, request.ProjectId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "users updated successfully"})
}

// DeleteProjectMember removes a member from a project.
// @Summary Removes a member from a project.
// @Tags System
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param email path string true "Member email"
// @Success 200 {object} model.SuccessResponse "Member removed successfully"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/project/member/{id}/{email} [delete]
func (sc *SystemController) DeleteProjectMember(c echo.Context) error {
	memberEmail := c.Param("email")
	projectId := c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	if err := sc.systemService.DeleteProjectMember(userId, projectId, memberEmail); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "member removed successfully"})
}

// DeleteProjectByID removes a project by ID.
// @Summary Removes a project by ID.
// @Tags System
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} model.SuccessResponse "Project deleted successfully"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/project/{id} [delete]
func (sc *SystemController) DeleteProjectByID(c echo.Context) error {
	projectID := c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	if err := sc.systemService.DeleteProjectByID(projectID, userId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "project deleted successfully"})
}

// GetUserProfiles Returns users in the project.
// @Summary Returns users in the project.
// @Tags System
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} model.SuccessResponse "Project deleted successfully"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /system/project/{id} [delete]
func (sc *SystemController) GetUserProfiles(c echo.Context) error {
	projectID := c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	users, err := sc.systemService.GetUserProfiles(projectID, userId)
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
