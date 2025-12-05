package model

type UserTeam struct {
	UserID uint `json:"user_id" gorm:"column:user_id;unique_index:idx_user_team"`
	TeamID uint `json:"team_id" gorm:"column:team_id;unique_index:idx_user_team"`
}

func (UserTeam) TableName() string {
	return "user_team"
}
