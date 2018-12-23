package models

import (
	"database/sql"
	"time"
)

// OpPayment model.
type OpPayment struct {
	Date        time.Time `json:"date"`
	Value       int64     `json:"value"`
	Beneficiary string    `json:"beneficiary"`
	IrisCode    string    `json:"iris_code"`
}

// OpPayments embeddes an array of OpPayment.
type OpPayments struct {
	OpPayments []OpPayment `json:"Payment"`
}

// GetOpAll fetches formatted payments attached to a physical operation from database.
func (o *OpPayments) GetOpAll(opID int64, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT p.date, (p.value - p.cancelled_value) AS value, 
	b.name AS beneficiary, f.iris_code FROM payment p 
	JOIN financial_commitment f ON p.financial_commitment_id = f.id 
	JOIN beneficiary b ON b.code = f.beneficiary_code 
	WHERE p.financial_commitment_id IN 
	(SELECT f.id FROM financial_commitment f WHERE f.physical_op_id = $1)`, opID)
	if err != nil {
		return err
	}
	var r OpPayment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Date, &r.Value, &r.Beneficiary, &r.IrisCode); err != nil {
			return err
		}
		o.OpPayments = append(o.OpPayments, r)
	}
	err = rows.Err()
	return err
}
