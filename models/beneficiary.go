package models

// Beneficiary model
type Beneficiary struct {
	ID   int    `json:"id" gorm:"column:id"`
	Code int    `json:"code" gorm:"column:code"`
	Name string `json:"name" gorm:"column:name"`
}

// TableName ensures table name for beneficiaries
func (Beneficiary) TableName() string {
	return "beneficiary"
}
