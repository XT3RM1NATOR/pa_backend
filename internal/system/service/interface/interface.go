package infrastructureInterface

import (
	"github.com/Point-AI/backend/internal/system/domain/entity"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StorageClient interface {
	SaveFile(fileBytes []byte, bucketName, objectName string) error
}

type SystemRepository interface {
	ValidateTeam(team []string) ([]primitive.ObjectID, error)
	CreateProject(team []primitive.ObjectID, projectId, name string, ownerId primitive.ObjectID) error
	RemoveUserFromProject(project entity.Project, userId primitive.ObjectID) error
	FindProjectById(projectId string) (entity.Project, error)
}
