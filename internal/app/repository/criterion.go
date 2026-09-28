package repository

import (
	"fmt"
	"time"

	"wells-risk-backend/internal/app/ds"
)

func (r *Repository) GetPublishedCriteria(minPoints *float64, physicianID int) ([]ds.WellsCriterion, error) {
	var criteria []ds.WellsCriterion

	query := r.db.Where("criterion_status = ?", ds.StatusPublished)

	if minPoints != nil {
		query = query.Where("wells_points >= ?", *minPoints)
	}

	if err := query.Order("criterion_id").Find(&criteria).Error; err != nil {
		return nil, err
	}

	if err := r.fillLikes(criteria, physicianID); err != nil {
		return nil, err
	}

	return criteria, nil
}

// GetPublishedCriterion возвращает один опубликованный критерий для ленты.
func (r *Repository) GetPublishedCriterion(criterionID, physicianID int) (ds.WellsCriterion, error) {
	var criterion ds.WellsCriterion

	err := r.db.
		Where("criterion_id = ? AND criterion_status = ?", criterionID, ds.StatusPublished).
		First(&criterion).Error
	if err != nil {
		return ds.WellsCriterion{}, err
	}

	return r.withLikes(criterion, physicianID)
}

// GetFirstCriterion — первый опубликованный критерий ленты.
func (r *Repository) GetFirstCriterion(physicianID int) (ds.WellsCriterion, error) {
	var criterion ds.WellsCriterion

	err := r.db.
		Where("criterion_status = ?", ds.StatusPublished).
		Order("criterion_id").
		First(&criterion).Error
	if err != nil {
		return ds.WellsCriterion{}, err
	}

	return r.withLikes(criterion, physicianID)
}

// GetNextCriterion — следующий критерий ленты
func (r *Repository) GetNextCriterion(afterCriterionID, physicianID int) (ds.WellsCriterion, error) {
	var criterion ds.WellsCriterion

	err := r.db.
		Where("criterion_status = ? AND criterion_id > ?", ds.StatusPublished, afterCriterionID).
		Order("criterion_id").
		First(&criterion).Error
	if err != nil {
		return r.GetFirstCriterion(physicianID)
	}

	return r.withLikes(criterion, physicianID)
}

// GetDraftCriterion — черновик текущего врача
func (r *Repository) GetDraftCriterion(physicianID int) (ds.WellsCriterion, error) {
	var criterion ds.WellsCriterion

	err := r.db.
		Where("criterion_status = ? AND creator_id = ?", ds.StatusDraft, physicianID).
		Order("criterion_id").
		First(&criterion).Error
	if err != nil {
		return ds.WellsCriterion{}, err
	}

	return r.withLikes(criterion, physicianID)
}

// CountDrafts считает черновики врача: их не может быть больше одного.
func (r *Repository) CountDrafts(physicianID int) (int64, error) {
	var drafts int64

	err := r.db.Model(&ds.WellsCriterion{}).
		Where("criterion_status = ? AND creator_id = ?", ds.StatusDraft, physicianID).
		Count(&drafts).Error

	return drafts, err
}

// CreateCriterion добавляет новую карточку критерия в статусе черновик.
func (r *Repository) CreateCriterion(criterion *ds.WellsCriterion) error {
	return r.db.Create(criterion).Error
}

// PublishCriterion переводит черновик врача в статус опубликован.
func (r *Repository) PublishCriterion(criterionID, physicianID int) error {
	now := time.Now()

	result := r.db.Model(&ds.WellsCriterion{}).
		Where("criterion_id = ? AND creator_id = ? AND criterion_status = ?",
			criterionID, physicianID, ds.StatusDraft).
		Updates(map[string]interface{}{
			"criterion_status": ds.StatusPublished,
			"formed_at":        now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("черновик критерия %d у врача %d не найден", criterionID, physicianID)
	}

	return nil
}

// DeleteCriterion — логическое удаление критерия одним SQL-запросом без ORM.
func (r *Repository) DeleteCriterion(criterionID, physicianID int) error {
	result := r.db.Exec(
		"UPDATE wells_criteria SET criterion_status = ? WHERE criterion_id = ? AND creator_id = ? AND criterion_status <> ?",
		ds.StatusDeleted, criterionID, physicianID, ds.StatusDeleted,
	)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("критерий %d у врача %d не найден", criterionID, physicianID)
	}

	return nil
}

// GetCriterion возвращает критерий врача в любом статусе, кроме удалённого.
func (r *Repository) GetCriterion(criterionID, physicianID int) (ds.WellsCriterion, error) {
	var criterion ds.WellsCriterion

	err := r.db.
		Where("criterion_id = ? AND criterion_status <> ?", criterionID, ds.StatusDeleted).
		First(&criterion).Error
	if err != nil {
		return ds.WellsCriterion{}, err
	}

	return r.withLikes(criterion, physicianID)
}

// SetLike ставит отметку врача критерию
func (r *Repository) SetLike(criterionID, physicianID, value int) error {
	if value == 0 {
		return r.db.
			Where("criterion_id = ? AND physician_id = ?", criterionID, physicianID).
			Delete(&ds.CriterionLike{}).Error
	}

	like := ds.CriterionLike{
		LikedCriterionID: criterionID,
		LikedByID:        physicianID,
	}

	return r.db.
		Where("criterion_id = ? AND physician_id = ?", criterionID, physicianID).
		FirstOrCreate(&like).Error
}

// likeStat — строка агрегата по таблице отметок.
type likeStat struct {
	CriterionID int
	LikeCount   int
	LikedByMe   int
}

// fillLikes одним запросом считает отметки для списка критериев.
func (r *Repository) fillLikes(criteria []ds.WellsCriterion, physicianID int) error {
	if len(criteria) == 0 {
		return nil
	}

	criterionIDs := make([]int, 0, len(criteria))
	for _, criterion := range criteria {
		criterionIDs = append(criterionIDs, criterion.CriterionID)
	}

	var stats []likeStat

	err := r.db.Model(&ds.CriterionLike{}).
		Select("criterion_id, COUNT(*) AS like_count, COUNT(CASE WHEN physician_id = ? THEN 1 END) AS liked_by_me", physicianID).
		Where("criterion_id IN ?", criterionIDs).
		Group("criterion_id").
		Scan(&stats).Error
	if err != nil {
		return err
	}

	byCriterion := make(map[int]likeStat, len(stats))
	for _, stat := range stats {
		byCriterion[stat.CriterionID] = stat
	}

	for i := range criteria {
		stat := byCriterion[criteria[i].CriterionID]
		criteria[i].LikeCount = stat.LikeCount
		criteria[i].LikedByMe = stat.LikedByMe > 0
	}

	return nil
}

// withLikes считает отметки для одного критерия.
func (r *Repository) withLikes(criterion ds.WellsCriterion, physicianID int) (ds.WellsCriterion, error) {
	criteria := []ds.WellsCriterion{criterion}

	if err := r.fillLikes(criteria, physicianID); err != nil {
		return ds.WellsCriterion{}, err
	}

	return criteria[0], nil
}
