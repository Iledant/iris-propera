package models

import (
	"database/sql"
	"fmt"
	"time"
)

// PossibleLinkedCmt model
type PossibleLinkedCmt struct {
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

// PossibleLinkedCmts embeddes an array of PossibleLinkedCmt for json export
// and dedicated query
type PossibleLinkedCmts struct {
	Lines []PossibleLinkedCmt `json:"Commitment"`
}

// Get fetches commitments from database that have common fields with the payment
// whose ID is given
func (p *PossibleLinkedCmts) Get(pmtID int64, db *sql.DB) error {
	rows, err := db.Query(`SELECT f.id,f.chapter,f.action,f.iris_code,f.coriolis_year,
	f.coriolis_egt_code,f.coriolis_egt_num,f.coriolis_egt_line,f.name,
	f.beneficiary_code,f.date,f.value,f.lapse_date,f.app
	FROM financial_commitment f
  JOIN (SELECT coriolis_year,coriolis_egt_code,coriolis_egt_num
    FROM payment WHERE id=$1) q
  ON f.coriolis_year=q.coriolis_year AND f.coriolis_egt_code=q.coriolis_egt_code
    AND levenshtein(f.coriolis_egt_num,q.coriolis_egt_num)<2;`, pmtID)
	if err != nil {
		return fmt.Errorf("select %v", err)
	}
	var line PossibleLinkedCmt
	for rows.Next() {
		if err = rows.Scan(&line.ID, &line.Chapter, &line.Action, &line.IrisCode,
			&line.CoriolisYear, &line.CoriolisEgtCode, &line.CoriolisEgtNum,
			&line.CoriolisEgtLine, &line.Name, &line.BeneficiaryCode, &line.Date,
			&line.Value, &line.LapseDate, &line.APP); err != nil {
			return fmt.Errorf("scan %v", err)
		}
		p.Lines = append(p.Lines, line)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("rows err %v", err)
	}
	if len(p.Lines) == 0 {
		p.Lines = []PossibleLinkedCmt{}
	}
	return nil
}
