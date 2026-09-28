package handler

import (
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

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

// errorHandler пишет причину в лог сервера и возвращает клиенту код ответа.
func (h *Handler) errorHandler(ctx *gin.Context, statusCode int, err error) {
	if err != nil {
		logrus.Error(err)
	}

	ctx.Status(statusCode)
}

// parseID разбирает ид критерия из пути.
func parseID(ctx *gin.Context) (int, error) {
	return strconv.Atoi(ctx.Param("id"))
}

// parsePoints разбирает баллы Уэллса.
func parsePoints(value string) (float64, error) {
	normalized := strings.ReplaceAll(strings.TrimSpace(value), ",", ".")

	return strconv.ParseFloat(normalized, 64)
}

// imageTypes — типы содержимого, допустимые для изображения критерия.
var imageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// videoTypes — типы содержимого, допустимые для короткого видео критерия.
var videoTypes = map[string]bool{
	"video/mp4":       true,
	"video/quicktime": true,
	"video/webm":      true,
}

// validateUpload проверяет, что загруженный файл нужного вида.
func validateUpload(header *multipart.FileHeader, allowed map[string]bool) error {
	contentType, err := repository.DetectContentType(header)
	if err != nil {
		return err
	}

	if !allowed[contentType] {
		return errUnsupportedFile{contentType: contentType}
	}

	return nil
}

type errUnsupportedFile struct {
	contentType string
}

func (e errUnsupportedFile) Error() string {
	return "недопустимый тип файла: " + e.contentType
}

// formFile достаёт файл из формы.
func formFile(ctx *gin.Context, field string) (*multipart.FileHeader, bool, error) {
	header, err := ctx.FormFile(field)
	if err != nil {
		if err == http.ErrMissingFile {
			return nil, false, nil
		}

		return nil, false, err
	}

	return header, true, nil
}
