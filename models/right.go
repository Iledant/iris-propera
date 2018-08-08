package models

// Right model for the right of a user on physical operations.
type Right struct {
	ID           int `json:"id" gorm:"column:id"`
	PhysicalOpID int `json:"physical_op_id" gorm:"column:physical_op_id"`
	UserID       int `json:"users_id" gorm:"column:users_id"`
}

// TableName ensures table name for rights
func (Right) TableName() string {
	return "rights"
}
