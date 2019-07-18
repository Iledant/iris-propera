package models

import "database/sql"

// FullCodeBudgetAction is used to decode a row of the query that fetches
// a budget action with an explicit string code
type FullCodeBudgetAction struct {
	ID        int64  `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	ProgramID int64  `json:"program_id"`
	SectorID  int64  `json:"sector_id"`
	FullCode  string `json:"full_code"`
}

// FullCodeBudgetActions embeddes an array of FullCodeBudgetAction
type FullCodeBudgetActions struct {
	FullCodeBudgetActions []FullCodeBudgetAction `json:"BudgetAction"`
}

// GetAll fetches all budget actions with full code name from database.
func (f *FullCodeBudgetActions) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT ba.id, ba.code, ba.name, ba.program_id, ba.sector_id, 
	bp.code_contract||bp.code_function||bp.code_number||ba.code AS full_code 
	FROM budget_action ba, budget_program bp WHERE ba.program_id = bp.id`)
	if err != nil {
		return err
	}
	var r FullCodeBudgetAction
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Code, &r.Name, &r.ProgramID, &r.SectorID, &r.FullCode); err != nil {
			return err
		}
		f.FullCodeBudgetActions = append(f.FullCodeBudgetActions, r)
	}
	err = rows.Err()
	return err
}
