package models

// PrevPayment model
type PrevPayment struct {
	ID           int        `json:"id" gorm:"column:id"`
	PhysicalOpID int        `json:"physical_op_id" gorm:"column:physical_op_id"`
	Year         int        `json:"year" gorm:"column:year"`
	Value        int64      `json:"value" gorm:"column:value"`
	Descript     NullString `json:"descript" gorm:"column:descript"`
}

// TableName ensures table name for prev_payment
func (u PrevPayment) TableName() string {
	return "prev_payment"
}
