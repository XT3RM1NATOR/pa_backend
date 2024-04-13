package client

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Point-AI/backend/internal/system/service/interface"
	"github.com/minio/minio-go/v7"
)

type StorageClientImpl struct {
	str *minio.Client
}

func NewStorageClientImpl(str *minio.Client) infrastructureInterface.StorageClient {
	return &StorageClientImpl{
		str: str,
	}
}

func (sc *StorageClientImpl) SaveFile(fileBytes []byte, bucketName, objectName string) error {
	reader := bytes.NewReader(fileBytes)

	size := int64(len(fileBytes))
	ctx := context.Background()

	if _, err := sc.str.PutObject(ctx, bucketName, objectName+".jpg", reader, size, minio.PutObjectOptions{}); err != nil {
		return fmt.Errorf("failed to upload photo: %w", err)
	}

	return nil
}
