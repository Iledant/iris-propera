package models

import (
	"database/sql"
	"fmt"
	"sync"
)

// CsfWeekTrend model fetches the trend of payment demands with no csf from one
// day to another
type CsfWeekTrend struct {
	LastWeekCount NullInt64
	ThisWeekCount NullInt64
}

var cwt CsfWeekTrend

func (c *CsfWeekTrend) copy(src *CsfWeekTrend) {
	c.LastWeekCount.Valid = src.LastWeekCount.Valid
	c.LastWeekCount.Int64 = src.LastWeekCount.Int64
	c.ThisWeekCount.Valid = src.ThisWeekCount.Valid
	c.ThisWeekCount.Int64 = src.ThisWeekCount.Int64
}

// Get fetches count of payments demands with no csf from last and current week
// from database
func (c *CsfWeekTrend) Get(db *sql.DB) error {
	if !needUpdate(csfWeekTrendUpdate, paymentDemandsUpdate) {
		c.copy(&cwt)
		return nil
	}
	if err := db.QueryRow(`SELECT last_week.c,this_week.c
	 FROM (SELECT count(1) c FROM payment_demands 
	WHERE receipt_date<= CURRENT_DATE-7 AND excluded!=TRUE
		AND (csf_date ISNULL OR csf_date>= CURRENT_DATE-7) ) last_week,
(SELECT count(1) c FROM payment_demands 
	WHERE excluded!=TRUE AND csf_date ISNULL) this_week`).Scan(&c.LastWeekCount,
		&c.ThisWeekCount); err != nil {
		return fmt.Errorf("select %v", err)
	}
	var mutex = &sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()
	cwt.copy(c)
	update(csfWeekTrendUpdate)
	return nil
}
