package model

type UserProject struct {
	UserID    uint `json:"user_id" gorm:"column:user_id;unique_index:idx_user_project"`
	ProjectID uint `json:"project_id" gorm:"column:project_id;unique_index:idx_user_project"`
}

func (UserProject) TableName() string {
	return "user_project"
}
