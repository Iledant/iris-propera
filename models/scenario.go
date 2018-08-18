package models

// Scenario model
type Scenario struct {
	ID       int        `json:"id" gorm:"column:id"`
	Name     string     `json:"name" gorm:"column:name"`
	Descript NullString `json:"descript" gorm:"column:descript"`
}

// TableName ensures table name for scenario
func (u Scenario) TableName() string {
	return "scenario"
}
