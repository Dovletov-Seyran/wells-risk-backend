package ds

type CriterionLike struct {
	LikeID           int `gorm:"primaryKey;size:32;column:like_id" json:"like_id"`
	LikedCriterionID int `gorm:"not null;size:32;column:criterion_id;uniqueIndex:idx_like_criterion_physician" json:"criterion_id"`
	LikedByID        int `gorm:"not null;size:32;column:physician_id;uniqueIndex:idx_like_criterion_physician" json:"physician_id"`

	Criterion *WellsCriterion `gorm:"foreignKey:LikedCriterionID;references:CriterionID;constraint:OnDelete:NO ACTION" json:"-"`
	Physician *Physician      `gorm:"foreignKey:LikedByID;references:PhysicianID;constraint:OnDelete:NO ACTION" json:"-"`
}

func (CriterionLike) TableName() string {
	return "criterion_likes"
}
