package models

// BudgetAction model
type BudgetAction struct {
	ID        int    `json:"id" gorm:"column:id"`
	Code      string `json:"code" gorm:"column:code"`
	Name      string `json:"name" gorm:"column:name"`
	ProgramID int    `json:"program_id" gorm:"column:program_id"`
	SectorID  int    `json:"sector_id" gorm:"column:sector_id"`
}

// TableName ensures table name for budget_action
func (u BudgetAction) TableName() string {
	return "budget_action"
}
