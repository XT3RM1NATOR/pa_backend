package storage

import (
	"fmt"
	"github.com/Point-AI/backend/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func ConnectToStorage(cfg *config.Config) *minio.Client {
	minioClient, err := minio.New("play.min.io", &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIo.AccessKey, cfg.MinIo.SecretKey, ""),
		Secure: true,
	})

	if err != nil {
		panic(fmt.Errorf("failed to create MinIO client: %w", err))
	}

	//err = minioClient.MakeBucket(context.Background(), cfg.MinIo.BucketName, minio.MakeBucketOptions{})
	//if err != nil {
	//	panic(fmt.Errorf("failed to create bucket: %w", err))
	//}

	return minioClient
}
