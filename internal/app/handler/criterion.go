package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"wells-risk-backend/internal/app/auth"
	"wells-risk-backend/internal/app/ds"
	"wells-risk-backend/internal/app/repository"
)

// GetCriteria — GET /api/criteria
func (h *Handler) GetCriteria(ctx *gin.Context) {
	physicianID := auth.Current().PhysicianID

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

// GetCriterionFeed — GET /api/criteria/feed
func (h *Handler) GetCriterionFeed(ctx *gin.Context) {
	physicianID := auth.Current().PhysicianID

	criterion, err := h.Repository.GetFirstCriterion(physicianID)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	ctx.JSON(http.StatusOK, ds.NewCriterionResponse(criterion, physicianID))
}

// GetCriterionFeedByID — GET /api/criteria/feed/:id
func (h *Handler) GetCriterionFeedByID(ctx *gin.Context) {
	physicianID := auth.Current().PhysicianID

	criterionID, err := parseID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var criterion ds.WellsCriterion

	if ctx.Query("next") == "true" {
		criterion, err = h.Repository.GetNextCriterion(criterionID, physicianID)
	} else {
		criterion, err = h.Repository.GetPublishedCriterion(criterionID, physicianID)
	}

	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	ctx.JSON(http.StatusOK, ds.NewCriterionResponse(criterion, physicianID))
}

// GetCriterionDraft — GET /api/criteria/draft
func (h *Handler) GetCriterionDraft(ctx *gin.Context) {
	physicianID := auth.Current().PhysicianID

	criterion, err := h.Repository.GetDraftCriterion(physicianID)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	ctx.JSON(http.StatusOK, ds.NewCriterionResponse(criterion, physicianID))
}

// CreateCriterion — POST /api/criteria
func (h *Handler) CreateCriterion(ctx *gin.Context) {
	physicianID := auth.Current().PhysicianID

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
		CreatorID:        &physicianID,
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

// PublishCriterion — PUT /api/criteria/:id/publish
func (h *Handler) PublishCriterion(ctx *gin.Context) {
	physicianID := auth.Current().PhysicianID

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

// DeleteCriterion — DELETE /api/criteria/:id
func (h *Handler) DeleteCriterion(ctx *gin.Context) {
	physicianID := auth.Current().PhysicianID

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
	Value *int `json:"value" binding:"required"`
}

// LikeCriterion — POST /api/criteria/:id/like
func (h *Handler) LikeCriterion(ctx *gin.Context) {
	physicianID := auth.Current().PhysicianID

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
