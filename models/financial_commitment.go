package models

import (
	"time"
)

// FinancialCommitment model
type FinancialCommitment struct {
	ID              int       `json:"id" gorm:"column:id"`
	PhysicalOpID    NullInt64 `json:"physical_op_id" gorm:"column:physical_op_id"`
	PlanLineID      NullInt64 `json:"plan_line_id" gorm:"column:plan_line_id"`
	Chapter         string    `json:"chapter" gorm:"column:chapter"`
	Action          string    `json:"action" gorm:"column:action"`
	IrisCode        string    `json:"iris_code" gorm:"column:iris_code"`
	CoriolisYear    string    `json:"coriolis_year" gorm:"column:coriolis_year"`
	CoriolisEgtCode string    `json:"coriolis_egt_code" gorm:"column:coriolis_egt_code"`
	CoriolisEgtNum  string    `json:"coriolis_egt_num" gorm:"column:coriolis_egt_num"`
	CoriolisEgtLine string    `json:"coriolis_egt_line" gorm:"column:coriolis_egt_line"`
	Name            string    `json:"name" gorm:"column:name"`
	BeneficiaryCode int       `json:"beneficiary_code" gorm:"column:beneficiary_code"`
	Date            time.Time `json:"date" gorm:"column:date"`
	Value           int64     `json:"value" gorm:"column:value"`
	ActionID        NullInt64 `json:"action_id" gorm:"column:action_id"`
	LapseDate       NullTime  `json:"lapse_date" gorm:"column:lapse_date"`
}

// TableName ensures table name for financial_commitment
func (u FinancialCommitment) TableName() string {
	return "financial_commitment"
}
