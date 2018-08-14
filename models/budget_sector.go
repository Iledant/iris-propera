package models

// BudgetSector model
type BudgetSector struct {
	ID   int    `json:"id" gorm:"column:id"`
	Code string `json:"code" gorm:"column:code"`
	Name string `json:"name" gorm:"column:name"`
}

// TableName ensures table name for budget_sector
func (u BudgetSector) TableName() string {
	return "budget_sector"
}
