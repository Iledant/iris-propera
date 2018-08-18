package models

import (
	"time"
)

// PendingCommitment model
type PendingCommitment struct {
	ID             int       `json:"id" gorm:"column:id"`
	PhysicalOpID   NullInt64 `json:"physical_op_id" gorm:"column:physical_op_id"`
	IrisCode       string    `json:"iris_code" gorm:"column:iris_code"`
	Name           string    `json:"name" gorm:"column:name"`
	Chapter        string    `json:"chapter" gorm:"column:chapter"`
	ProposedValue  int64     `json:"proposed_value" gorm:"column:proposed_value"`
	Action         string    `json:"action" gorm:"column:action"`
	CommissionDate time.Time `json:"commission_date" gorm:"column:commission_date"`
	Beneficiary    string    `json:"beneficiary" gorm:"column:beneficiary"`
}

// TableName ensures table name for pending_commitments
func (u PendingCommitment) TableName() string {
	return "pending_commitments"
}
