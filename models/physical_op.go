package models

// PhysicalOp is the model for physical operations. Number is unique.
type PhysicalOp struct {
	ID             int        `json:"id" db:"id"`
	Number         string     `json:"number" gorm:"column:number"`
	Name           string     `json:"name" gorm:"column:name"`
	Descript       NullString `json:"descript" gorm:"column:descript"`
	Isr            bool       `json:"isr" gorm:"column:isr"`
	Value          NullInt64  `json:"value" gorm:"column:value"`
	ValueDate      NullTime   `json:"valuedate" gorm:"column:valuedate"`
	Length         NullInt64  `json:"length" gorm:"column:length"`
	TRI            NullInt64  `json:"tri" gorm:"column:tri"`
	VAN            NullInt64  `json:"van" gorm:"column:van"`
	BudgetActionID NullInt64  `json:"budget_action_id" gorm:"column:budget_action_id"`
	PaymentTypeID  NullInt64  `json:"payment_types_id" gorm:"column:payment_types_id"`
	PlanLineID     NullInt64  `json:"plan_line_id" gorm:"column:plan_line_id"`
	StepID         NullInt64  `json:"step_id" gorm:"column:step_id"`
	CategoryID     NullInt64  `json:"category_id" gorm:"column:category_id"`
}

// TableName ensures the correct table name for physical operations.
func (PhysicalOp) TableName() string {
	return "physical_op"
}
