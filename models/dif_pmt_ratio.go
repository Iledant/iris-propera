package models

import (
	"database/sql"
	"fmt"
)

// DifPmtRatio model
type DifPmtRatio struct {
	Year  int64   `json:"year"`
	Idx   int64   `json:"idx"`
	Ratio float64 `json:"ratio"`
}

// DifPmtRatios is used for queries and json export
type DifPmtRatios struct {
	Lines []DifPmtRatio `json:"dif_pmt_ratio"`
}

// GetAll fetches all dif pmt ratios from database of a given commitment year
func (d *DifPmtRatios) GetAll(year int64, db *sql.DB) error {
	rows, err := db.Query(`select idx,ratio FROM dif_pmt_ratio WHERE year=$1`,
		year)
	if err != nil {
		return fmt.Errorf("select %v", err)
	}
	var line DifPmtRatio
	line.Year = year
	for rows.Next() {
		if err = rows.Scan(&line.Idx, &line.Ratio); err != nil {
			return fmt.Errorf("scan %v", err)
		}
		d.Lines = append(d.Lines, line)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("rows err %v", err)
	}
	if len(d.Lines) == 0 {
		d.Lines = []DifPmtRatio{}
	}
	return nil
}

// SetYear calculates the dif pmt ratios of a commitment year and update or save
// the ratios into the database
func (d *DifPmtRatios) SetYear(year int64, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("tx begin %v", err)
	}
	if _, err = tx.Exec(`delete FROM dif_pmt_ratio WHERE year=$1`, year); err != nil {
		tx.Rollback()
		return fmt.Errorf("delete %v", err)
	}
	q := `
	with fc as (select sum(value)::bigint as val from financial_commitment 
		where extract(year from date)=$1),
	pmt as (select sum(value)::bigint as val, extract(year from date)::int-$1 idx
		from payment where financial_commitment_id in 
			(select id from financial_commitment where extract(year from date)=$1)
			and extract(year from date)>=$1
			group by 2 order by 2),
	c_pmt as (select idx, sum(val) over (order by idx) from pmt),
	ram as (select pmt.idx,fc.val-COALESCE(c_pmt.sum,0) as val
		from fc, pmt left outer join c_pmt on c_pmt.idx=pmt.idx-1)
	insert into dif_pmt_ratio (year,idx,ratio) 
		select $1,pmt.idx,pmt.val::double precision/ram.val as r
			from pmt, ram where pmt.idx=ram.idx
		returning idx,ratio`
	rows, err := tx.Query(q, year)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("insert %v", err)
	}
	var line DifPmtRatio
	line.Year = year
	for rows.Next() {
		if err = rows.Scan(&line.Idx, &line.Ratio); err != nil {
			tx.Rollback()
			return fmt.Errorf("scan %v", err)
		}
		d.Lines = append(d.Lines, line)
	}
	if err = rows.Err(); err != nil {
		tx.Rollback()
		return fmt.Errorf("rows err %v", err)
	}
	if len(d.Lines) == 0 {
		d.Lines = []DifPmtRatio{}
	}
	tx.Commit()
	return nil
}
