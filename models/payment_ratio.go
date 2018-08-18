package models

// PaymentRatio model
type PaymentRatio struct {
	ID            int       `json:"id" gorm:"column:id"`
	PaymentTypeID NullInt64 `json:"payment_types_id" gorm:"column:payment_types_id"`
	Ratio         float64   `json:"ratio" gorm:"column:ratio"`
	Index         int       `json:"index" gorm:"column:index"`
}

// TableName ensures table name for payment_ratios
func (u PaymentRatio) TableName() string {
	return "payment_ratios"
}
