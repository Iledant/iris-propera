package models

import "database/sql"

// MonthCommitment embeddes a row for financial commitment request.
type MonthCommitment struct {
	Month int64 `json:"month"`
	Value int64 `json:"value"`
}

// MonthCommitments embeddes an array of MonthFinancialCommit for json export.
type MonthCommitments struct {
	MonthCommitments []MonthCommitment `json:"FinancialCommitmentsPerMonth"`
}

// GetAll fetches financial commitments per month of a given year.
func (m *MonthCommitments) GetAll(year int, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT extract(month FROM date), sum(value)
	FROM financial_commitment WHERE extract(year FROM date)=$1 GROUP BY 1 ORDER BY 1`, year)
	if err != nil {
		return err
	}
	var r MonthCommitment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Month, &r.Value); err != nil {
			return err
		}
		m.MonthCommitments = append(m.MonthCommitments, r)
	}
	err = rows.Err()
	if len(m.MonthCommitments) == 0 {
		m.MonthCommitments = []MonthCommitment{}
	}
	return err
}
