package model

type UserRole struct {
	UserID uint `json:"user_id" gorm:"column:user_id;unique_index:idx_user_role"`
	RoleID uint `json:"role_id" gorm:"column:role_id;unique_index:idx_user_role"`
}

func (UserRole) TableName() string {
	return "user_role"
}
