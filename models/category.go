package models

// Category model
type Category struct {
	ID   int    `json:"id" gorm:"column:id"`
	Name string `json:"name" gorm:"column:name"`
}

// TableName ensures table name for category
func (u Category) TableName() string {
	return "category"
}
