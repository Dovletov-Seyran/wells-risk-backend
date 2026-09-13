package ds

type Physician struct {
	PhysicianID int    `gorm:"primaryKey;size:32;column:physician_id"`
	Login       string `gorm:"type:varchar(50);not null;unique;column:login"`
	Password    string `gorm:"type:varchar(100);not null;column:password"`
	FullName    string `gorm:"type:varchar(150);column:full_name"`
	IsModerator bool   `gorm:"not null;default:false;column:is_moderator"`
}

func (Physician) TableName() string {
	return "physicians"
}
