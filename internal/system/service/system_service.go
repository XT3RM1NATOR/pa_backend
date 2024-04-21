package service

import (
	"errors"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/system/domain/entity"
	_interface "github.com/Point-AI/backend/internal/system/domain/interface"
	"github.com/Point-AI/backend/internal/system/infrastructure/model"
	infrastructureInterface "github.com/Point-AI/backend/internal/system/service/interface"
	"github.com/Point-AI/backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SystemServiceImpl struct {
	systemRepo    infrastructureInterface.SystemRepository
	storageClient infrastructureInterface.StorageClient
	emailService  _interface.EmailService
	config        *config.Config
}

func NewSystemServiceImpl(cfg *config.Config, storageClient infrastructureInterface.StorageClient, systemRepo infrastructureInterface.SystemRepository, emailService _interface.EmailService) *SystemServiceImpl {
	return &SystemServiceImpl{
		systemRepo:    systemRepo,
		emailService:  emailService,
		storageClient: storageClient,
		config:        cfg,
	}
}

func (ss *SystemServiceImpl) CreateWorkspace(logo []byte, team map[string]string, ownerId primitive.ObjectID, workspaceId, name string) error {
	if err := utils.ValidateWorkspaceId(workspaceId); err != nil {
		return err
	}

	if logo != nil {
		if err := utils.ValidatePhoto(logo); err != nil {
			return err
		}
	}

	teamRoles, err := utils.ValidateTeamRoles(team)
	if err != nil {
		return err
	}

	_, err = ss.systemRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if errors.Is(err, mongo.ErrNoDocuments) {
		if err := ss.systemRepo.CreateWorkspace(ownerId, &teamRoles, workspaceId, name); err != nil {
			return err
		}

		go ss.storageClient.SaveFile(logo, ss.config.MinIo.BucketName, name)

		for email, _ := range teamRoles {
			id, err := ss.systemRepo.FindUserByEmail(email)
			if errors.Is(err, mongo.ErrNoDocuments) {
				go ss.emailService.SendInvitationEmail(email, ss.config.Website.BaseURL+"/auth/signup")
			} else if err == nil {
				if err = ss.systemRepo.AddPendingInviteToUser(id, workspaceId); err != nil {

				}
			}
		}

		return nil
	} else if err != nil {
		return err
	}

	return errors.New("workspace with this id already exists")
}

func (ss *SystemServiceImpl) LeaveWorkspace(WorkspaceId string, userId primitive.ObjectID) error {
	workspace, err := ss.systemRepo.FindWorkspaceByWorkspaceId(WorkspaceId)
	if err != nil {
		return err
	}

	if err := ss.systemRepo.RemoveUserFromWorkspace(workspace, userId); err != nil {
		return err
	}

	return nil
}

// GetWorkspaceById TODO: update function not to return team
func (ss *SystemServiceImpl) GetWorkspaceById(workspaceId string, userId primitive.ObjectID) (model.Workspace, error) {
	workspace, err := ss.systemRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return model.Workspace{}, err
	}

	if _, exists := workspace.Team[userId]; !exists {
		return model.Workspace{}, errors.New("user is not in the Workspace")
	}

	fmtWorkspace, err := ss.formatWorkspaces([]entity.Workspace{*workspace})
	if err != nil {
		return model.Workspace{}, err
	}

	return fmtWorkspace[0], nil
}

func (ss *SystemServiceImpl) GetAllWorkspaces(userId primitive.ObjectID) ([]model.Workspace, error) {
	workspaces, err := ss.systemRepo.FindWorkspacesByUser(userId)
	if err != nil {
		return []model.Workspace{}, err
	}

	fmtWorkspaces, err := ss.formatWorkspaces(*workspaces)
	if err != nil {
		return []model.Workspace{}, err
	}
	return fmtWorkspaces, err
}

func (ss *SystemServiceImpl) UpdateWorkspace(userId primitive.ObjectID, newLogo []byte, workspaceId, newWorkspaceId, newName string) error {
	workspace, err := ss.systemRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	if ss.isOwner(workspace.Team[userId]) || ss.isAdmin(workspace.Team[userId]) {
		if newWorkspaceId != "" {
			if err := utils.ValidateWorkspaceId(workspaceId); err != nil {
				return err
			}
			if err := ss.storageClient.UpdateFileName(workspace.WorkspaceId, newWorkspaceId, ss.config.MinIo.BucketName); err != nil {
				return err
			}
			workspace.WorkspaceId = newWorkspaceId
		}

		if newLogo != nil {
			if err := utils.ValidatePhoto(newLogo); err != nil {
				return err
			}
			if err := ss.storageClient.UpdateFile(newLogo, workspace.WorkspaceId, ss.config.MinIo.BucketName); err != nil {
				return err
			}
		}

		if newName != "" {
			workspace.Name = newName
		}

		if err := ss.systemRepo.UpdateWorkspace(workspace); err != nil {
			return err
		}
		return nil
	}
	return errors.New("unauthorized to make the changes")
}

func (ss *SystemServiceImpl) AddWorkspaceMembers(userId primitive.ObjectID, team map[string]string, workspaceId string) error {
	workspace, err := ss.systemRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	if ss.isAdmin(workspace.Team[userId]) || ss.isOwner(workspace.Team[userId]) {
		teamRoles, err := ss.systemRepo.ValidateTeam(&team, userId)
		if err != nil {
			return err
		}

		if err := ss.systemRepo.AddUsersToWorkspace(workspace, teamRoles); err != nil {
			return err
		}
	}

	return nil
}

func (ss *SystemServiceImpl) UpdateWorkspaceMembers(userId primitive.ObjectID, team map[string]string, workspaceId string) error {
	workspace, err := ss.systemRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	if ss.isAdmin(workspace.Team[userId]) || ss.isOwner(workspace.Team[userId]) {
		teamRoles, err := ss.systemRepo.ValidateTeam(&team, userId)
		if err != nil {
			return err
		}

		if err := ss.systemRepo.UpdateUsersInWorkspace(workspace, teamRoles); err != nil {
			return err
		}
	}

	return nil
}

func (ss *SystemServiceImpl) DeleteWorkspaceMember(userId primitive.ObjectID, workspaceId, memberEmail string) error {
	workspace, err := ss.systemRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	if ss.isAdmin(workspace.Team[userId]) || ss.isOwner(workspace.Team[userId]) {
		userId, err := ss.systemRepo.FindUserByEmail(memberEmail)
		if err != nil {
			return err
		}

		if err := ss.systemRepo.RemoveUserFromWorkspace(workspace, userId); err != nil {
			return err
		}
	}

	return nil
}

func (ss *SystemServiceImpl) DeleteWorkspaceByID(workspaceId string, userId primitive.ObjectID) error {
	workspace, err := ss.systemRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	if ss.isOwner(workspace.Team[userId]) {
		if err := ss.systemRepo.DeleteWorkspace(workspace.Id); err != nil {
			return err
		}
	}

	return errors.New("user does not have a valid permission")
}

func (ss *SystemServiceImpl) GetUserProfiles(workspaceId string, userId primitive.ObjectID) ([]model.User, error) {
	workspace, err := ss.systemRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return nil, err
	}

	if _, exists := workspace.Team[userId]; exists {
		users, err := ss.systemRepo.GetUserProfiles(*workspace)
		if err != nil {
			return nil, err
		}

		for _, user := range *users {
			user.Logo, _ = ss.storageClient.LoadFile(user.Email, ss.config.MinIo.BucketName)
		}

		return *users, nil
	}

	return nil, errors.New("user does not have a valid permission")
}

func (ss *SystemServiceImpl) formatWorkspaces(workspaces []entity.Workspace) ([]model.Workspace, error) {
	formattedWorkspaces := make([]model.Workspace, len(workspaces))
	for i, p := range workspaces {
		logo, _ := ss.storageClient.LoadFile(p.WorkspaceId, ss.config.MinIo.BucketName)
		team, _ := ss.systemRepo.FormatTeam(&p.Team)

		formattedWorkspace := model.Workspace{
			Name:        p.Name,
			WorkspaceId: p.WorkspaceId,
			Team:        *team,
			Logo:        logo,
		}

		formattedWorkspaces[i] = formattedWorkspace
	}

	return formattedWorkspaces, nil
}

func (ss *SystemServiceImpl) UpdateWorkspacePendingStatus(userId primitive.ObjectID, workspaceId string, status bool) error {
	if err := ss.systemRepo.ClearPendingStatus(userId, workspaceId); err != nil {
		return err
	}

	if err := ss.systemRepo.UpdateWorkspaceUserStatus(userId, workspaceId, status); err != nil {
		return err
	}
	return nil
}

func (ss *SystemServiceImpl) isAdmin(userRole entity.WorkspaceRole) bool {
	return userRole == entity.RoleAdmin
}

func (ss *SystemServiceImpl) isOwner(userRole entity.WorkspaceRole) bool {
	return userRole == entity.RoleOwner
}
