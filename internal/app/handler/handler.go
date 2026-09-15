package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"wells-risk-backend/internal/app/ds"
	"wells-risk-backend/internal/app/repository"
)

const currentPhysicianID = 1

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
	normalized := strings.ReplaceAll(minPointsInput, ",", ".")

	minPoints, parseErr := strconv.ParseFloat(normalized, 64)

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

// CreateCriterionDraft — создание карточки критерия в статусе черновик
func (h *Handler) CreateCriterionDraft(ctx *gin.Context) {
	name := strings.TrimSpace(ctx.PostForm("criterionName"))
	description := strings.TrimSpace(ctx.PostForm("shortDescription"))
	group := strings.TrimSpace(ctx.PostForm("criterionGroup"))
	imageKey := strings.TrimSpace(ctx.PostForm("imageKey"))
	videoKey := strings.TrimSpace(ctx.PostForm("videoKey"))

	pointsInput := strings.ReplaceAll(strings.TrimSpace(ctx.PostForm("wellsPoints")), ",", ".")
	points, parseErr := strconv.ParseFloat(pointsInput, 64)

	if name == "" || description == "" || group == "" || imageKey == "" || parseErr != nil {
		ctx.String(http.StatusBadRequest, "Заполните все поля критерия")
		return
	}

	creatorID := currentPhysicianID

	criterion := ds.WellsCriterion{
		CriterionName:    name,
		ShortDescription: description,
		CriterionStatus:  ds.StatusDraft,
		ImageKey:         imageKey,
		VideoKey:         videoKey,
		WellsPoints:      points,
		CriterionGroup:   group,
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

	name := strings.TrimSpace(ctx.PostForm("criterionName"))
	description := strings.TrimSpace(ctx.PostForm("shortDescription"))
	group := strings.TrimSpace(ctx.PostForm("criterionGroup"))
	imageKey := strings.TrimSpace(ctx.PostForm("imageKey"))
	videoKey := strings.TrimSpace(ctx.PostForm("videoKey"))

	pointsInput := strings.ReplaceAll(strings.TrimSpace(ctx.PostForm("wellsPoints")), ",", ".")
	points, parseErr := strconv.ParseFloat(pointsInput, 64)

	if name == "" || description == "" || group == "" || imageKey == "" || parseErr != nil {
		ctx.String(http.StatusBadRequest, "Заполните все поля критерия перед публикацией")
		return
	}

	draft := ds.WellsCriterion{
		CriterionName:    name,
		ShortDescription: description,
		CriterionGroup:   group,
		WellsPoints:      points,
		ImageKey:         imageKey,
		VideoKey:         videoKey,
	}

	if err := h.Repository.PublishCriterion(criterionID, draft); err != nil {
		logrus.Error(err)
		ctx.String(http.StatusNotFound, "Черновик критерия не найден")
		return
	}

	ctx.Redirect(http.StatusFound, "/criteria")
}

// DeleteCriterion — логическое удаление критерия из справочника
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
