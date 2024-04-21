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
	ValidateTeam(team map[string]string, ownerId primitive.ObjectID) (map[primitive.ObjectID]entity.WorkspaceRole, error)
	CreateWorkspace(team map[primitive.ObjectID]entity.WorkspaceRole, WorkspaceId, name string) error
	RemoveUserFromWorkspace(Workspace entity.Workspace, userId primitive.ObjectID) error
	FindWorkspaceByWorkspaceId(WorkspaceId string) (entity.Workspace, error)
	DeleteWorkspace(id primitive.ObjectID) error
	FindWorkspacesByUser(userID primitive.ObjectID) ([]entity.Workspace, error)
	FindUserByEmail(email string) (primitive.ObjectID, error)
	FindUserById(userID primitive.ObjectID) (string, error)
	AddUsersToWorkspace(Workspace entity.Workspace, teamRoles map[primitive.ObjectID]entity.WorkspaceRole) error
	UpdateUsersInWorkspace(Workspace entity.Workspace, teamRoles map[primitive.ObjectID]entity.WorkspaceRole) error
	UpdateWorkspace(Workspace entity.Workspace) error
	FormatTeam(team map[primitive.ObjectID]entity.WorkspaceRole) (map[string]string, error)
	GetUserProfiles(Workspace entity.Workspace) ([]model.User, error)
}
