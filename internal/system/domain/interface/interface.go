package _interface

import (
	"github.com/Point-AI/backend/internal/system/infrastructure/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SystemService interface {
	CreateProject(logo []byte, team map[string]string, ownerId primitive.ObjectID, projectId, name string) error
	LeaveProject(projectId string, userId primitive.ObjectID) error
	GetAllProjects(userId primitive.ObjectID) ([]model.Project, error)
	AddProjectMembers(userId primitive.ObjectID, team map[string]string, projectId string) error
	DeleteProjectMember(userId primitive.ObjectID, projectId, memberEmail string) error
	DeleteProjectByID(projectId string, userId primitive.ObjectID) error
	GetProjectById(projectId string, userId primitive.ObjectID) (model.Project, error)
	UpdateProject(userId primitive.ObjectID, newLogo []byte, projectId, newProjectId, newName string) error
	UpdateProjectMembers(userId primitive.ObjectID, team map[string]string, projectId string) error
}
