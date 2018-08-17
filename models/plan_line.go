package models

// PlanLine model
type PlanLine struct {
	ID         int        `json:"id" gorm:"column:id"`
	PlanID     int        `json:"plan_id" gorm:"column:plan_id"`
	Name       string     `json:"name" gorm:"column:name"`
	Descript   NullString `json:"descript" gorm:"column:descript"`
	Value      int64      `json:"value" gorm:"column:value"`
	TotalValue NullInt64  `json:"total_value" gorm:"column:total_value"`
}

// TableName ensures table name for plan_line
func (u PlanLine) TableName() string {
	return "plan_line"
}
