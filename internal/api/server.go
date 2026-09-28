package api

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"wells-risk-backend/internal/app/ds"
	"wells-risk-backend/internal/app/dsn"
	"wells-risk-backend/internal/app/handler"
	"wells-risk-backend/internal/app/repository"
)

func StartServer() {
	log.Println("Server start up")

	repo, err := repository.New(&repository.RepositorySettings{
		PostgresDSN:     dsn.FromEnv(),
		MinioEndpoint:   os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey:  os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey:  os.Getenv("MINIO_SECRET_KEY"),
		MinioBucketName: os.Getenv("MINIO_BUCKET_NAME"),
		MinioUseSSL:     false,
	})
	if err != nil {
		logrus.Fatalf("ошибка подключения к хранилищам: %v", err)
	}

	// Публичный адрес файлов критериев для ответов веб-сервиса.
	ds.SetFilesBaseURL(os.Getenv("MINIO_PUBLIC_URL") + "/" + os.Getenv("MINIO_BUCKET_NAME"))

	criterionHandler := handler.NewHandler(repo)

	r := gin.Default()

	api := r.Group("/api")
	{
		criteria := api.Group("/criteria")
		{
			criteria.GET("", criterionHandler.GetCriteria)
			criteria.GET("/feed", criterionHandler.GetCriterionFeed)
			criteria.GET("/feed/:id", criterionHandler.GetCriterionFeedByID)
			criteria.GET("/draft", criterionHandler.GetCriterionDraft)
			criteria.POST("", criterionHandler.CreateCriterion)
			criteria.PUT("/:id/publish", criterionHandler.PublishCriterion)
			criteria.DELETE("/:id", criterionHandler.DeleteCriterion)
			criteria.POST("/:id/like", criterionHandler.LikeCriterion)
		}

		physicians := api.Group("/physicians")
		{
			physicians.POST("/register", criterionHandler.RegisterPhysician)
			physicians.POST("/login", criterionHandler.LoginPhysician)
			physicians.POST("/logout", criterionHandler.LogoutPhysician)
		}
	}

	if err := r.Run(); err != nil {
		logrus.Fatalf("ошибка запуска сервера: %v", err)
	}

	log.Println("Server down")
}
