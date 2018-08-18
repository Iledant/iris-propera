package models

// Programming model
type Programming struct {
	ID           int         `json:"id" gorm:"column:id"`
	Value        int64       `json:"value" gorm:"column:value"`
	PhysicalOpID int         `json:"physical_op_id" gorm:"column:physical_op_id"`
	CommissionID int         `json:"commission_id" gorm:"column:commission_id"`
	Year         NullInt64   `json:"year" gorm:"column:year"`
	StateRatio   NullFloat64 `json:"state_ratio" gorm:"column:state_ratio"`
	TotalValue   NullInt64   `json:"total_value" gorm:"column:total_value"`
}

// TableName ensures table name for programmings
func (u Programming) TableName() string {
	return "programmings"
}
