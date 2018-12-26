package models

import "database/sql"

// ActionProgrammation is used to decode the dedicated query.
type ActionProgrammation struct {
	ActionCode NullString `json:"action_code"`
	ActionName NullString `json:"action_name"`
	Value      int64      `json:"value"`
}

// ActionProgrammations embeddes an array of ActionProgrammation for json export.
type ActionProgrammations struct {
	ActionProgrammations []ActionProgrammation `json:"BudgetProgrammation"`
}

// GetAll calculates programmation per budget action from database.
func (a *ActionProgrammations) GetAll(year int64, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT b.action_code, b.name AS action_name, SUM(p.value) AS value 
	FROM physical_op op
	JOIN programmings p ON p.physical_op_id = op.id 
	LEFT OUTER JOIN
	(SELECT ba.id, bp.code_contract||bp.code_function||bp.code_number||COALESCE(bp.code_subfunction,'')||ba.code as action_code, ba.name
		FROM budget_program bp, budget_action ba
		WHERE ba.program_id = bp.id) b
		ON op.budget_action_id = b.id
	WHERE p.year = $1
	GROUP BY 1,2 ORDER BY substring(b.action_code from 2), substring(b.action_code for 1)`, year)
	if err != nil {
		return err
	}
	var r ActionProgrammation
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ActionCode, &r.ActionName, &r.Value); err != nil {
			return err
		}
		a.ActionProgrammations = append(a.ActionProgrammations, r)
	}
	if len(a.ActionProgrammations) == 0 {
		a.ActionProgrammations = []ActionProgrammation{}
	}
	err = rows.Err()
	return err
}
