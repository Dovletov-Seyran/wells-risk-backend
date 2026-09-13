package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"wells-risk-backend/internal/app/handler"
	"wells-risk-backend/internal/app/repository"

	"wells-risk-backend/internal/app/dsn"
)

func StartServer() {
	log.Println("Server start up")

	repo, err := repository.New(dsn.FromEnv())
	if err != nil {
		logrus.Fatalf("ошибка подключения к базе: %v", err)
	}

	criterioHandler := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/criteria")
	})

	r.GET("/criteria", criterioHandler.GetCriterionTiles)
	r.GET("/criteria/feed", criterioHandler.GetCriterionFeed)
	r.GET("/criteria/feed/:id", criterioHandler.GetCriterionFeed)
	r.GET("/criteria/draft", criterioHandler.GetCriterionDraft)
	r.POST("/criteria/draft", criterioHandler.CreateCriterionDraft)
	r.POST("/criteria/publish", criterioHandler.PublishCriterion)
	r.POST("/criteria/delete", criterioHandler.DeleteCriterion)

	r.Run()

	log.Println("Server down")
}
