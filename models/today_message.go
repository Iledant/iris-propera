package models

// TodayMessage model
type TodayMessage struct {
	ID    int        `json:"id" gorm:"column:id"`
	Title NullString `json:"title" gorm:"column:title"`
	Text  NullString `json:"text" gorm:"column:text"`
}

// TableName ensures table name for today_messages
func (u TodayMessage) TableName() string {
	return "today_messages"
}
