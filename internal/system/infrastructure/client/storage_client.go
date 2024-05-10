package client

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Point-AI/backend/internal/system/service/interface"
	"github.com/minio/minio-go/v7"
	"io"
	"sync"
)

type StorageClientImpl struct {
	str *minio.Client
	mu  *sync.RWMutex
}

func NewStorageClientImpl(str *minio.Client, mu *sync.RWMutex) infrastructureInterface.StorageClient {
	return &StorageClientImpl{
		str: str,
		mu:  mu,
	}
}

func (sc *StorageClientImpl) SaveFile(fileBytes []byte, bucketName, fileName string) error {
	reader := bytes.NewReader(fileBytes)

	size := int64(len(fileBytes))
	ctx := context.Background()

	if _, err := sc.str.PutObject(ctx, bucketName, fileName+".jpg", reader, size, minio.PutObjectOptions{}); err != nil {
		return fmt.Errorf("failed to upload photo: %w", err)
	}

	return nil
}

func (sc *StorageClientImpl) LoadFile(fileName, bucketName string) ([]byte, error) {
	ctx := context.Background()
	reader, err := sc.str.GetObject(ctx, bucketName, fileName+".jpg", minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	fileBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}

func (sc *StorageClientImpl) UpdateFileName(oldName, newName string, bucketName string) error {
	newSource := fmt.Sprintf("%s/%s", bucketName, newName)

	_, err := sc.str.CopyObject(context.Background(),
		minio.CopyDestOptions{
			Bucket: bucketName,
			Object: newName,
		},
		minio.CopySrcOptions{
			Bucket: bucketName,
			Object: oldName,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to copy object %s to %s: %w", oldName, newSource, err)
	}

	err = sc.str.RemoveObject(context.Background(), bucketName, oldName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object %s: %w", oldName, err)
	}

	return nil
}

func (sc *StorageClientImpl) UpdateFile(newFileBytes []byte, fileName string, bucketName string) error {
	reader := bytes.NewReader(newFileBytes)

	objectPath := fmt.Sprintf("%s/%s.jpg", bucketName, fileName)

	sc.str.RemoveObject(context.Background(), bucketName, fileName+".jpg", minio.RemoveObjectOptions{})

	_, err := sc.str.PutObject(context.Background(), objectPath, "", reader, int64(len(newFileBytes)), minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to upload new content for %s: %w", fileName, err)
	}

	return nil
}
