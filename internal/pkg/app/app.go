package app

import (
	"context"
	"log"
	"os"

	"wells-risk-backend/internal/app/config"
	"wells-risk-backend/internal/app/ds"
	"wells-risk-backend/internal/app/dsn"
	"wells-risk-backend/internal/app/handler"
	"wells-risk-backend/internal/app/redis"
	"wells-risk-backend/internal/app/repository"
)

type Application struct {
	config  *config.Config
	repo    *repository.Repository
	redis   *redis.Client
	handler *handler.Handler
}

func New(ctx context.Context) (*Application, error) {
	cfg, err := config.NewConfig(ctx)
	if err != nil {
		return nil, err
	}

	repo, err := repository.New(&repository.RepositorySettings{
		PostgresDSN:     dsn.FromEnv(),
		MinioEndpoint:   os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey:  os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey:  os.Getenv("MINIO_SECRET_KEY"),
		MinioBucketName: os.Getenv("MINIO_BUCKET_NAME"),
		MinioUseSSL:     false,
	})
	if err != nil {
		return nil, err
	}

	// Публичный адрес файлов критериев для ответов веб-сервиса.
	ds.SetFilesBaseURL(os.Getenv("MINIO_PUBLIC_URL") + "/" + os.Getenv("MINIO_BUCKET_NAME"))

	redisClient, err := redis.New(ctx, cfg.Redis)
	if err != nil {
		return nil, err
	}

	return &Application{
		config:  cfg,
		repo:    repo,
		redis:   redisClient,
		handler: handler.NewHandler(repo),
	}, nil
}

func (a *Application) Run() error {
	log.Println("application start running")
	a.StartServer()
	log.Println("application shut down")

	return nil
}
