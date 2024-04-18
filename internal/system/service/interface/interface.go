package infrastructureInterface

import (
	"github.com/Point-AI/backend/internal/system/domain/entity"
	"github.com/Point-AI/backend/internal/system/infrastructure/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StorageClient interface {
	SaveFile(fileBytes []byte, bucketName, objectName string) error
	LoadFile(fileName, bucketName string) ([]byte, error)
	UpdateFileName(oldName, newName string, bucketName string) error
	UpdateFile(newFileBytes []byte, fileName string, bucketName string) error
}

type SystemRepository interface {
	ValidateTeam(team map[string]string, ownerId primitive.ObjectID) (map[primitive.ObjectID]entity.ProjectRole, error)
	CreateProject(team map[primitive.ObjectID]entity.ProjectRole, projectId, name string) error
	RemoveUserFromProject(project entity.Project, userId primitive.ObjectID) error
	FindProjectByProjectId(projectId string) (entity.Project, error)
	DeleteProject(id primitive.ObjectID) error
	FindProjectsByUser(userID primitive.ObjectID) ([]entity.Project, error)
	FindUserByEmail(email string) (primitive.ObjectID, error)
	FindUserById(userID primitive.ObjectID) (string, error)
	AddUsersToProject(project entity.Project, teamRoles map[primitive.ObjectID]entity.ProjectRole) error
	UpdateUsersInProject(project entity.Project, teamRoles map[primitive.ObjectID]entity.ProjectRole) error
	UpdateProject(project entity.Project) error
	FormatTeam(team map[primitive.ObjectID]entity.ProjectRole) (map[string]string, error)
	GetUserProfiles(project entity.Project) ([]model.User, error)
}
