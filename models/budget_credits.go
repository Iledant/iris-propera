package models

// BudgetCredit model
type BudgetCredit struct {
	ID                 int       `json:"id" gorm:"column:id"`
	CommissionDate     NullTime  `json:"commission_date" gorm:"column:commission_date"`
	ChapterID          NullInt64 `json:"chapter_id" gorm:"column:chapter_id"`
	PrimaryCommitment  int64     `json:"primary_commitment" gorm:"column:primary_commitment"`
	FrozenCommitment   int64     `json:"frozen_commitment" gorm:"column:frozen_commitment"`
	ReservedCommitment int64     `json:"reserved_commitment" gorm:"column:reserved_commitment"`
}

// TableName ensures table name for budget_credits
func (u BudgetCredit) TableName() string {
	return "budget_credits"
}
