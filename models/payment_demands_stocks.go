package models

import (
	"database/sql"
	"fmt"
	"time"
)

// PaymentDemandsStock model
type PaymentDemandsStock struct {
	Day    time.Time `json:"day"`
	Number int64     `json:"number"`
}

// PaymentDemandsStocks embeddes an array of PaymentDemandsStock for json export
// and dedicated query
type PaymentDemandsStocks struct {
	Lines []PaymentDemandsStock `json:"PaymentDemandsStock"`
}

// GetAll fetches all payment demands count for every last 30 days
func (p *PaymentDemandsStocks) GetAll(db *sql.DB) error {
	rows, err := db.Query(`SELECT d.d,count(1)
  FROM payment_demands p,(SELECT CURRENT_DATE-generate_series(0,30) d) d
  WHERE p.receipt_date<=d.d
    AND (p.processed_date ISNULL OR p.processed_date>d)
    AND p.excluded<>FALSE
 GROUP BY 1 ORDER BY 1;`)
	if err != nil {
		return fmt.Errorf("select %v", err)
	}
	var line PaymentDemandsStock
	for rows.Next() {
		if err = rows.Scan(&line.Day, &line.Number); err != nil {
			return fmt.Errorf("scan %v", err)
		}
		p.Lines = append(p.Lines, line)
	}
	if len(p.Lines) == 0 {
		p.Lines = []PaymentDemandsStock{}
	}
	return nil
}
