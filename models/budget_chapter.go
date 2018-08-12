package models

// BudgetChapter model
type BudgetChapter struct {
	ID   int    `json:"id" gorm:"column:id"`
	Code int    `json:"code" gorm:"column:code"`
	Name string `json:"name" gorm:"column:name"`
}

// TableName ensures table name for budget_chapter
func (u BudgetChapter) TableName() string {
	return "budget_chapter"
}
