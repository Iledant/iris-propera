package models

import (
	"database/sql"
	"fmt"
)

// WeekPaymentCount model
type WeekPaymentCount struct {
	WeekNumber     int64 `json:"week_number"`
	ReceivedNumber int64 `json:"received_number"`
	PaymentNumber  int64 `json:"payment_number"`
}

// WeekPaymentCounts embeddes an array of WeekPaymentCount for json export and
// dedicated queries
type WeekPaymentCounts struct {
	Lines []WeekPaymentCount `json:"WeekPaymentCount"`
}

// GetAll fetches all payment count par week of the given year
func (w *WeekPaymentCounts) GetAll(year int64, db *sql.DB) error {
	rows, err := db.Query(`SELECT w,COALESCE(r.n,0),COALESCE(p.n,0) FROM
	generate_series(1,52) w
	LEFT JOIN 
	(SELECT EXTRACT(week FROM receipt_date) m,count(1) n FROM payment
	WHERE EXTRACT(year FROM receipt_date)=$1 GROUP BY 1) r
	ON w=r.m
	LEFT JOIN
	(SELECT EXTRACT(week FROM date) m,count(1) n FROM payment
	WHERE EXTRACT(year FROM date)=$1 GROUP BY 1) p
	ON w=p.m
	ORDER BY 1;`, year)
	if err != nil {
		return fmt.Errorf("select %v", err)
	}
	var line WeekPaymentCount
	for rows.Next() {
		if err = rows.Scan(&line.WeekNumber, &line.ReceivedNumber, &line.PaymentNumber); err != nil {
			return fmt.Errorf("scan %v", err)
		}
		w.Lines = append(w.Lines, line)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("rows err %v", err)
	}
	if len(w.Lines) == 0 {
		w.Lines = []WeekPaymentCount{}
	}
	return nil
}
