package service

import (
	"errors"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/system/delivery/model"
	"github.com/Point-AI/backend/internal/system/domain/entity"
	_interface "github.com/Point-AI/backend/internal/system/domain/interface"
	"github.com/Point-AI/backend/internal/system/infrastructure/model"
	infrastructureInterface "github.com/Point-AI/backend/internal/system/service/interface"
	"github.com/Point-AI/backend/utils"
	"github.com/go-resty/resty/v2"
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

func (ss *SystemServiceImpl) CreateWorkspace(logo []byte, team map[string]string, ownerId primitive.ObjectID, workspaceId, name string, teams []string) error {
	if err := utils.ValidateWorkspaceId(workspaceId); err != nil {
		return err
	}

	var teamRoles map[string]entity.WorkspaceRole
	if logo != nil {
		if err := utils.ValidatePhoto(logo); err != nil {
			return err
		}
	}
	if team != nil {
		roles, err := utils.ValidateTeamRoles(team)
		if err != nil {
			return err
		}
		teamRoles = roles
	}
	if teams != nil {
		if err := utils.ValidateTeamNames(teams); err != nil {
			return err
		}
	}

	_, err := ss.systemRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if errors.Is(err, mongo.ErrNoDocuments) {
		if err := ss.systemRepo.CreateWorkspace(ownerId, teamRoles, workspaceId, name, teams); err != nil {
			return err
		}

		go ss.storageClient.SaveFile(logo, ss.config.MinIo.BucketName, name)

		for email, _ := range teamRoles {
			id, err := ss.systemRepo.FindUserByEmail(email)
			if errors.Is(err, mongo.ErrNoDocuments) {
				go ss.emailService.SendInvitationEmail(email, ss.config.Website.BaseURL+"/auth/signup")
			} else if err == nil {
				go ss.systemRepo.AddPendingInviteToUser(id, workspaceId)
			}
		}

		return nil
	} else if err != nil {
		return err
	}

	return errors.New("workspace with this id already exists")
}

func (ss *SystemServiceImpl) LeaveWorkspace(workspaceId string, userId primitive.ObjectID) error {
	workspace, err := ss.systemRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	if err := ss.systemRepo.RemoveUserFromWorkspace(workspace, userId); err != nil {
		return err
	}

	return nil
}

func (ss *SystemServiceImpl) SetFirstTeam(userId primitive.ObjectID, teamName, workspaceId string) error {
	workspace, err := ss.systemRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	if ss.isAdmin(workspace.Team[userId]) || ss.isOwner(workspace.Team[userId]) {
		if _, exists := workspace.InternalTeams[teamName]; !exists {
			return errors.New("team not found")
		}
		workspace.FirstTeam = teamName

		err := ss.systemRepo.UpdateWorkspace(workspace)
		if err != nil {
			return err
		}
	}

	return errors.New("unauthorised")
}

func (ss *SystemServiceImpl) UpdateMemberStatus(userId primitive.ObjectID, status, workspaceId string) error {
	workspace, err := ss.systemRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	for _, team := range workspace.InternalTeams {
		if _, exists := team[userId]; exists {
			switch entity.UserStatus(status) {
			case entity.StatusAvailable, entity.StatusOffline, entity.StatusBusy:
				team[userId] = entity.UserStatus(status)
			default:
				return errors.New("invalid status")
			}
		}
	}

	err = ss.systemRepo.UpdateWorkspace(workspace)
	if err != nil {
		return err
	}

	return nil
}

func (ss *SystemServiceImpl) AddTeamsMember(userId primitive.ObjectID, memberEmail, teamName, workspaceId string) error {
	workspace, err := ss.systemRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	if ss.isAdmin(workspace.Team[userId]) || ss.isOwner(workspace.Team[userId]) {
		teamMembers, ok := workspace.InternalTeams[teamName]
		if !ok {
			return errors.New("team not found")
		}

		id, err := ss.systemRepo.FindUserByEmail(memberEmail)
		if err != nil {
			return err
		}

		for _, team := range workspace.InternalTeams {
			if _, exists := team[id]; exists {
				return errors.New("user is already a member of another team")
			}
		}

		teamMembers[id] = entity.StatusOffline
		err = ss.systemRepo.UpdateWorkspace(workspace)
		if err != nil {
			return err
		}
		return nil
	}

	return errors.New("unauthorised")
}

// GetWorkspaceById TODO: update function not to return team
func (ss *SystemServiceImpl) GetWorkspaceById(workspaceId string, userId primitive.ObjectID) (infrastructureModel.Workspace, error) {
	workspace, err := ss.systemRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return infrastructureModel.Workspace{}, err
	}

	if _, exists := workspace.Team[userId]; !exists {
		return infrastructureModel.Workspace{}, errors.New("user is not in the Workspace")
	}

	fmtWorkspace, err := ss.formatWorkspaces([]entity.Workspace{*workspace})
	if err != nil {
		return infrastructureModel.Workspace{}, err
	}

	return fmtWorkspace[0], nil
}

func (ss *SystemServiceImpl) GetAllWorkspaces(userId primitive.ObjectID) ([]infrastructureModel.Workspace, error) {
	workspaces, err := ss.systemRepo.FindWorkspacesByUser(userId)
	if err != nil {
		return []infrastructureModel.Workspace{}, err
	}

	fmtWorkspaces, err := ss.formatWorkspaces(*workspaces)
	if err != nil {
		return []infrastructureModel.Workspace{}, err
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
		teamRoles, err := ss.systemRepo.ValidateTeam(team, userId)
		if err != nil {
			return err
		}

		if err := ss.systemRepo.AddUsersToWorkspace(workspace, teamRoles); err != nil {
			return err
		}
	}

	return errors.New("user does not have the permissions")
}

func (ss *SystemServiceImpl) UpdateWorkspaceMembers(userId primitive.ObjectID, team map[string]string, workspaceId string) error {
	workspace, err := ss.systemRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	if ss.isAdmin(workspace.Team[userId]) || ss.isOwner(workspace.Team[userId]) {
		teamRoles, err := ss.systemRepo.ValidateTeam(team, userId)
		if err != nil {
			return err
		}

		if err := ss.systemRepo.UpdateUsersInWorkspace(workspace, teamRoles); err != nil {
			return err
		}
	}

	return errors.New("user does not have the permissions")
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

	return errors.New("user does not have the permissions")
}

func (ss *SystemServiceImpl) DeleteWorkspaceById(workspaceId string, userId primitive.ObjectID) error {
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

func (ss *SystemServiceImpl) GetUserProfiles(workspaceId string, userId primitive.ObjectID) ([]infrastructureModel.User, error) {
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

func (ss *SystemServiceImpl) RegisterTelegramIntegration(userId primitive.ObjectID, workspaceId, stage, value string) (int, error) {
	workspace, err := ss.systemRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return 500, err
	}

	if _, exists := workspace.Team[userId]; !exists {
		return 500, errors.New("user does not have valid permissions")
	}

	client := resty.New()
	var resp *resty.Response

	switch stage {
	case "phone":
		reqBody := map[string]string{
			"workspace_id": workspaceId,
			"phone_number": value,
		}
		resp, err = client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Authorization", "Bearer "+ss.config.Auth.IntegrationsServerSecretKey).
			SetBody(reqBody).
			Post(ss.config.Website.IntegrationsServerURL + "/point_ai/telegram_wrapper/send_code")

		return resp.StatusCode(), err
	case "code":
		reqBody := map[string]string{
			"workspace_id": workspaceId,
			"code":         value,
			"phone_number": "placeholder_phone_number",
		}
		resp, err = client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Authorization", "Bearer "+ss.config.Auth.IntegrationsServerSecretKey).
			SetBody(reqBody).
			Post(ss.config.Website.IntegrationsServerURL + "/point_ai/telegram_wrapper/verify_init_code")

		return resp.StatusCode(), err
	case "password":
		reqBody := map[string]string{
			"workspace_id": workspaceId,
			"password":     value,
		}
		resp, err = client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Authorization", "Bearer "+ss.config.Auth.IntegrationsServerSecretKey).
			SetBody(reqBody).
			Post(ss.config.Website.IntegrationsServerURL + "/point_ai/telegram_wrapper/verify_2fa_password")

		return resp.StatusCode(), err
	}

	return 500, errors.New("invalid authentication stage")
}

func (ss *SystemServiceImpl) formatWorkspaces(workspaces []entity.Workspace) ([]infrastructureModel.Workspace, error) {
	formattedWorkspaces := make([]model.Workspace, len(workspaces))
	for i, p := range workspaces {
		logo, _ := ss.storageClient.LoadFile(p.WorkspaceId, ss.config.MinIo.BucketName)
		team, _ := ss.systemRepo.FormatTeam(p.Team)

		formattedWorkspace := infrastructureModel.Workspace{
			Name:        p.Name,
			WorkspaceId: p.WorkspaceId,
			Team:        team,
			Logo:        logo,
		}

		formattedWorkspaces[i] = formattedWorkspace
	}

	return formattedWorkspaces, nil
}

func (ss *SystemServiceImpl) EditFolders(userId primitive.ObjectID, workspaceId string, folders map[string][]string) error {
	workspace, err := ss.systemRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return err
	}

	if _, exists := workspace.Team[userId]; exists {
		workspace.Folders = folders
		return ss.systemRepo.UpdateWorkspace(workspace)
	}
	return errors.New("unauthorized")
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

func (ss *SystemServiceImpl) GetAllFolders(userId primitive.ObjectID, workspaceId string) ([]model.TeamResponse, error) {
	workspace, err := ss.systemRepo.FindWorkspaceByWorkspaceId(workspaceId)
	if err != nil {
		return nil, err
	}

	if _, exists := workspace.Team[userId]; !exists {
		return nil, errors.New("unauthorised")
	}

	var teams []model.TeamResponse
	for name, team := range workspace.InternalTeams {
		var memberCount int
		var admins []string

		for userId, _ := range team {
			memberCount++
			if ss.isAdmin(workspace.Team[userId]) || ss.isOwner(workspace.Team[userId]) {
				user, _ := ss.systemRepo.FindUserById(userId)
				admins = append(admins, user.FullName)
			}
		}
		teams = append(teams, *ss.createTeamResponse(name, memberCount, 0, admins))
	}

	return teams, nil
}

func (ss *SystemServiceImpl) createTeamResponse(teamName string, memberCount, chatCount int, adminNames []string) *model.TeamResponse {
	return &model.TeamResponse{
		TeamName:    teamName,
		MemberCount: memberCount,
		AdminNames:  adminNames,
		ChatCount:   chatCount,
	}
}

func (ss *SystemServiceImpl) isAdmin(userRole entity.WorkspaceRole) bool {
	return userRole == entity.RoleAdmin
}

func (ss *SystemServiceImpl) isOwner(userRole entity.WorkspaceRole) bool {
	return userRole == entity.RoleOwner
}
