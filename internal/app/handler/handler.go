package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"wells-risk-backend/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) GetCriterionTiles(ctx *gin.Context) {
	var criteria []repository.WellsCriterion
	var err error

	minPointsInput := ctx.Query("minPoints")

	minPoints, parseErr := strconv.ParseFloat(minPointsInput, 64)
	if minPointsInput == "" || parseErr != nil {
		criteria, err = h.Repository.GetPublishedCriteria()
	} else {
		criteria, err = h.Repository.GetCriteriaByMinPoints(minPoints)
	}
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "criterion_tiles.html", gin.H{
		"criteria":       criteria,
		"minPointsInput": minPointsInput,
		"activeTab":      "tiles",
	})
}

func (h *Handler) GetCriterionFeed(ctx *gin.Context) {
	var criterion repository.WellsCriterion
	var err error

	idStr := ctx.Param("id")

	if idStr == "" {
		criterion, err = h.Repository.GetFirstCriterion()
	} else {
		criterionID, convErr := strconv.Atoi(idStr)
		if convErr != nil {
			logrus.Error(convErr)
			ctx.String(http.StatusBadRequest, "Некорректный идентификатор критерия")
			return
		}

		if ctx.Query("next") == "true" {
			criterion, err = h.Repository.GetNextCriterion(criterionID)
		} else {
			criterion, err = h.Repository.GetCriterion(criterionID)
		}
	}

	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusNotFound, "Критерий не найден или удалён из справочника")
		return
	}

	ctx.HTML(http.StatusOK, "criterion_feed.html", gin.H{
		"criterion": criterion,
		"activeTab": "feed",
	})
}

func (h *Handler) GetCriterionDraft(ctx *gin.Context) {
	criterion, err := h.Repository.GetDraftCriterion()
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusNotFound, "Черновик критерия не найден")
		return
	}

	ctx.HTML(http.StatusOK, "criterion_draft.html", gin.H{
		"criterion": criterion,
		"activeTab": "draft",
	})
}
