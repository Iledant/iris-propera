package models

// ScenarioOffset model
type ScenarioOffset struct {
	ID           int `json:"id" gorm:"column:id"`
	ScenarioID   int `json:"scenario_id" gorm:"column:scenario_id"`
	PhysicalOpID int `json:"physical_op_id" gorm:"column:physical_op_id"`
	Offset       int `json:"offset" gorm:"column:offset"`
}

// TableName ensures table name for scenario_offset
func (u ScenarioOffset) TableName() string {
	return "scenario_offset"
}
