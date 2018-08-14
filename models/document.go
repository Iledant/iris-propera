package models

// Document model
type Document struct {
	ID           int    `json:"id" gorm:"column:id"`
	PhysicalOpID int    `json:"physical_op_id" gorm:"column:physical_op_id"`
	Name         string `json:"name" gorm:"column:name"`
	Link         string `json:"link" gorm:"column:link"`
}

// TableName ensures table name for documents
func (u Document) TableName() string {
	return "documents"
}
