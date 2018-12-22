package models

import (
	"database/sql"
	"strconv"
)

// DetailedActionCommitment is used to decode a line of the dedicated query.
type DetailedActionCommitment struct {
	Chapter     NullInt64   `json:"chapter"`
	Sector      NullString  `json:"sector"`
	Subfunction NullString  `json:"subfunction"`
	Program     NullString  `json:"program"`
	Action      NullString  `json:"action"`
	ActionName  NullString  `json:"action_name"`
	Name        string      `json:"name"`
	Number      string      `json:"number"`
	Y0          NullFloat64 `json:"y0"`
	Y1          NullFloat64 `json:"y1"`
	Y2          NullFloat64 `json:"y2"`
	Y3          NullFloat64 `json:"y3"`
}

// DetailedActionCommitments embeddes an array of DetailedActionCommitment
// for json export.
type DetailedActionCommitments struct {
	DetailedActionCommitments []DetailedActionCommitment `json:"DetailedCommitmentPerBudgetAction"`
}

// GetAll fetches detailed commitments previsions per budget action
func (d *DetailedActionCommitments) GetAll(year int64, db *sql.DB) (err error) {
	sy := strconv.FormatInt(year, 10)
	rows, err := db.Query(`SELECT budget.chapter, budget.sector, budget.subfunction, budget.program, budget.action, 
	budget.action_name, op.number, op.name, pg.value AS y0, ct.y1, ct.y2, ct.y3 FROM 
	physical_op op
	LEFT OUTER JOIN (SELECT * FROM crosstab('SELECT pc.physical_op_id, pc.year, pc.value * 0.01 FROM 
		(SELECT * FROM prev_commitment WHERE year >= `+sy+` AND year <=`+sy+` +2) pc ORDER BY 1,2',
'SELECT m FROM generate_series(`+sy+`,`+sy+`+ 2) AS m') AS (physical_op_id INTEGER, y1 NUMERIC, y2 NUMERIC, y3 NUMERIC)) ct
ON ct.physical_op_id = op.id 
LEFT OUTER JOIN (SELECT physical_op_id, SUM(value) * 0.01 AS value FROM programmings WHERE year = $1 GROUP BY 1) pg ON pg.physical_op_id = op.id
LEFT OUTER JOIN 
(SELECT ba.id, bc.code AS chapter, bs.code AS sector, bp.code_function || COALESCE(bp.code_subfunction, '') AS subfunction,
			bp.code_contract || bp.code_function || bp.code_number as program,
			bp.code_contract || bp.code_function || bp.code_number || ba.code as action, ba.name AS action_name FROM 
					budget_chapter bc, budget_program bp, budget_action ba, budget_sector bs
					WHERE ba.program_id = bp.id AND bp.chapter_id = bc.id AND ba.sector_id = bs.id) AS budget
ON op.budget_action_id = budget.id
WHERE pg.value IS NOT NULL OR (ct.y1  <> 0 AND ct.y1 IS NOT NULL) OR (ct.y2 <> 0 AND ct.y2 IS NOT NULL) OR (ct.y3 <> 0 AND ct.y3 IS NOT NULL)
ORDER BY 1, 2, 3, 4, 5`, year)
	if err != nil {
		return err
	}
	var r DetailedActionCommitment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Chapter, &r.Sector, &r.Subfunction, &r.Program, &r.Action,
			&r.ActionName, &r.Name, &r.Number, &r.Y0, &r.Y1, &r.Y2, &r.Y3); err != nil {
			return err
		}
		d.DetailedActionCommitments = append(d.DetailedActionCommitments, r)
	}
	err = rows.Err()
	return err
}
