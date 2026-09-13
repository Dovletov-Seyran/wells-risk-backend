package ds

import (
	"strconv"
	"time"
)

type CriterionStatus string

const (
	StatusDraft     CriterionStatus = "черновик"
	StatusPublished CriterionStatus = "опубликован"
	StatusDeleted   CriterionStatus = "удалён"
)

type WellsCriterion struct {
	CriterionID      int             `gorm:"primaryKey;size:32;column:criterion_id"`
	CriterionName    string          `gorm:"type:varchar(150);not null;column:criterion_name"`
	ShortDescription string          `gorm:"type:varchar(500);column:short_description"`
	CriterionStatus  CriterionStatus `gorm:"type:varchar(20);not null;column:criterion_status"`
	ImageURL         string          `gorm:"type:varchar(255);column:image_url"`
	VideoURL         string          `gorm:"type:varchar(255);column:video_url"`
	WellsPoints      float64         `gorm:"type:numeric(3,1);column:wells_points"`
	CriterionGroup   string          `gorm:"type:varchar(50);column:criterion_group"`
	CreatedAt        time.Time       `gorm:"type:timestamp;not null;column:created_at"`
	FormedAt         *time.Time      `gorm:"type:timestamp;column:formed_at"`

	CreatorID *int       `gorm:"size:32;column:creator_id"`
	Creator   *Physician `gorm:"foreignKey:CreatorID;references:PhysicianID;constraint:OnDelete:NO ACTION"`

	LikeCount int `gorm:"-"`
}

func (WellsCriterion) TableName() string {
	return "wells_criteria"
}

func (c WellsCriterion) PointsLabel() string {
	text := strconv.FormatFloat(c.WellsPoints, 'g', -1, 64)
	if c.WellsPoints > 0 {
		return "+" + text
	}
	return text
}
