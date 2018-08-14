package models

import "time"

// Commission model
type Commission struct {
	ID   int       `json:"id" gorm:"column:id"`
	Date time.Time `json:"date" gorm:"column:date"`
	Name string    `json:"name" gorm:"column:name"`
}

// TableName ensures table name for commissions
func (u Commission) TableName() string {
	return "commissions"
}
