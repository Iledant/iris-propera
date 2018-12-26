package models

import "database/sql"

// YearBudgetCredit is used to decode the query that calculates
//  budget credits of a year
type YearBudgetCredit struct {
	Month              int   `json:"month"`
	ChapterID          int   `json:"chapter_id"`
	PrimaryCommitment  int64 `json:"primary_commitment"`
	FrozenCommitment   int64 `json:"frozen_commitment"`
	ReservedCommitment int64 `json:"reserved_commitment"`
}

// YearBudgetCredits embeddes an array of YearBudgetCredit for json export.
type YearBudgetCredits struct {
	YearBudgetCredits []YearBudgetCredit `json:"BudgetCredits"`
}

// GetAll fetches all budget credits of a given year from database
func (b *YearBudgetCredits) GetAll(year int, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT extract(month FROM commission_date)::integer as month, 
	chapter_id, primary_commitment, frozen_commitment, reserved_commitment
	FROM budget_credits 
	WHERE (extract(day FROM commission_date), extract(month FROM commission_date)) in
	(SELECT max(extract(day FROM commission_date)), extract(month FROM commission_date)
		FROM budget_credits 
		WHERE extract(year FROM commission_date)=$1 GROUP BY 2 ORDER BY 2,1)
		 AND extract(year FROM commission_date)=$1
	ORDER BY 1,2`, year)
	if err != nil {
		return err
	}
	var r YearBudgetCredit
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Month, &r.ChapterID, &r.PrimaryCommitment,
			&r.FrozenCommitment, &r.ReservedCommitment); err != nil {
			return err
		}
		b.YearBudgetCredits = append(b.YearBudgetCredits, r)
	}
	err = rows.Err()
	if len(b.YearBudgetCredits) == 0 {
		b.YearBudgetCredits = []YearBudgetCredit{}
	}
	return err
}
