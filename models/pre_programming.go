package models

// PreProgramming model
type PreProgramming struct {
	ID           int         `json:"id" gorm:"column:id"`
	Year         int         `json:"year" gorm:"column:year"`
	PhysicalOpID int         `json:"physical_op_id" gorm:"column:physical_op_id"`
	CommissionID int         `json:"commission_id" gorm:"column:commission_id"`
	Value        int64       `json:"value" gorm:"column:value"`
	StateRatio   NullFloat64 `json:"state_ratio" gorm:"column:state_ratio"`
	TotalValue   NullInt64   `json:"total_value" gorm:"column:total_value"`
	Descript     NullString  `json:"descript" gorm:"column:descript"`
}

// TableName ensures table name for pre_programmings
func (u PreProgramming) TableName() string {
	return "pre_programmings"
}
