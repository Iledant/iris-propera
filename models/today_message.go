package models

import (
	"database/sql"
	"errors"
)

// TodayMessage model
type TodayMessage struct {
	ID    int        `json:"id" gorm:"column:id"`
	Title NullString `json:"title" gorm:"column:title"`
	Text  NullString `json:"text" gorm:"column:text"`
}

// Get fetches the first entry of today messages from database.
func (t *TodayMessage) Get(db *sql.DB) (err error) {
	err = db.QueryRow(`SELECT id,title,text FROM today_messages ORDER BY 1 LIMIT 1`).
		Scan(&t.ID, &t.Title, &t.Text)
	return err
}

// Update modifies the first entry of today messages in database.
func (t *TodayMessage) Update(db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE today_messages SET title = $1, text = $2 WHERE id = 1`,
		t.Title, t.Text)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("today message introuvable")
	}
	t.ID = 1
	return err
}
