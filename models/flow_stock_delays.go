package models

import (
	"database/sql"
	"fmt"
)

// FlowStockDelays model
type FlowStockDelays struct {
	StockCount        NullInt64   `json:"stock_count"`
	StockAverageDelay NullFloat64 `json:"stock_average_delay"`
	FlowCount         NullInt64   `json:"flow_count"`
	FlowAverageDelay  NullFloat64 `json:"flow_average_delay"`
}

// Get fetches from database flow and stock count and average delay
func (f *FlowStockDelays) Get(days int64, db *sql.DB) error {
	if err := db.QueryRow(`SELECT stock.c,stock.avg,flow.c,flow.avg FROM 
  (SELECT count(1) c,avg(CURRENT_DATE-receipt_date) 
  FROM payment_demands WHERE excluded=FALSE AND processed_date ISNULL) stock,
  (SELECT count(1) c,avg(date-receipt_date) 
	FROM payment WHERE date>CURRENT_DATE-90) flow;`).Scan(&f.StockCount,
		&f.StockAverageDelay, &f.FlowCount, &f.FlowAverageDelay); err != nil {
		return fmt.Errorf("select %v ", err)
	}
	return nil
}
