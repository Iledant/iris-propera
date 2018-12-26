package models

import (
	"database/sql"
	"strconv"
)

// ActionCommitment is used to decode a line of the dedicated query
type ActionCommitment struct {
	Chapter     int64       `json:"chapter"`
	Sector      string      `json:"sector"`
	Subfunction string      `json:"subfunction"`
	Program     string      `json:"program"`
	Action      string      `json:"action"`
	ActionName  string      `json:"action_name"`
	Y0          NullFloat64 `json:"y0"`
	Y1          NullFloat64 `json:"y1"`
	Y2          NullFloat64 `json:"y2"`
	Y3          NullFloat64 `json:"y3"`
}

// ActionCommitments embeddes an array of ActionCommitments for json export.
type ActionCommitments struct {
	ActionCommitments []ActionCommitment `json:"CommitmentPerBudgetAction"`
}

// GetAll fetches commitments per budget action for the given year from the database.
func (a *ActionCommitments) GetAll(year int64, db *sql.DB) (err error) {
	sy := strconv.FormatInt(year, 10)
	rows, err := db.Query(`WITH budget as (SELECT ba.id, bc.code AS chapter, bs.code AS sector, 
		bp.code_function || COALESCE(bp.code_subfunction, '') AS subfunction,
		bp.code_contract || bp.code_function || bp.code_number as program,
		bp.code_contract || bp.code_function || bp.code_number || ba.code as action, 
		ba.name AS action_name 
	FROM budget_chapter bc, budget_program bp, budget_action ba, budget_sector bs
	WHERE ba.program_id = bp.id AND bp.chapter_id = bc.id AND ba.sector_id = bs.id) 
	
SELECT budget.chapter, budget.sector, budget.subfunction, budget.program, budget.action, budget.action_name,
SUM(y0) AS y0, SUM(tot.y1) AS y1, SUM(tot.y2) AS y2, SUM(tot.y3) AS y3
FROM 
(SELECT *, NULL as y0 
FROM crosstab('SELECT op.budget_action_id, pc.year, SUM(pc.value) * 0.01 
		FROM
			(SELECT * FROM prev_commitment WHERE year >= `+sy+` 
		 AND year <= `+sy+` + 2) pc, physical_op op 
		WHERE pc.physical_op_id = op.id GROUP BY 1,2 ORDER BY 1,2', 
		'SELECT m FROM generate_series(`+sy+`, `+sy+` + 2) AS m')
AS (budget_action_id INTEGER, y1 NUMERIC, y2 NUMERIC, y3 NUMERIC)
UNION ALL 
SELECT op.budget_action_id, NULL as y1, NULL as y2, NULL as y3, SUM(pg.value) * 0.01 AS y0
FROM programmings pg, physical_op op
WHERE pg.year = $1 - 1 AND pg.physical_op_id = op.id GROUP BY 1) tot, budget
WHERE tot.budget_action_id = budget.id
GROUP BY 1,2,3,4,5,6 ORDER BY 1,2,3,4,5`, year)
	if err != nil {
		return err
	}
	var r ActionCommitment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Chapter, &r.Sector, &r.Subfunction, &r.Program, &r.Action,
			&r.ActionName, &r.Y0, &r.Y1, &r.Y2, &r.Y3); err != nil {
			return err
		}
		a.ActionCommitments = append(a.ActionCommitments, r)
	}
	if len(a.ActionCommitments) == 0 {
		a.ActionCommitments = []ActionCommitment{}
	}
	err = rows.Err()
	return err
}
