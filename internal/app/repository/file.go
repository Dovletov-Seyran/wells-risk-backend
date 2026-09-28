package repository

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/minio/minio-go/v7"

	"wells-risk-backend/internal/app/ds"
)

type FileKind string

const (
	FileImage FileKind = "image"
	FileVideo FileKind = "video"
)

var extensions = map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/gif":       ".gif",
	"image/webp":      ".webp",
	"video/mp4":       ".mp4",
	"video/quicktime": ".mov",
	"video/webm":      ".webm",
}

func DetectContentType(header *multipart.FileHeader) (string, error) {
	file, err := header.Open()
	if err != nil {
		return "", fmt.Errorf("открытие файла: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 512)

	read, err := file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("чтение файла: %w", err)
	}

	return http.DetectContentType(buffer[:read]), nil
}

func (r *Repository) UploadCriterionFile(criterionID int, kind FileKind, header *multipart.FileHeader) (string, error) {
	contentType, err := DetectContentType(header)
	if err != nil {
		return "", err
	}

	extension, ok := extensions[contentType]
	if !ok {
		return "", fmt.Errorf("неподдерживаемый тип файла %s", contentType)
	}

	objectKey := fmt.Sprintf("criterion_%d_%s_%d%s", criterionID, kind, time.Now().Unix(), extension)

	file, err := header.Open()
	if err != nil {
		return "", fmt.Errorf("открытие файла: %w", err)
	}
	defer file.Close()

	_, err = r.minio.PutObject(
		context.Background(),
		r.minioBucketName,
		objectKey,
		file,
		header.Size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", fmt.Errorf("загрузка файла в MinIO: %w", err)
	}

	column := "image_key"
	if kind == FileVideo {
		column = "video_key"
	}

	err = r.db.Model(&ds.WellsCriterion{}).
		Where("criterion_id = ?", criterionID).
		UpdateColumn(column, objectKey).Error
	if err != nil {
		_ = r.minio.RemoveObject(context.Background(), r.minioBucketName, objectKey, minio.RemoveObjectOptions{})

		return "", fmt.Errorf("сохранение ключа файла: %w", err)
	}

	return objectKey, nil
}
