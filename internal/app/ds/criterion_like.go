package ds

type CriterionLike struct {
	LikeID           int `gorm:"primaryKey;size:32;column:like_id"`
	LikedCriterionID int `gorm:"not null;size:32;column:criterion_id;uniqueIndex:idx_like_criterion_physician"`
	LikedByID        int `gorm:"not null;size:32;column:physician_id;uniqueIndex:idx_like_criterion_physician"`

	Criterion *WellsCriterion `gorm:"foreignKey:LikedCriterionID;references:CriterionID;constraint:OnDelete:NO ACTION"`
	Physician *Physician      `gorm:"foreignKey:LikedByID;references:PhysicianID;constraint:OnDelete:NO ACTION"`
}

func (CriterionLike) TableName() string {
	return "criterion_likes"
}
