package models

// BudgetProgram model
type BudgetProgram struct {
	ID              int        `json:"id" gorm:"column:id"`
	CodeContract    string     `json:"code_contract" gorm:"column:code_contract"`
	CodeFunction    string     `json:"code_function" gorm:"column:code_function"`
	CodeNumber      string     `json:"code_number" gorm:"column:code_number"`
	CodeSubfunction NullString `json:"code_subfunction" gorm:"column:code_subfunction"`
	Name            string     `json:"name" gorm:"column:name"`
	ChapterID       int        `json:"chapter_id" gorm:"column:chapter_id"`
}

// TableName ensures table name for budget_program
func (u BudgetProgram) TableName() string {
	return "budget_program"
}
