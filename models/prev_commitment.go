package models

// PrevCommitment model
type PrevCommitment struct {
	ID           int         `json:"id" gorm:"column:id"`
	PhysicalOpID int         `json:"physical_op_id" gorm:"column:physical_op_id"`
	Year         int         `json:"year" gorm:"column:year"`
	Value        int64       `json:"value" gorm:"column:value"`
	Descript     NullString  `json:"descript" gorm:"column:descript"`
	StateRatio   NullFloat64 `json:"state_ratio" gorm:"column:state_ratio"`
	TotalValue   NullInt64   `json:"total_value" gorm:"column:total_value"`
}

// TableName ensures table name for prev_commitment
func (u PrevCommitment) TableName() string {
	return "prev_commitment"
}
