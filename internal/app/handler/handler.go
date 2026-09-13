package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"wells-risk-backend/internal/app/ds"
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
	var criteria []ds.WellsCriterion
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

	ctx.HTML(http.StatusOK, "criteria_tiles.html", gin.H{
		"criteria":       criteria,
		"minPointsInput": minPointsInput,
		"activeTab":      "tiles",
	})
}

func (h *Handler) GetCriterionFeed(ctx *gin.Context) {
	var criterion ds.WellsCriterion
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

	ctx.HTML(http.StatusOK, "criteria_feed.html", gin.H{
		"criterion": criterion,
		"activeTab": "feed",
	})
}

func (h *Handler) GetCriterionDraft(ctx *gin.Context) {
	criterion, err := h.Repository.GetDraftCriterion()
	hasDraft := err == nil

	ctx.HTML(http.StatusOK, "criteria_draft.html", gin.H{
		"criterion": criterion,
		"hasDraft":  hasDraft,
		"activeTab": "draft",
	})
}

// CreateCriterionDraft — создание карточки критерия в статусе «черновик».
func (h *Handler) CreateCriterionDraft(ctx *gin.Context) {
	points, err := strconv.ParseFloat(ctx.PostForm("wellsPoints"), 64)
	if err != nil {
		points = 0
	}

	creatorID := 1 // пока авторизации нет, черновик заводит первый врач

	criterion := ds.WellsCriterion{
		CriterionName:    ctx.PostForm("criterionName"),
		ShortDescription: ctx.PostForm("shortDescription"),
		CriterionStatus:  ds.StatusDraft,
		WellsPoints:      points,
		CriterionGroup:   ctx.PostForm("criterionGroup"),
		CreatedAt:        time.Now(),
		CreatorID:        &creatorID,
	}

	if err := h.Repository.CreateCriterion(&criterion); err != nil {
		logrus.Error(err)
		ctx.String(http.StatusInternalServerError, "Не удалось сохранить черновик критерия")
		return
	}

	ctx.Redirect(http.StatusFound, "/criteria/draft")
}

// PublishCriterion — публикация черновика.
func (h *Handler) PublishCriterion(ctx *gin.Context) {
	criterionID, err := strconv.Atoi(ctx.PostForm("criterionID"))
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusBadRequest, "Некорректный идентификатор критерия")
		return
	}

	if err := h.Repository.PublishCriterion(criterionID); err != nil {
		logrus.Error(err)
		ctx.String(http.StatusNotFound, "Черновик критерия не найден")
		return
	}

	ctx.Redirect(http.StatusFound, "/criteria")
}

// DeleteCriterion — логическое удаление критерия из справочника.
func (h *Handler) DeleteCriterion(ctx *gin.Context) {
	criterionID, err := strconv.Atoi(ctx.PostForm("criterionID"))
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusBadRequest, "Некорректный идентификатор критерия")
		return
	}

	if err := h.Repository.DeleteCriterion(criterionID); err != nil {
		logrus.Error(err)
		ctx.String(http.StatusInternalServerError, "Не удалось удалить критерий")
		return
	}

	ctx.Redirect(http.StatusFound, "/criteria")
}
