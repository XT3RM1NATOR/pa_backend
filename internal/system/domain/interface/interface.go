package _interface

import (
	"github.com/Point-AI/backend/internal/system/infrastructure/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SystemService interface {
	CreateProject(logo []byte, team map[string]string, ownerId primitive.ObjectID, projectId, name string) error
	LeaveProject(projectId string, userId primitive.ObjectID) error
	GetProjectByID() error
	GetAllProjects(userId primitive.ObjectID) ([]model.Project, error)
	UpdateProjectByID() error
	AddProjectMembers(userId primitive.ObjectID, team map[string]string, projectId string) error
	UpdateProjectMember() error
	DeleteProjectMember(userId primitive.ObjectID, projectId, memberEmail string) error
	DeleteProjectByID(projectId string, userId primitive.ObjectID) error
}
