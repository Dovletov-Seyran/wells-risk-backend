package repository

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"wells-risk-backend/internal/app/ds"
)

const currentPhysicianID = 1

type Repository struct {
	db *gorm.DB
}

// New открывает подключение к PostgreSQL по строке подключения.
func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}

// GetPublishedCriteria возвращает опубликованные критерии для плитки.
func (r *Repository) GetPublishedCriteria() ([]ds.WellsCriterion, error) {
	var criteria []ds.WellsCriterion

	err := r.db.
		Where("criterion_status = ?", ds.StatusPublished).
		Order("criterion_id").
		Find(&criteria).Error
	if err != nil {
		return nil, err
	}

	if err := r.fillLikeCounts(criteria); err != nil {
		return nil, err
	}

	return criteria, nil
}

// GetCriteriaByMinPoints — поиск: опубликованные критерии весом не ниже заданного.
func (r *Repository) GetCriteriaByMinPoints(minPoints float64) ([]ds.WellsCriterion, error) {
	var criteria []ds.WellsCriterion

	err := r.db.
		Where("criterion_status = ? AND wells_points >= ?", ds.StatusPublished, minPoints).
		Order("criterion_id").
		Find(&criteria).Error
	if err != nil {
		return nil, err
	}

	if err := r.fillLikeCounts(criteria); err != nil {
		return nil, err
	}

	return criteria, nil
}

// GetCriterion возвращает один критерий
func (r *Repository) GetCriterion(criterionID int) (ds.WellsCriterion, error) {
	var criterion ds.WellsCriterion

	err := r.db.
		Where("criterion_id = ? AND criterion_status <> ?", criterionID, ds.StatusDeleted).
		First(&criterion).Error
	if err != nil {
		return ds.WellsCriterion{}, err
	}

	return r.withLikeCount(criterion)
}

// GetFirstCriterion — первый опубликованный критерий
func (r *Repository) GetFirstCriterion() (ds.WellsCriterion, error) {
	var criterion ds.WellsCriterion

	err := r.db.
		Where("criterion_status = ?", ds.StatusPublished).
		Order("criterion_id").
		First(&criterion).Error
	if err != nil {
		return ds.WellsCriterion{}, err
	}

	return r.withLikeCount(criterion)
}

// GetNextCriterion — следующий критерий в ленте
func (r *Repository) GetNextCriterion(afterCriterionID int) (ds.WellsCriterion, error) {
	var criterion ds.WellsCriterion

	err := r.db.
		Where("criterion_status = ? AND criterion_id > ?", ds.StatusPublished, afterCriterionID).
		Order("criterion_id").
		First(&criterion).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.GetFirstCriterion()
	}
	if err != nil {
		return ds.WellsCriterion{}, err
	}

	return r.withLikeCount(criterion)
}

// GetDraftCriterion — черновик текущего врача
func (r *Repository) GetDraftCriterion() (ds.WellsCriterion, error) {
	var criterion ds.WellsCriterion

	err := r.db.
		Where("criterion_status = ? AND creator_id = ?", ds.StatusDraft, currentPhysicianID).
		Order("criterion_id").
		First(&criterion).Error
	if err != nil {
		return ds.WellsCriterion{}, err
	}

	return r.withLikeCount(criterion)
}

// CreateCriterion — INSERT новой карточки
func (r *Repository) CreateCriterion(criterion *ds.WellsCriterion) error {
	var drafts int64

	err := r.db.Model(&ds.WellsCriterion{}).
		Where("criterion_status = ? AND creator_id = ?", ds.StatusDraft, criterion.CreatorID).
		Count(&drafts).Error
	if err != nil {
		return err
	}
	if drafts > 0 {
		return fmt.Errorf("у врача уже есть черновик критерия")
	}

	return r.db.Create(criterion).Error
}

// PublishCriterion - UPDATE полей
func (r *Repository) PublishCriterion(criterionID int, draft ds.WellsCriterion) error {
	now := time.Now()

	result := r.db.Model(&ds.WellsCriterion{}).
		Where("criterion_id = ? AND criterion_status = ?", criterionID, ds.StatusDraft).
		Updates(map[string]interface{}{
			"criterion_name":    draft.CriterionName,
			"short_description": draft.ShortDescription,
			"wells_points":      draft.WellsPoints,
			"criterion_group":   draft.CriterionGroup,
			"image_key":         draft.ImageKey,
			"video_key":         draft.VideoKey,
			"criterion_status":  ds.StatusPublished,
			"formed_at":         now,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("черновик критерия %d не найден", criterionID)
	}

	return nil
}

// DeleteCriterion — удаление без ORM
func (r *Repository) DeleteCriterion(criterionID int) error {
	return r.db.Exec(
		"UPDATE wells_criteria SET criterion_status = $1 WHERE criterion_id = $2",
		ds.StatusDeleted, criterionID,
	).Error
}

// countLikes считает отметки врачей
func (r *Repository) countLikes(criterionID int) (int, error) {
	var count int64

	err := r.db.Model(&ds.CriterionLike{}).
		Where("criterion_id = ?", criterionID).
		Count(&count).Error

	return int(count), err
}

func (r *Repository) withLikeCount(criterion ds.WellsCriterion) (ds.WellsCriterion, error) {
	count, err := r.countLikes(criterion.CriterionID)
	if err != nil {
		return ds.WellsCriterion{}, err
	}

	criterion.LikeCount = count
	return criterion, nil
}

func (r *Repository) fillLikeCounts(criteria []ds.WellsCriterion) error {
	for i := range criteria {
		count, err := r.countLikes(criteria[i].CriterionID)
		if err != nil {
			return err
		}
		criteria[i].LikeCount = count
	}
	return nil
}
