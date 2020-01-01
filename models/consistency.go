package models

import (
	"database/sql"
	"fmt"
	"time"
)

// CommitmentWithoutAction model
type CommitmentWithoutAction struct {
	ID              int64     `json:"id"`
	Chapter         string    `json:"chapter"`
	Action          string    `json:"action"`
	IrisCode        string    `json:"iris_code"`
	CoriolisYear    string    `json:"coriolis_year"`
	CoriolisEgtCode string    `json:"coriolis_egt_code"`
	CoriolisEgtNum  string    `json:"coriolis_egt_num"`
	CoriolisEgtLine string    `json:"coriolis_egt_line"`
	Name            string    `json:"name"`
	BeneficiaryCode int       `json:"beneficiary_code"`
	Date            time.Time `json:"date"`
	Value           int64     `json:"value"`
	LapseDate       NullTime  `json:"lapse_date"`
	APP             bool      `json:"app"`
}

// CommitmentWithoutActions is used for json export and dedicated query
type CommitmentWithoutActions struct {
	Lines []CommitmentWithoutAction `json:"CommitmentWithoutAction"`
}

// UnlinkedPayment model
type UnlinkedPayment struct {
	ID              int64     `json:"id"`
	CoriolisYear    string    `json:"coriolis_year"`
	CoriolisEgtCode string    `json:"coriolis_egt_code"`
	CoriolisEgtNum  string    `json:"coriolis_egt_num"`
	CoriolisEgtLine string    `json:"coriolis_egt_line"`
	Date            time.Time `json:"date"`
	Number          string    `json:"number"`
	Value           int64     `json:"value"`
	CancelledValue  int64     `json:"cancelled_value"`
	BeneficiaryCode int       `json:"beneficiary_code"`
}

// UnlinkedPayments embeddes an array of UnlinkedPayment for json export
// and dedicated query
type UnlinkedPayments struct {
	Lines []UnlinkedPayment `json:"UnlinkedPayment"`
}

// Get fetches all commitments not linked to a budget action
func (c *CommitmentWithoutActions) Get(db *sql.DB) error {
	q := `SELECT id,chapter,action,iris_code,coriolis_year,coriolis_egt_code,
		coriolis_egt_num,coriolis_egt_line,name,beneficiary_code,date,value,
		lapse_date,app FROM financial_commitment WHERE action_id ISNULL`

	rows, err := db.Query(q)
	if err != nil {
		return fmt.Errorf("select %v", err)
	}
	defer rows.Close()
	var line CommitmentWithoutAction
	for rows.Next() {
		if err = rows.Scan(&line.ID, &line.Chapter, &line.Action, &line.IrisCode,
			&line.CoriolisYear, &line.CoriolisEgtCode, &line.CoriolisEgtNum,
			&line.CoriolisEgtLine, &line.Name, &line.BeneficiaryCode, &line.Date,
			&line.Value, &line.LapseDate, &line.APP); err != nil {
			return fmt.Errorf("scan %v", err)
		}
		c.Lines = append(c.Lines, line)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("rows err %v", err)
	}
	if len(c.Lines) == 0 {
		c.Lines = []CommitmentWithoutAction{}
	}
	return nil
}

// Get fetches all payments not linked to a financial commitment
func (u *UnlinkedPayments) Get(db *sql.DB) error {
	q := `SELECT id,coriolis_year,coriolis_egt_code,coriolis_egt_num,
			coriolis_egt_line,date,number,value,cancelled_value,beneficiary_code
		FROM payment WHERE financial_commitment_id ISNULL`

	rows, err := db.Query(q)
	if err != nil {
		return fmt.Errorf("select %v", err)
	}
	defer rows.Close()
	var line UnlinkedPayment
	for rows.Next() {
		if err = rows.Scan(&line.ID, &line.CoriolisYear, &line.CoriolisEgtCode,
			&line.CoriolisEgtNum, &line.CoriolisEgtLine, &line.Date, &line.Number,
			&line.Value, &line.CancelledValue, &line.BeneficiaryCode); err != nil {
			return fmt.Errorf("scan %v", err)
		}
		u.Lines = append(u.Lines, line)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("rows err %v", err)
	}
	if len(u.Lines) == 0 {
		u.Lines = []UnlinkedPayment{}
	}
	return nil
}
