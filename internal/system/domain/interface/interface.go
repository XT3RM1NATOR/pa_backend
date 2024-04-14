package _interface

import "go.mongodb.org/mongo-driver/bson/primitive"

type SystemService interface {
	CreateProject(logo []byte, team []string, ownerId primitive.ObjectID, projectId, name string) error
	LeaveProject(projectId string, userId primitive.ObjectID) error
	GetProjectByID() error
	GetAllProjects() error
	UpdateProjectByID() error
	AddProjectMember() error
	UpdateProjectMember() error
	DeleteProjectMember() error
	DeleteProjectByID() error
}
