package models

// Step model
type Step struct {
	ID   int    `json:"id" gorm:"column:id"`
	Name string `json:"name" gorm:"column:name"`
}

// TableName ensures table name for step
func (u Step) TableName() string {
	return "step"
}
