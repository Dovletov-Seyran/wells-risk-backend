package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"wells-risk-backend/internal/app/ds"
	"wells-risk-backend/internal/app/repository"
)

// GetCriteria godoc
// @Summary      Список критериев
// @Description  Опубликованные критерии шкалы Уэллса с фильтром по минимальному весу. Доступно гостю; с токеном is_mine = 1 у критериев текущего врача
// @Tags         Критерии
// @Produce      json
// @Param        min_points  query     number  false  "минимальный вес критерия в баллах Уэллса"
// @Success      200         {array}   ds.CriterionResponse
// @Failure      400
// @Failure      500
// @Router       /api/criteria [get]
func (h *Handler) GetCriteria(ctx *gin.Context) {
	physicianID := currentPhysicianID(ctx)

	var minPoints *float64

	if input := ctx.Query("min_points"); input != "" {
		points, err := parsePoints(input)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}

		minPoints = &points
	}

	criteria, err := h.Repository.GetPublishedCriteria(minPoints, physicianID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, ds.NewCriterionListResponse(criteria, physicianID))
}

// GetCriterionFeed godoc
// @Summary      Лента критериев
// @Description  Без параметров — первый критерий ленты, с criterion_id — этот критерий, с next=true — следующий после него. Доступно гостю; с токеном is_liked = 1, если текущий врач поставил лайк
// @Tags         Критерии
// @Produce      json
// @Param        criterion_id  query     int     false  "ид критерия ленты"
// @Param        next          query     bool    false  "следующий критерий после указанного"
// @Success      200           {object}  ds.CriterionResponse
// @Failure      400
// @Failure      404
// @Router       /api/criteria/feed [get]
func (h *Handler) GetCriterionFeed(ctx *gin.Context) {
	physicianID := currentPhysicianID(ctx)

	var criterion ds.WellsCriterion
	var err error

	// Без ид — первый критерий ленты, с ид — этот критерий или следующий при next=true.
	if input := ctx.Query("criterion_id"); input == "" {
		criterion, err = h.Repository.GetFirstCriterion(physicianID)
	} else {
		criterionID, parseErr := strconv.Atoi(input)
		if parseErr != nil {
			h.errorHandler(ctx, http.StatusBadRequest, parseErr)
			return
		}

		if ctx.Query("next") == "true" {
			criterion, err = h.Repository.GetNextCriterion(criterionID, physicianID)
		} else {
			criterion, err = h.Repository.GetPublishedCriterion(criterionID, physicianID)
		}
	}

	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	ctx.JSON(http.StatusOK, ds.NewCriterionResponse(criterion, physicianID))
}

// GetCriterionDraft godoc
// @Summary      Черновик критерия
// @Description  Черновик текущего врача, ид не указывается. У врача не более одного черновика
// @Tags         Критерии
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  ds.CriterionResponse
// @Failure      403
// @Failure      404
// @Router       /api/criteria/draft [get]
func (h *Handler) GetCriterionDraft(ctx *gin.Context) {
	physicianID := currentPhysicianID(ctx)

	criterion, err := h.Repository.GetDraftCriterion(physicianID)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	ctx.JSON(http.StatusOK, ds.NewCriterionResponse(criterion, physicianID))
}

// CreateCriterion godoc
// @Summary      Добавление критерия
// @Description  Новый критерий в статусе черновик, создатель берётся из JWT. Изображение и видео загружаются в MinIO, в БД сохраняются ключи файлов
// @Tags         Критерии
// @Accept       multipart/form-data
// @Produce      json
// @Security     ApiKeyAuth
// @Param        criterion_name     formData  string  true   "название критерия"
// @Param        short_description  formData  string  false  "краткое описание"
// @Param        criterion_group    formData  string  false  "группа критерия"
// @Param        wells_points       formData  number  false  "вес критерия в баллах Уэллса"
// @Param        image              formData  file    false  "изображение критерия"
// @Param        video              formData  file    false  "короткое видео критерия"
// @Success      201                {object}  ds.CriterionResponse
// @Failure      400
// @Failure      403
// @Failure      409
// @Failure      500
// @Router       /api/criteria [post]
func (h *Handler) CreateCriterion(ctx *gin.Context) {
	physicianID := currentPhysicianID(ctx)

	if err := ctx.Request.ParseMultipartForm(32 << 20); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	name := strings.TrimSpace(ctx.Request.FormValue("criterion_name"))
	if name == "" {
		h.errorHandler(ctx, http.StatusBadRequest, nil)
		return
	}

	criterion := ds.WellsCriterion{
		CriterionName:    name,
		ShortDescription: strings.TrimSpace(ctx.Request.FormValue("short_description")),
		CriterionGroup:   strings.TrimSpace(ctx.Request.FormValue("criterion_group")),
		CriterionStatus:  ds.StatusDraft,
		CreatedAt:        time.Now(),
		CreatorID:        physicianID,
	}

	if input := ctx.Request.FormValue("wells_points"); input != "" {
		points, err := parsePoints(input)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}

		criterion.WellsPoints = points
	}

	imageHeader, hasImage, err := formFile(ctx, "image")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if hasImage {
		if err := validateUpload(imageHeader, imageTypes); err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
	}

	videoHeader, hasVideo, err := formFile(ctx, "video")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if hasVideo {
		if err := validateUpload(videoHeader, videoTypes); err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
	}

	// У врача не может быть больше одного черновика критерия.
	drafts, err := h.Repository.CountDrafts(physicianID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if drafts > 0 {
		h.errorHandler(ctx, http.StatusConflict, nil)
		return
	}

	if err := h.Repository.CreateCriterion(&criterion); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if hasImage {
		if _, err := h.Repository.UploadCriterionFile(criterion.CriterionID, repository.FileImage, imageHeader); err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	if hasVideo {
		if _, err := h.Repository.UploadCriterionFile(criterion.CriterionID, repository.FileVideo, videoHeader); err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	created, err := h.Repository.GetCriterion(criterion.CriterionID, physicianID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, ds.NewCriterionResponse(created, physicianID))
}

// PublishCriterion godoc
// @Summary      Публикация критерия
// @Description  Смена статуса черновик → опубликован, заполняется дата формирования. Только свой черновик
// @Tags         Критерии
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      int  true  "ид критерия"
// @Success      200  {object}  ds.CriterionResponse
// @Failure      400
// @Failure      403
// @Failure      404
// @Router       /api/criteria/{id}/publish [put]
func (h *Handler) PublishCriterion(ctx *gin.Context) {
	physicianID := currentPhysicianID(ctx)

	criterionID, err := parseID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.PublishCriterion(criterionID, physicianID); err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	criterion, err := h.Repository.GetCriterion(criterionID, physicianID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, ds.NewCriterionResponse(criterion, physicianID))
}

// DeleteCriterion godoc
// @Summary      Удаление критерия
// @Description  Логическое удаление: статус меняется на удалён. Только свои критерии
// @Tags         Критерии
// @Security     ApiKeyAuth
// @Param        id   path  int  true  "ид критерия"
// @Success      204
// @Failure      400
// @Failure      403
// @Failure      404
// @Router       /api/criteria/{id} [delete]
func (h *Handler) DeleteCriterion(ctx *gin.Context) {
	physicianID := currentPhysicianID(ctx)

	criterionID, err := parseID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.DeleteCriterion(criterionID, physicianID); err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// likeRequest — тело запроса отметки
type likeRequest struct {
	Value *int `json:"value" binding:"required" example:"1"`
}

// LikeCriterion godoc
// @Summary      Лайк критерия
// @Description  Лайк текущего врача: value = 1 ставит лайк, value = 0 отменяет
// @Tags         Критерии
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id    path      int          true  "ид критерия"
// @Param        body  body      likeRequest  true  "значение лайка"
// @Success      200   {object}  ds.CriterionResponse
// @Failure      400
// @Failure      403
// @Failure      404
// @Router       /api/criteria/{id}/like [post]
func (h *Handler) LikeCriterion(ctx *gin.Context) {
	physicianID := currentPhysicianID(ctx)

	criterionID, err := parseID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var request likeRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if *request.Value != 0 && *request.Value != 1 {
		h.errorHandler(ctx, http.StatusBadRequest, nil)
		return
	}

	criterion, err := h.Repository.GetPublishedCriterion(criterionID, physicianID)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	if err := h.Repository.SetLike(criterion.CriterionID, physicianID, *request.Value); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	updated, err := h.Repository.GetPublishedCriterion(criterionID, physicianID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, ds.NewCriterionResponse(updated, physicianID))
}
