package models

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
)

// PaymentRatio model
type PaymentRatio struct {
	ID            int64     `json:"id" gorm:"column:id"`
	PaymentTypeID NullInt64 `json:"payment_types_id" gorm:"column:payment_types_id"`
	Ratio         float64   `json:"ratio" gorm:"column:ratio"`
	Index         int64     `json:"index" gorm:"column:index"`
}

// PaymentRatios embeddes an array of PaymentRatio for json export.
type PaymentRatios struct {
	PaymentRatios []PaymentRatio `json:"PaymentRatio"`
}

// PaymentRatioLine embeddes a line sent to set the payments ratios of a payment type.
type PaymentRatioLine struct {
	Ratio float64 `json:"ratio"`
	Index int64   `json:"index"`
}

// PaymentRatiosBatch embeddes an array of payment ratios lines
type PaymentRatiosBatch struct {
	PaymentRatios []PaymentRatioLine `json:"PaymentRatio"`
}

// YearRatio is used to scan and encode an year ratio
type YearRatio struct {
	Index int64   `json:"index"`
	Ratio float64 `json:"ratio"`
}

// YearRatios embeddes an array of YearRatio for json export.
type YearRatios struct {
	YearRatios []YearRatio `json:"Ratios"`
}

// GetAll fetches all payment ratios from database.
func (p *PaymentRatios) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id,payment_types_id,ratio,index FROM payment_ratios`)
	if err != nil {
		return err
	}
	var r PaymentRatio
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.PaymentTypeID, &r.Ratio, &r.Index); err != nil {
			return err
		}
		p.PaymentRatios = append(p.PaymentRatios, r)
	}
	err = rows.Err()
	return err
}

// GetPaymentTypeAll fetches all payment ratios linked to a payment type from database.
func (p *PaymentRatios) GetPaymentTypeAll(paymentTypeID int64, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id,payment_types_id,ratio,index 
	FROM payment_ratios WHERE payment_types_id = $1`, paymentTypeID)
	if err != nil {
		return err
	}
	var r PaymentRatio
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.PaymentTypeID, &r.Ratio, &r.Index); err != nil {
			return err
		}
		p.PaymentRatios = append(p.PaymentRatios, r)
	}
	err = rows.Err()
	return err
}

// DeleteRatios removes a payment ratios linked to a payment type from database.
func (p *PaymentType) DeleteRatios(db *sql.DB) (err error) {
	res, err := db.Exec("DELETE FROM payment_ratios WHERE payment_types_id = $1", p.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("Ratios de paiement introuvables")
	}
	return nil
}

// Save a batch of payment ratios to the database
func (p *PaymentRatiosBatch) Save(paymentTypeID int64, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM payment_ratios WHERE payment_types_id = $1`, paymentTypeID); err != nil {
		tx.Rollback()
		return err
	}
	sPtID := toSQL(paymentTypeID)
	var values []string
	var value string
	for _, pr := range p.PaymentRatios {
		value = "(" + sPtID + ", " + toSQL(pr.Ratio) + "," + toSQL(pr.Index) + ")"
		values = append(values, value)
	}
	qry := "INSERT into payment_ratios (payment_types_id, ratio, index) VALUES " + strings.Join(values, ",")
	if _, err = tx.Exec(qry); err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}

// GetAll fetches the ratios of payment transformation of commitment of a given year.
func (y *YearRatios) GetAll(year int64, db *sql.DB) (err error) {
	sy := strconv.FormatInt(year, 10)
	rows, err := db.Query(`WITH yc AS (SELECT f.id FROM financial_commitment f WHERE f.coriolis_year = $1),
	total AS (SELECT sum(f.value) as total FROM financial_commitment f WHERE f.id IN (SELECT id FROM yc))
	SELECT extract(YEAR from p.date) - $2 AS index, SUM(p.value/total.total)::double precision AS ratio
	FROM payment p, total
	WHERE p.financial_commitment_id IN (SELECT id FROM yc) GROUP BY index ORDER BY index`, sy, year)
	if err != nil {
		return err
	}
	var r YearRatio
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Index, &r.Ratio); err != nil {
			return err
		}
		y.YearRatios = append(y.YearRatios, r)
	}
	err = rows.Err()
	return err
}
