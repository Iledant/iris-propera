package models

import "database/sql"

// MonthFinancialCommitment embeddes a row for financial commitment request.
type MonthFinancialCommitment struct {
	Month int64 `json:"month"`
	Value int64 `json:"value"`
}

// MonthFinancialCommitments embeddes an array of MonthFinancialCommit for json export.
type MonthFinancialCommitments struct {
	MonthFinancialCommitments []MonthFinancialCommitment `json:"FinancialCommitmentsPerMonth"`
}

// GetAll fetches financial commitments per month of a given year.
func (m *MonthFinancialCommitments) GetAll(year int, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT extract(month from date) AS month, sum(value) AS value
	FROM financial_commitment WHERE extract(year FROM date) = $1 GROUP BY 1 ORDER BY 1`, year)
	if err != nil {
		return err
	}
	var r MonthFinancialCommitment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Month, &r.Value); err != nil {
			return err
		}
		m.MonthFinancialCommitments = append(m.MonthFinancialCommitments, r)
	}
	err = rows.Err()
	return err
}
