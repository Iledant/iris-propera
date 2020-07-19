package models

import (
	"database/sql"
	"fmt"
	"sync"
)

// PaymentRate model is used to calculate the ratio of payment over the last
// years and the ratio of the actual year payment compared to available payments
type PaymentRate struct {
	PastRate   NullFloat64 `json:"PastRate"`
	ActualRate NullFloat64 `json:"ActualRate"`
}

var pr PaymentRate

func (p *PaymentRate) copy(src *PaymentRate) {
	p.PastRate = src.PastRate
	p.ActualRate = src.ActualRate
}

// Get fetches the ratios from database using the cache technique to avoid
// launching unnecessary queries
func (p *PaymentRate) Get(db *sql.DB) error {
	if !needUpdate(paymentRateUpdate, paymentDemandsUpdate, everyDayUpdate) {
		p.copy(&pr)
		return nil
	}
	query := `SELECT past_dow_payment.s/past_payment.s,actual_payment.s/available_credits.s
	FROM
	(SELECT sum(value) s FROM payment 
		WHERE extract(year from date)<EXTRACT(year FROM current_date) 
			AND extract(doy from date)<extract(doy from current_date)) past_dow_payment,
	(SELECT sum(value) s FROM payment 
		WHERE extract(year from date)<EXTRACT(year FROM current_date)) past_payment,
	(SELECT SUM(pc.primitive)+SUM(pc.added)+SUM(pc.modified)+SUM(pc.movement) s FROM payment_credit pc
		JOIN budget_chapter bc ON pc.chapter_id=bc.id
	WHERE pc.year=2020 AND code IN (907,908)) available_credits,
	(SELECT sum(value) s FROM payment 
		WHERE extract(year from date)=EXTRACT(year FROM current_date)) actual_payment;`
	if err := db.QueryRow(query).Scan(&p.PastRate, &p.ActualRate); err != nil {
		return fmt.Errorf("select %v", err)
	}
	var mutex = &sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()
	pr.copy(p)
	update(paymentRateUpdate)
	return nil
}
