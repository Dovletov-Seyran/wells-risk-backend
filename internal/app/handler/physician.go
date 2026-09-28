package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"wells-risk-backend/internal/app/auth"
	"wells-risk-backend/internal/app/ds"
)

type registerRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
	FullName string `json:"full_name"`
}

type loginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterPhysician — POST /api/physicians/register
func (h *Handler) RegisterPhysician(ctx *gin.Context) {
	var request registerRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	login := strings.TrimSpace(request.Login)
	if login == "" {
		h.errorHandler(ctx, http.StatusBadRequest, nil)
		return
	}

	exists, err := h.Repository.PhysicianExists(login)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if exists {
		h.errorHandler(ctx, http.StatusConflict, nil)
		return
	}

	physician := ds.Physician{
		Login:       login,
		Password:    request.Password,
		FullName:    strings.TrimSpace(request.FullName),
		IsModerator: false,
	}

	if err := h.Repository.CreatePhysician(&physician); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, ds.NewPhysicianResponse(physician))
}

// LoginPhysician — POST /api/physicians/login
func (h *Handler) LoginPhysician(ctx *gin.Context) {
	var request loginRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	physician, err := h.Repository.GetPhysician(auth.Current().PhysicianID)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	ctx.JSON(http.StatusOK, ds.NewPhysicianResponse(physician))
}

// LogoutPhysician — POST /api/physicians/logout
func (h *Handler) LogoutPhysician(ctx *gin.Context) {
	ctx.Status(http.StatusNoContent)
}
