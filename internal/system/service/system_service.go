package service

import (
	"errors"
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/system/domain/entity"
	"github.com/Point-AI/backend/internal/system/infrastructure/client"
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

func NewSystemServiceImpl(cfg *config.Config, storageClient *client.StorageClientImpl, systemRepo infrastructureInterface.SystemRepository) *SystemServiceImpl {
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

//func (ss *SystemServiceImpl) GetProjectByID() error {
//	// Implement logic to get a project by ID
//}

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

//
//func (ss *SystemServiceImpl) UpdateProjectByID() error {
//	// Implement logic to update a project by ID
//}

func (ss *SystemServiceImpl) AddProjectMembers(userId primitive.ObjectID, team map[string]string, projectId string) error {
	project, err := ss.systemRepo.FindProjectByProjectId(projectId)
	if err != nil {
		return err
	}

	if ss.isAdmin(project.Team[userId]) {
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

//
//func (ss *SystemServiceImpl) UpdateProjectMember() error {
//	// Implement logic to update a project member
//}
//
//func (ss *SystemServiceImpl) LeaveProject() error {
//	// Implement logic to update a project member
//}

func (ss *SystemServiceImpl) DeleteProjectMember(userId primitive.ObjectID, projectId, memberEmail string) error {
	project, err := ss.systemRepo.FindProjectByProjectId(projectId)
	if err != nil {
		return err
	}

	if ss.isAdmin(project.Team[userId]) {
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

	if ss.isAdmin(project.Team[userId]) {
		if err := ss.systemRepo.DeleteProject(project.ID); err != nil {
			return err
		}
	}

	return errors.New("user does not have a valid permission")
}

func (ss *SystemServiceImpl) formatProjects(projects []entity.Project) ([]model.Project, error) {
	formattedProjects := make([]model.Project, len(projects))

	for i, p := range projects {
		logo, _ := ss.storageClient.LoadFile(p.ProjectID+".jpg", ss.config.MinIo.BucketName)
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
