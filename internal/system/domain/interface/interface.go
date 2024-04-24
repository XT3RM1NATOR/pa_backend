package _interface

import (
	"github.com/Point-AI/backend/internal/system/infrastructure/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SystemService interface {
	CreateWorkspace(logo []byte, team map[string]string, ownerId primitive.ObjectID, workspaceId, name string, teams []string) error
	LeaveWorkspace(WorkspaceId string, userId primitive.ObjectID) error
	GetAllWorkspaces(userId primitive.ObjectID) ([]model.Workspace, error)
	AddWorkspaceMembers(userId primitive.ObjectID, team map[string]string, WorkspaceId string) error
	DeleteWorkspaceMember(userId primitive.ObjectID, WorkspaceId, memberEmail string) error
	DeleteWorkspaceById(workspaceId string, userId primitive.ObjectID) error
	GetWorkspaceById(WorkspaceId string, userId primitive.ObjectID) (model.Workspace, error)
	UpdateWorkspace(userId primitive.ObjectID, newLogo []byte, WorkspaceId, newWorkspaceId, newName string) error
	UpdateWorkspaceMembers(userId primitive.ObjectID, team map[string]string, WorkspaceId string) error
	GetUserProfiles(WorkspaceId string, userId primitive.ObjectID) ([]model.User, error)
	UpdateWorkspacePendingStatus(userId primitive.ObjectID, workspaceId string, status bool) error
	AddTeamsMember(userId primitive.ObjectID, memberEmail, teamName, workspaceId string) error
	UpdateMemberStatus(userId primitive.ObjectID, status string, workspaceId string) error
	SetFirstTeam(userId primitive.ObjectID, teamName, workspaceId string) error
}

type EmailService interface {
	SendInvitationEmail(recipientEmail, confirmationLink string) error
}
