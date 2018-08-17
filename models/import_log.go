package models

// ImportLog model
type ImportLog struct {
	ID       int    `json:"id" gorm:"column:id"`
	Category string `json:"category" gorm:"column:category"`
	LastDate string `json:"last_date" gorm:"column:last_date"`
}

// TableName ensures table name for import logs
func (u ImportLog) TableName() string {
	return "import_logs"
}
