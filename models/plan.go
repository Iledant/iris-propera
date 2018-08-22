package models

// Plan model
type Plan struct {
	ID        int64      `json:"id" gorm:"column:id"`
	Name      string     `json:"name" gorm:"column:name"`
	Descript  NullString `json:"descript" gorm:"column:descript"`
	FirstYear NullInt64  `json:"first_year" gorm:"column:first_year"`
	LastYear  NullInt64  `json:"last_year" gorm:"column:last_year"`
}

// TableName ensures table name for plan
func (u Plan) TableName() string {
	return "plan"
}
