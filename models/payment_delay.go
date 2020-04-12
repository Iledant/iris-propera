package models

import (
	"database/sql"
	"fmt"
	"time"
)

// PaymentDelay model used to query the database for statistics about delays
// between receipt date and payment date
type PaymentDelay struct {
	Delay  int64 `json:"delay"`
	Number int64 `json:"number"`
}

// PaymentDelays is used for json export ans dedicated queries
type PaymentDelays struct {
	Lines []PaymentDelay `json:"payment_delay"`
}

// GetSome fetches all payment delays from database for payment newer than the
// given date
func (p *PaymentDelays) GetSome(after time.Time, db *sql.DB) error {
	query := `SELECT d.d,count(1) FROM payment p,
	 (SELECT * FROM (VALUES (15),(30),(45),(60),(75),(90),(105),(120),(135),(180),(365),(730),(3000)) as d (d)) d
	WHERE p.date-p.receipt_date<= d.d AND p.date>=$1
	GROUP BY 1 ORDER BY 1`
	rows, err := db.Query(query, after)
	if err != nil {
		return fmt.Errorf("select %v", err)
	}
	var l PaymentDelay
	for rows.Next() {
		if err := rows.Scan(&l.Delay, &l.Number); err != nil {
			return fmt.Errorf("scan %v", err)
		}
		p.Lines = append(p.Lines, l)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows err %v", err)
	}
	if len(p.Lines) == 0 {
		p.Lines = []PaymentDelay{}
	}
	return nil
}
