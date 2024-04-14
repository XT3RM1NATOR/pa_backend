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

	ownerId := c.Request().Context().Value("userID").(primitive.ObjectID)
	if err := sc.systemService.CreateProject(request.Logo, request.Team, ownerId, request.ProjectID, request.Name); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, model.SuccessResponse{Message: "project added successfully"})
}

func (sc *SystemController) LeaveProject(c echo.Context) error {
	var request model.LeaveProjectRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	}

	userId := c.Request().Context().Value("userID").(primitive.ObjectID)
	if err := sc.systemService.LeaveProject(request.ProjectID, userId); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, model.SuccessResponse{Message: "project left successfully"})
}

//func (sc *SystemController) GetProjectByID(c echo.Context) error {
//	var request model.UserRequest
//	if err := c.Bind(&request); err != nil {
//		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
//	}
//}
//
//func (sc *SystemController) GetAllProjects(c echo.Context) error {
//	var request model.UserRequest
//	if err := c.Bind(&request); err != nil {
//		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
//	}
//}
//
//func (sc *SystemController) UpdateProjectByID(c echo.Context) error {
//	var request model.UserRequest
//	if err := c.Bind(&request); err != nil {
//		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
//	}
//}
//
//func (sc *SystemController) AddProjectMember(c echo.Context) error {
//	var request model.UserRequest
//	if err := c.Bind(&request); err != nil {
//		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
//	}
//}
//
//func (sc *SystemController) UpdateProjectMember(c echo.Context) error {
//	var request model.UserRequest
//	if err := c.Bind(&request); err != nil {
//		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
//	}
//}
//
//func (sc *SystemController) LeaveProject(c echo.Context) error {
//	var request model.UserRequest
//	if err := c.Bind(&request); err != nil {
//		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
//	}
//}
//
//func (sc *SystemController) DeleteProjectMember(c echo.Context) error {
//	var request model.UserRequest
//	if err := c.Bind(&request); err != nil {
//		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
//	}
//}
//
//func (sc *SystemController) DeleteProjectByID(c echo.Context) error {
//	var request model.UserRequest
//	if err := c.Bind(&request); err != nil {
//		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
//	}
//}
