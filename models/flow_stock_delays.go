package models

import (
	"database/sql"
	"fmt"
	"sync"
)

// FlowStockDelays model
type FlowStockDelays struct {
	ActualStockCount        NullInt64   `json:"ActualStockCount"`
	ActualStockAverageDelay NullFloat64 `json:"ActualStockAverageDelay"`
	ActualFlowCount         NullInt64   `json:"ActualFlowCount"`
	ActualFlowAverageDelay  NullFloat64 `json:"ActualFlowAverageDelay"`
	FormerStockCount        NullInt64   `json:"FormerStockCount"`
	FormerStockAverageDelay NullFloat64 `json:"FormerStockAverageDelay"`
	FormerFlowCount         NullInt64   `json:"FormerFlowCount"`
	FormerFlowAverageDelay  NullFloat64 `json:"FormerFlowAverageDelay"`
}

var fsd FlowStockDelays

func (f *FlowStockDelays) copy(src *FlowStockDelays) {
	f.ActualStockCount.Valid = src.ActualStockCount.Valid
	f.ActualStockCount.Int64 = src.ActualStockCount.Int64
	f.ActualStockAverageDelay.Valid = src.ActualStockAverageDelay.Valid
	f.ActualStockAverageDelay.Float64 = src.ActualStockAverageDelay.Float64
	f.ActualFlowCount.Valid = src.ActualFlowCount.Valid
	f.ActualFlowCount.Int64 = src.ActualFlowCount.Int64
	f.ActualFlowAverageDelay.Valid = src.ActualFlowAverageDelay.Valid
	f.ActualFlowAverageDelay.Float64 = src.ActualFlowAverageDelay.Float64
	f.FormerStockCount.Valid = src.FormerStockCount.Valid
	f.FormerStockCount.Int64 = src.FormerStockCount.Int64
	f.FormerStockAverageDelay.Valid = src.FormerStockAverageDelay.Valid
	f.FormerStockAverageDelay.Float64 = src.FormerStockAverageDelay.Float64
	f.FormerFlowCount.Valid = src.FormerFlowCount.Valid
	f.FormerFlowCount.Int64 = src.FormerFlowCount.Int64
	f.FormerFlowAverageDelay.Valid = src.FormerFlowAverageDelay.Valid
	f.FormerFlowAverageDelay.Float64 = src.FormerFlowAverageDelay.Float64
}

// Get fetches from database flow and stock count and average delay
func (f *FlowStockDelays) Get(days int64, db *sql.DB) error {
	if !needUpdate(flowStockDelaysUpdate, paymentDemandsUpdate, paymentUpdate,
		everyDayUpdate) {
		f.copy(&fsd)
		return nil
	}

	query := fmt.Sprintf(`SELECT actual_stock.c,actual_stock.avg,actual_flow.c,
	actual_flow.avg, former_stock.c,former_stock.avg,former_flow.c,former_flow.avg FROM 
	(SELECT count(1) c,avg(CURRENT_DATE-receipt_date) 
	FROM payment_demands WHERE excluded=FALSE AND processed_date ISNULL) actual_stock,
	(SELECT count(1) c,avg(date-receipt_date)
	FROM payment WHERE date>CURRENT_DATE-%d) actual_flow,
	(SELECT count(1) c,avg(CURRENT_DATE-receipt_date) 
	FROM payment_demands WHERE excluded=FALSE AND 
		(processed_date ISNULL OR processed_date>= CURRENT_DATE-7)) former_stock,
	(SELECT count(1) c,avg(date-receipt_date) 
	FROM payment WHERE date>CURRENT_DATE-%d-7 AND date<=CURRENT_DATE-7) former_flow;`, days, days)
	if err := db.QueryRow(query).Scan(&f.ActualStockCount, &f.ActualStockAverageDelay, &f.ActualFlowCount,
		&f.ActualFlowAverageDelay, &f.FormerStockCount, &f.FormerStockAverageDelay,
		&f.FormerFlowCount, &f.FormerFlowAverageDelay); err != nil {
		return fmt.Errorf("select %v ", err)
	}
	var mutex = &sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()
	fsd.copy(f)
	update(csfWeekTrendUpdate)
	return nil
}
