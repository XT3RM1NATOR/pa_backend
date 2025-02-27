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
	ValidateTeam(team map[string]string, ownerId primitive.ObjectID) (map[primitive.ObjectID]entity.WorkspaceRole, map[string]entity.WorkspaceRole, error)
	CreateWorkspace(ownerId primitive.ObjectID, workspaceId, name string) error
	RemoveUserFromWorkspace(workspace *entity.Workspace, userId primitive.ObjectID) error
	FindWorkspaceByWorkspaceId(workspaceId string) (*entity.Workspace, error)
	DeleteWorkspace(id primitive.ObjectID) error
	FindWorkspacesByUser(userId primitive.ObjectID) (*[]entity.Workspace, error)
	FindUserByEmail(email string) (primitive.ObjectID, error)
	FindUserEmailById(userId primitive.ObjectID) (string, error)
	AddUsersToWorkspace(workspace *entity.Workspace, teamRoles map[primitive.ObjectID]entity.WorkspaceRole, pendingTeamRoles map[string]entity.WorkspaceRole) error
	UpdateUsersInWorkspace(workspace *entity.Workspace, teamRoles map[primitive.ObjectID]entity.WorkspaceRole) error
	UpdateWorkspace(workspace *entity.Workspace) error
	FormatTeam(team map[primitive.ObjectID]entity.WorkspaceRole) (map[string]string, error)
	GetUserProfiles(workspace entity.Workspace) (*[]infrastructureModel.User, error)
	AddPendingInviteToUser(userId primitive.ObjectID, projectId string) error
	ClearPendingStatus(userId primitive.ObjectID, workspaceId string) error
	UpdateWorkspaceUserStatus(userId primitive.ObjectID, workspaceId string, status bool) error
	FindUserById(userId primitive.ObjectID) (*entity.User, error)
	ValidateNewTeamUsers(team map[string]string) (map[primitive.ObjectID]entity.WorkspaceRole, map[string]entity.WorkspaceRole, error)
	FindTeamByTeamId(teamId string) (*entity.Team, error)
	InsertNewTeam(team *entity.Team) error
	UpdateTeam(team *entity.Team) error
	FindTeamByTeamIdAndWorkspaceId(teamId string, workspaceId primitive.ObjectID) (*entity.Team, error)
	FindTeamsByWorkspaceId(workspaceId primitive.ObjectID) ([]*entity.Team, error)
	CountChatsByTeamId(teamId primitive.ObjectID) (int, error)
	DeleteTeam(id primitive.ObjectID) error
	UpdateChatTeamIdToNil(teamId primitive.ObjectID) error
	GetTeamNamesByUserId(userId primitive.ObjectID) []entity.Team
}

type EmailClient interface {
	SendEmail(to, subject, body string) error
}
