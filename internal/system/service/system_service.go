package service

import (
	"github.com/Point-AI/backend/config"
	"github.com/Point-AI/backend/internal/system/infrastructure/client"
	infrastructureInterface "github.com/Point-AI/backend/internal/system/service/interface"
	"github.com/Point-AI/backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (ss *SystemServiceImpl) CreateProject(logo []byte, team []string, ownerId primitive.ObjectID, projectId, name string) error {
	if err := utils.ValidateProjectId(projectId); err != nil {
		return err
	}

	if err := utils.ValidatePhoto(logo); err != nil {
		return err
	}

	teamIds, err := ss.systemRepo.ValidateTeam(team)
	if err != nil {
		return err
	}

	if err := ss.systemRepo.CreateProject(teamIds, projectId, name, ownerId); err != nil {
		return err
	}

	if err := ss.storageClient.SaveFile(logo, ss.config.MinIo.BucketName, name); err != nil {
		return err
	}

	return nil
}

func (ss *SystemServiceImpl) LeaveProject(projectId string, userId primitive.ObjectID) error {
	project, err := ss.systemRepo.FindProjectById(projectId)
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
//
//func (ss *SystemServiceImpl) GetAllProjects() error {
//	// Implement logic to get all projects
//}
//
//func (ss *SystemServiceImpl) UpdateProjectByID() error {
//	// Implement logic to update a project by ID
//}
//
//func (ss *SystemServiceImpl) AddProjectMember() error {
//	// Implement logic to update a project member
//}
//
//func (ss *SystemServiceImpl) UpdateProjectMember() error {
//	// Implement logic to update a project member
//}
//
//func (ss *SystemServiceImpl) LeaveProject() error {
//	// Implement logic to update a project member
//}
//
//func (ss *SystemServiceImpl) DeleteProjectMember() error {
//	// Implement logic to delete a project member
//}
//
//func (ss *SystemServiceImpl) DeleteProjectByID() error {
//	// Implement logic to delete a project by ID
//}
