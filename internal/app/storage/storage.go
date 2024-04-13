package storage

import (
	"context"
	"fmt"
	"github.com/Point-AI/backend/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func ConnectToStorage(cfg *config.Config) *minio.Client {
	minioClient, err := minio.New(cfg.MinIo.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(cfg.MinIo.AccessKey, cfg.MinIo.SecretKey, ""),
		//Secure: cfg.MinIo.UseSSL,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create MinIO client: %w", err))
	}

	ctx := context.Background()
	found, err := minioClient.BucketExists(ctx, cfg.MinIo.BucketName)
	if err != nil {
		panic(fmt.Errorf("error checking bucket existence: %w", err))
	}
	if !found {
		err := minioClient.MakeBucket(ctx, cfg.MinIo.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			panic(fmt.Errorf("failed to create bucket: %w", err))
		}
	}

	return minioClient
}
