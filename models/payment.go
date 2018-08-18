package models

import (
	"time"
)

// Payment model
type Payment struct {
	ID                    int       `json:"id" gorm:"column:id"`
	FinancialCommitmentID NullInt64 `json:"financial_commitment_id" gorm:"column:financial_commitment_id"`
	CoriolisYear          string    `json:"coriolis_year" gorm:"column:coriolis_year"`
	CoriolisEgtCode       string    `json:"coriolis_egt_code" gorm:"column:coriolis_egt_code"`
	CoriolisEgtNum        string    `json:"coriolis_egt_num" gorm:"column:coriolis_egt_num"`
	CoriolisEgtLine       string    `json:"coriolis_egt_line" gorm:"column:coriolis_egt_line"`
	Date                  time.Time `json:"date" gorm:"column:date"`
	Number                string    `json:"number" gorm:"column:number"`
	Value                 int64     `json:"value" gorm:"column:value"`
	CancelledValue        int64     `json:"cancelled_value" gorm:"column:cancelled_value"`
	BeneficiaryCode       int       `json:"beneficiary_code" gorm:"column:beneficiary_code"`
}

// TableName ensures table name for payments
func (u Payment) TableName() string {
	return "payment"
}
