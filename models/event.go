package models

import (
	"time"
)

// Event model
type Event struct {
	ID           int        `json:"id" gorm:"column:id"`
	PhysicalOpID int        `json:"physical_op_id" gorm:"column:physical_op_id"`
	Name         string     `json:"name" gorm:"column:name"`
	Date         time.Time  `json:"date" gorm:"column:date"`
	IsCertain    bool       `json:"iscertain" gorm:"column:iscertain"`
	Descript     NullString `json:"descript" gorm:"column:descript"`
}

// TableName ensures table name for events
func (u Event) TableName() string {
	return "event"
}
