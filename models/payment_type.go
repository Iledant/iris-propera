package models

// PaymentType model
type PaymentType struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

// TableName ensures table name for payments_types
func (u PaymentType) TableName() string {
	return "payment_types"
}
