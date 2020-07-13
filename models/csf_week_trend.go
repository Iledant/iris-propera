package models

import (
	"database/sql"
	"fmt"
)

// CsfWeekTrend model fetches the trend of payment demands with no csf from one
// day to another
type CsfWeekTrend struct {
	LastWeekCount NullInt64
	ThisWeekCount NullInt64
}

// Get fetches count of payments demands with no csf from last and current week
// from database
func (c *CsfWeekTrend) Get(db *sql.DB) error {
	if err := db.QueryRow(`SELECT last_week.c,this_week.c
	 FROM (SELECT count(1) c FROM payment_demands 
	WHERE receipt_date<= CURRENT_DATE-7 AND excluded!=TRUE
		AND (csf_date ISNULL OR csf_date>= CURRENT_DATE-7) ) last_week,
(SELECT count(1) c FROM payment_demands 
	WHERE excluded!=TRUE AND csf_date ISNULL) this_week`).Scan(&c.LastWeekCount,
		&c.ThisWeekCount); err != nil {
		return fmt.Errorf("select %v", err)
	}
	return nil
}
