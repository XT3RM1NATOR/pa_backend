package service

import (
	"errors"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/system/domain/entity"
	"github.com/Point-AI/backend/internal/system/infrastructure/model"
	infrastructureInterface "github.com/Point-AI/backend/internal/system/service/interface"
	"github.com/Point-AI/backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SystemServiceImpl struct {
	systemRepo    infrastructureInterface.SystemRepository
	storageClient infrastructureInterface.StorageClient
	config        *config.Config
}

func NewSystemServiceImpl(cfg *config.Config, storageClient infrastructureInterface.StorageClient, systemRepo infrastructureInterface.SystemRepository) *SystemServiceImpl {
	return &SystemServiceImpl{
		systemRepo:    systemRepo,
		storageClient: storageClient,
		config:        cfg,
	}
}

func (ss *SystemServiceImpl) CreateProject(logo []byte, team map[string]string, ownerId primitive.ObjectID, projectId, name string) error {
	if err := utils.ValidateProjectId(projectId); err != nil {
		return err
	}

	if err := utils.ValidatePhoto(logo); err != nil {
		return err
	}

	teamRoles, err := ss.systemRepo.ValidateTeam(team, ownerId)
	if err != nil {
		return err
	}

	_, err = ss.systemRepo.FindProjectByProjectId(projectId)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}

	if err := ss.systemRepo.CreateProject(teamRoles, projectId, name); err != nil {
		return err
	}

	if err := ss.storageClient.SaveFile(logo, ss.config.MinIo.BucketName, name); err != nil {
		return err
	}

	return nil
}

func (ss *SystemServiceImpl) LeaveProject(projectId string, userId primitive.ObjectID) error {
	project, err := ss.systemRepo.FindProjectByProjectId(projectId)
	if err != nil {
		return err
	}

	if err := ss.systemRepo.RemoveUserFromProject(project, userId); err != nil {
		return err
	}

	return nil
}

func (ss *SystemServiceImpl) GetProjectById(projectId string, userId primitive.ObjectID) (model.Project, error) {
	project, err := ss.systemRepo.FindProjectByProjectId(projectId)
	if err != nil {
		return model.Project{}, err
	}

	if _, exists := project.Team[userId]; !exists {
		return model.Project{}, errors.New("user is not in the project")
	}

	fmtProject, err := ss.formatProjects([]entity.Project{project})
	if err != nil {
		return model.Project{}, err
	}

	return fmtProject[0], nil
}

func (ss *SystemServiceImpl) GetAllProjects(userId primitive.ObjectID) ([]model.Project, error) {
	projects, err := ss.systemRepo.FindProjectsByUser(userId)
	if err != nil {
		return []model.Project{}, err
	}

	fmtProjects, err := ss.formatProjects(projects)
	if err != nil {
		return []model.Project{}, err
	}
	return fmtProjects, err
}

func (ss *SystemServiceImpl) UpdateProject(userId primitive.ObjectID, newLogo []byte, projectId, newProjectId, newName string) error {
	project, err := ss.systemRepo.FindProjectByProjectId(projectId)
	if err != nil {
		return err
	}

	if ss.isOwner(project.Team[userId]) || ss.isAdmin(project.Team[userId]) {
		if newProjectId != "" {
			if err := utils.ValidateProjectId(projectId); err != nil {
				return err
			}
			if err := ss.storageClient.UpdateFileName(projectId, newProjectId, ss.config.MinIo.BucketName); err != nil {
				return err
			}
			project.ProjectID = newProjectId
		}

		if newLogo != nil {
			if err := utils.ValidatePhoto(newLogo); err != nil {
				return err
			}
			if err := ss.storageClient.UpdateFile(newLogo, project.ProjectID, ss.config.MinIo.BucketName); err != nil {
				return err
			}
		}

		if newName != "" {
			project.Name = newName
		}

		if err := ss.systemRepo.UpdateProject(project); err != nil {
			return err
		}
		return nil
	}
	return errors.New("unauthorized to make the changes")
}

func (ss *SystemServiceImpl) AddProjectMembers(userId primitive.ObjectID, team map[string]string, projectId string) error {
	project, err := ss.systemRepo.FindProjectByProjectId(projectId)
	if err != nil {
		return err
	}

	if ss.isAdmin(project.Team[userId]) || ss.isOwner(project.Team[userId]) {
		teamRoles, err := ss.systemRepo.ValidateTeam(team, userId)
		if err != nil {
			return err
		}

		if err := ss.systemRepo.AddUsersToProject(project, teamRoles); err != nil {
			return err
		}
	}

	return nil
}

func (ss *SystemServiceImpl) UpdateProjectMembers(userId primitive.ObjectID, team map[string]string, projectId string) error {
	project, err := ss.systemRepo.FindProjectByProjectId(projectId)
	if err != nil {
		return err
	}

	if ss.isAdmin(project.Team[userId]) || ss.isOwner(project.Team[userId]) {
		teamRoles, err := ss.systemRepo.ValidateTeam(team, userId)
		if err != nil {
			return err
		}

		if err := ss.systemRepo.UpdateUsersInProject(project, teamRoles); err != nil {
			return err
		}
	}

	return nil
}

func (ss *SystemServiceImpl) DeleteProjectMember(userId primitive.ObjectID, projectId, memberEmail string) error {
	project, err := ss.systemRepo.FindProjectByProjectId(projectId)
	if err != nil {
		return err
	}

	if ss.isAdmin(project.Team[userId]) || ss.isOwner(project.Team[userId]) {
		userId, err := ss.systemRepo.FindUserByEmail(memberEmail)
		if err != nil {
			return err
		}

		if err := ss.systemRepo.RemoveUserFromProject(project, userId); err != nil {
			return err
		}
	}

	return nil
}

func (ss *SystemServiceImpl) DeleteProjectByID(projectId string, userId primitive.ObjectID) error {
	project, err := ss.systemRepo.FindProjectByProjectId(projectId)
	if err != nil {
		return err
	}

	if ss.isOwner(project.Team[userId]) {
		if err := ss.systemRepo.DeleteProject(project.ID); err != nil {
			return err
		}
	}

	return errors.New("user does not have a valid permission")
}

func (ss *SystemServiceImpl) formatProjects(projects []entity.Project) ([]model.Project, error) {
	formattedProjects := make([]model.Project, len(projects))

	for i, p := range projects {
		logo, _ := ss.storageClient.LoadFile(p.ProjectID, ss.config.MinIo.BucketName)
		formattedProject := model.Project{
			Name:      p.Name,
			ProjectID: p.ProjectID,
			Logo:      logo,
		}

		formattedProjects[i] = formattedProject
	}

	return formattedProjects, nil
}

func (ss *SystemServiceImpl) isAdmin(userRole entity.ProjectRole) bool {
	return userRole == entity.RoleAdmin
}

func (ss *SystemServiceImpl) isOwner(userRole entity.ProjectRole) bool {
	return userRole == entity.RoleOwner
}
