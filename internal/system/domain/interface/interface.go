package _interface

import (
	"github.com/Point-AI/backend/internal/system/delivery/model"
	infrastructureModel "github.com/Point-AI/backend/internal/system/infrastructure/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SystemService interface {
	CreateWorkspace(logo []byte, team map[string]string, ownerId primitive.ObjectID, workspaceId, name string, teams []string) error
	LeaveWorkspace(WorkspaceId string, userId primitive.ObjectID) error
	GetAllWorkspaces(userId primitive.ObjectID) ([]infrastructureModel.Workspace, error)
	AddWorkspaceMembers(userId primitive.ObjectID, team map[string]string, WorkspaceId string) error
	DeleteWorkspaceMember(userId primitive.ObjectID, WorkspaceId, memberEmail string) error
	DeleteWorkspaceById(workspaceId string, userId primitive.ObjectID) error
	GetWorkspaceById(WorkspaceId string, userId primitive.ObjectID) (infrastructureModel.Workspace, error)
	UpdateWorkspace(userId primitive.ObjectID, newLogo []byte, WorkspaceId, newWorkspaceId, newName string) error
	UpdateWorkspaceMembers(userId primitive.ObjectID, team map[string]string, WorkspaceId string) error
	GetUserProfiles(WorkspaceId string, userId primitive.ObjectID) ([]infrastructureModel.User, error)
	UpdateWorkspacePendingStatus(userId primitive.ObjectID, workspaceId string, status bool) error
	AddTeamsMembers(userId primitive.ObjectID, members map[string]string, teamId, workspaceId string) error
	SetFirstTeam(userId primitive.ObjectID, teamId, workspaceId string) error
	EditFolders(userId primitive.ObjectID, workspaceId string, folders map[string][]string) error
	RegisterTelegramIntegration(userId primitive.ObjectID, workspaceId, stage, value string) (int, error)
	GetAllTeams(userId primitive.ObjectID, workspaceId string) ([]model.TeamResponse, error)
	GetAllFolders(userId primitive.ObjectID, workspaceId string) (map[string][]string, error, int)
	CreateTeam(userId primitive.ObjectID, workspaceId, teamName string, members map[string]string, logo []byte) error
	DeleteTeam(userId primitive.ObjectID, workspaceId, teamName string) error
	UpdateTeam(userId primitive.ObjectID, newLogo []byte, workspaceId, newTeamName, teamId string) error
	GetAllUsers(userId primitive.ObjectID, workspaceId string) ([]model.UserResponse, error)
}

type EmailService interface {
	SendInvitationEmail(recipientEmail, confirmationLink string) error
	SendWorkspaceInvitationEmail(recipientEmail, inviteLink string) error
}

type FileService interface {
	SaveFile(filename string, content []byte) error
	LoadFile(filename string) ([]byte, error)
	UpdateFileName(oldName, newName string) error
	UpdateFile(newFileBytes []byte, fileName string) error
}
