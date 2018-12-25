package models

import "database/sql"

// PlanLineRatio model
type PlanLineRatio struct {
	ID            int     `json:"id"`
	PlanLineID    int64   `json:"plan_line_id"`
	BeneficiaryID int64   `json:"beneficiary_id"`
	Ratio         float64 `json:"ratio"`
}

// PlanLineRatios embeddes an array of PlanLineRatio.
type PlanLineRatios struct {
	PlanLineRatios []PlanLineRatio `json:"ratios"`
}

// Save replace ratios linked to a plan line whose ID is given.
func (p *PlanLineRatios) Save(plID int64, tx *sql.Tx) (err error) {
	if _, err = tx.Exec("DELETE FROM plan_line_ratios WHERE plan_line_id=$1",
		plID); err != nil {
		return err
	}
	for _, r := range p.PlanLineRatios {
		if _, err = tx.Exec(`INSERT INTO plan_line_ratios 
		(plan_line_id, beneficiary_id, ratio) VALUES ($1,$2,$3)`,
			plID, r.BeneficiaryID, r.Ratio); err != nil {
			return err
		}
	}
	return err
}
