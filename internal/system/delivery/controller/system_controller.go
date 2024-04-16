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

func (sc *SystemController) LeaveProject(c echo.Context) error {
	projectID := c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	if err := sc.systemService.LeaveProject(projectID, userId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "project left successfully"})
}

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
		Team:      project.Team,
		ProjectID: project.ProjectID,
	})
}
func (sc *SystemController) GetAllProjects(c echo.Context) error {
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	projects, err := sc.systemService.GetAllProjects(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	var responseProjects []model.ProjectResponse
	for _, project := range projects {
		responseProject := model.ProjectResponse{
			Name:      project.Name,
			Logo:      project.Logo,
			Team:      project.Team,
			ProjectID: project.ProjectID,
		}
		responseProjects = append(responseProjects, responseProject)
	}

	return c.JSON(http.StatusOK, responseProjects)
}

func (sc *SystemController) UpdateProject(c echo.Context) error {
	projectID := c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	var request model.UpdateProjectRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	if err := sc.systemService.UpdateProject(userId, projectID, request.Logo, request.ProjectID, request.Name); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "project updated successfully"})
}

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

func (sc *SystemController) DeleteProjectMember(c echo.Context) error {
	memberEmail := c.Param("email")
	projectId := c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	if err := sc.systemService.DeleteProjectMember(userId, projectId, memberEmail); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "member removed successfully"})
}

func (sc *SystemController) DeleteProjectByID(c echo.Context) error {
	projectID := c.Param("id")
	userId := c.Request().Context().Value("userId").(primitive.ObjectID)

	if err := sc.systemService.DeleteProjectByID(projectID, userId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "project deleted successfully"})
}
