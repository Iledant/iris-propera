package models

import "database/sql"

// ImportLog model
type ImportLog struct {
	ID       int    `json:"id" gorm:"column:id"`
	Category string `json:"category" gorm:"column:category"`
	LastDate string `json:"last_date" gorm:"column:last_date"`
}

// ImportLogs embeddes an array of ImportLog for json export.
type ImportLogs struct {
	ImportLogs []ImportLog `json:"ImportLog"`
}

// GetAll fetches all import logs from database.
func (i *ImportLogs) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id,category,last_date FROM import_logs`)
	if err != nil {
		return err
	}
	var r ImportLog
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Category, &r.LastDate); err != nil {
			return err
		}
		i.ImportLogs = append(i.ImportLogs, r)
	}
	err = rows.Err()
	return err
}
