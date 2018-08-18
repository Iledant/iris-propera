package models

// PlanLineRatio model
type PlanLineRatio struct {
	ID            int     `json:"id" gorm:"column:id"`
	PlanLineID    int     `json:"plan_line_id" gorm:"column:plan_line_id"`
	BeneficiaryID int     `json:"beneficiary_id" gorm:"column:beneficiary_id"`
	Ratio         float64 `json:"ratio" gorm:"column:ratio"`
}

// TableName ensures table name for plan_line_ratios
func (u PlanLineRatio) TableName() string {
	return "plan_line_ratios"
}
