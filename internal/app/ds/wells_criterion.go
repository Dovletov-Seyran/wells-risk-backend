package ds

import "time"

type CriterionStatus string

const (
	StatusDraft     CriterionStatus = "черновик"
	StatusPublished CriterionStatus = "опубликован"
	StatusDeleted   CriterionStatus = "удалён"
)

type WellsCriterion struct {
	CriterionID      int             `gorm:"primaryKey;size:32;column:criterion_id" json:"criterion_id"`
	CriterionName    string          `gorm:"type:varchar(150);not null;column:criterion_name" json:"criterion_name"`
	ShortDescription string          `gorm:"type:varchar(500);column:short_description" json:"short_description"`
	CriterionStatus  CriterionStatus `gorm:"type:varchar(20);not null;column:criterion_status" json:"-"`
	ImageKey         string          `gorm:"type:varchar(255);column:image_key" json:"-"`
	VideoKey         string          `gorm:"type:varchar(255);column:video_key" json:"-"`
	WellsPoints      float64         `gorm:"type:numeric(3,1);column:wells_points" json:"wells_points"`
	CriterionGroup   string          `gorm:"type:varchar(50);column:criterion_group" json:"criterion_group"`
	CreatedAt        time.Time       `gorm:"type:timestamp;not null;column:created_at" json:"created_at"`
	FormedAt         *time.Time      `gorm:"type:timestamp;column:formed_at" json:"formed_at"`

	CreatorID *int       `gorm:"size:32;column:creator_id" json:"-"`
	Creator   *Physician `gorm:"foreignKey:CreatorID;references:PhysicianID;constraint:OnDelete:NO ACTION" json:"-"`

	LikeCount int  `gorm:"-" json:"-"`
	LikedByMe bool `gorm:"-" json:"-"`
}

func (WellsCriterion) TableName() string {
	return "wells_criteria"
}
