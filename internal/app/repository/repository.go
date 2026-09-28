package repository

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type RepositorySettings struct {
	PostgresDSN     string
	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioBucketName string
	MinioUseSSL     bool
}

type Repository struct {
	db              *gorm.DB
	minio           *minio.Client
	minioBucketName string
}

func New(settings *RepositorySettings) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(settings.PostgresDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("подключение к PostgreSQL: %w", err)
	}

	minioClient, err := minio.New(settings.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(settings.MinioAccessKey, settings.MinioSecretKey, ""),
		Secure: settings.MinioUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("подключение к MinIO: %w", err)
	}

	repository := &Repository{
		db:              db,
		minio:           minioClient,
		minioBucketName: settings.MinioBucketName,
	}

	if err := repository.prepareBucket(context.Background()); err != nil {
		return nil, err
	}

	return repository, nil
}

func (r *Repository) prepareBucket(ctx context.Context) error {
	exists, err := r.minio.BucketExists(ctx, r.minioBucketName)
	if err != nil {
		return fmt.Errorf("проверка бакета MinIO: %w", err)
	}

	if !exists {
		if err := r.minio.MakeBucket(ctx, r.minioBucketName, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("создание бакета MinIO: %w", err)
		}
	}

	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}
		]
	}`, r.minioBucketName)

	if err := r.minio.SetBucketPolicy(ctx, r.minioBucketName, policy); err != nil {
		return fmt.Errorf("политика доступа к бакету MinIO: %w", err)
	}

	return nil
}
