package models

import (
	"database/sql"
	"fmt"
	"time"
)

// BeneficiaryCmt model
type BeneficiaryCmt struct {
	ID        int64      `json:"id"`
	Date      time.Time  `json:"date"`
	IrisCode  string     `json:"irisCode"`
	Name      string     `json:"name"`
	Value     int64      `json:"value"`
	LapseDate NullTime   `json:"lapseDate"`
	Available int64      `json:"available"`
	OpID      NullInt64  `json:"opId"`
	OpNumber  NullString `json:"opNumber"`
	OpName    NullString `json:"opName"`
}

// BeneficiaryCmts embeddes an array of BeneficiaryCmt for json export and database
// fetching
type BeneficiaryCmts struct {
	Lines []BeneficiaryCmt `json:"BeneficiaryCommitment"`
}

// GetAll fetches all commitments linked to a beneficiary whose ID is given
func (b *BeneficiaryCmts) GetAll(ID int64, db *sql.DB) error {
	rows, err := db.Query(`SELECT f.id, f.date, f.iris_code, f.name AS name, f.value, 
	f.lapse_date, f.value - COALESCE(SUM(p.value - p.cancelled_value),0) AS available,
	op.id,op.number,op.name
	FROM financial_commitment f
	JOIN beneficiary b ON b.code = f.beneficiary_code
	LEFT JOIN payment p ON p.financial_commitment_id = f.id
	LEFT OUTER JOIN physical_op op ON f.physical_op_id=op.id
	WHERE b.id = $1 GROUP BY 1,2,3,5,6,8,9,10 ORDER BY 2`, ID)
	if err != nil {
		return fmt.Errorf("select %v", err)
	}
	var r BeneficiaryCmt
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Date, &r.IrisCode, &r.Name, &r.Value,
			&r.LapseDate, &r.Available, &r.OpID, &r.OpNumber, &r.OpName); err != nil {
			return fmt.Errorf("scan %v", err)
		}
		b.Lines = append(b.Lines, r)
	}
	err = rows.Err()
	if len(b.Lines) == 0 {
		b.Lines = []BeneficiaryCmt{}
	}
	return err
}
