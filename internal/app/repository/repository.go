package repository

import (
	"fmt"
	"sort"
	"strconv"
	"time"
)

const (
	minioPublicURL = "http://localhost:9000"
	minioBucket    = "wells-criteria"
)

type CriterionStatus string

const (
	StatusDraft     CriterionStatus = "черновик"
	StatusPublished CriterionStatus = "опубликован"
	StatusDeleted   CriterionStatus = "удалён"
)

type WellsCriterion struct {
	CriterionID       int
	CriterionName     string
	ShortDescription  string  // как трактовать критерий у постели больного
	WellsPoints       float64 // вес критерия: +3, +1.5, +1, -2
	ScaleType         string  // ТГВ или ТЭЛА
	CriterionGroup    string  // анамнез, осмотр, пальпация, измерение
	AssessmentMethod  string  // чем выявляется критерий
	ImageKey          string  // ключ файла изображения в Minio
	VideoKey          string  // ключ файла видео в Minio
	CriterionStatus   CriterionStatus
	CreatedAt         time.Time
	LikedByPhysicians []int // ID врачей, отметивших критерий полезным
}

func (c WellsCriterion) LikeCount() int {
	return len(c.LikedByPhysicians)
}

func (c WellsCriterion) PointsLabel() string {
	text := strconv.FormatFloat(c.WellsPoints, 'g', -1, 64)
	if c.WellsPoints > 0 {
		return "+" + text
	}
	return text
}

func (c WellsCriterion) ImageURL() string {
	key := c.ImageKey
	if key == "" {
		key = "no_image"
	}
	return fmt.Sprintf("%s/%s/%s.jpg", minioPublicURL, minioBucket, key)
}

func (c WellsCriterion) VideoURL() string {
	if c.VideoKey == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s.mp4", minioPublicURL, minioBucket, c.VideoKey)
}

type Repository struct {
	criteria []WellsCriterion
}

func NewRepository() (*Repository, error) {
	criteria := seedCriteria()
	if len(criteria) == 0 {
		return nil, fmt.Errorf("коллекция критериев пуста")
	}

	return &Repository{criteria: criteria}, nil
}

func (r *Repository) GetPublishedCriteria() ([]WellsCriterion, error) {
	result := make([]WellsCriterion, 0)
	for _, criterion := range r.criteria {
		if criterion.CriterionStatus == StatusPublished {
			result = append(result, criterion)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("опубликованных критериев не найдено")
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CriterionID < result[j].CriterionID
	})

	return result, nil
}

func (r *Repository) GetCriteriaByMinPoints(minPoints float64) ([]WellsCriterion, error) {
	criteria, err := r.GetPublishedCriteria()
	if err != nil {
		return nil, err
	}

	result := make([]WellsCriterion, 0)
	for _, criterion := range criteria {
		if criterion.WellsPoints >= minPoints {
			result = append(result, criterion)
		}
	}

	return result, nil
}

func (r *Repository) GetCriterion(criterionID int) (WellsCriterion, error) {
	for _, criterion := range r.criteria {
		if criterion.CriterionID != criterionID {
			continue
		}
		if criterion.CriterionStatus == StatusDeleted {
			return WellsCriterion{}, fmt.Errorf("критерий %d удалён из справочника", criterionID)
		}
		return criterion, nil
	}

	return WellsCriterion{}, fmt.Errorf("критерий %d не найден", criterionID)
}

func (r *Repository) GetFirstCriterion() (WellsCriterion, error) {
	criteria, err := r.GetPublishedCriteria()
	if err != nil {
		return WellsCriterion{}, err
	}

	return criteria[0], nil
}

func (r *Repository) GetNextCriterion(afterCriterionID int) (WellsCriterion, error) {
	criteria, err := r.GetPublishedCriteria()
	if err != nil {
		return WellsCriterion{}, err
	}

	for _, criterion := range criteria {
		if criterion.CriterionID > afterCriterionID {
			return criterion, nil
		}
	}

	return criteria[0], nil
}

func (r *Repository) GetDraftCriterion() (WellsCriterion, error) {
	for _, criterion := range r.criteria {
		if criterion.CriterionStatus == StatusDraft {
			return criterion, nil
		}
	}

	return WellsCriterion{}, fmt.Errorf("черновик критерия не найден")
}
