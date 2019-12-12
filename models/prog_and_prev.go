package models

import "database/sql"

// ProgrammingAndPrevision is used to decode one line of dedicated query.
type ProgrammingAndPrevision struct {
	Name              string     `json:"name"`
	Number            string     `json:"number"`
	Programmings      NullInt64  `json:"programmings"`
	ProgCommission    NullString `json:"programmings_commission"`
	ProgDate          NullTime   `json:"programmings_date"`
	PreProgrammings   NullInt64  `json:"pre_programmings"`
	PreProgCommission NullString `json:"pre_programmings_commission"`
	PreProgDate       NullTime   `json:"pre_programmings_date"`
	Prevision         NullInt64  `json:"prevision"`
	CategoryName      NullString `json:"category_name"`
	ChapterCode       NullInt64  `json:"chapter_code"`
}

// ProgrammingAndPrevisions embeddes an array of ProgrammingAndPrevision for json export.
type ProgrammingAndPrevisions struct {
	ProgrammingAndPrevisions []ProgrammingAndPrevision `json:"ProgrammingsPrevision"`
}

// GetAll fetches programmation and commitments prevision for the given year
func (p *ProgrammingAndPrevisions) GetAll(year int64, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT op.number, op.name, pr.value AS programmings, 
	ppr.value AS pre_programmings, pc.value AS prevision, cat.name AS category_name, 
	bud.code AS chapter_code FROM physical_op op
	LEFT OUTER JOIN category cat ON op.category_id = cat.id
	LEFT OUTER JOIN
	(SELECT ba.id, bc.code FROM budget_action ba
		JOIN budget_program bp ON ba.program_id = bp.id
		JOIN budget_chapter bc ON bp.chapter_id = bc.id ) bud 
	ON op.budget_action_id = bud.id 
	LEFT OUTER JOIN
	(SELECT p.physical_op_id, SUM(value) AS value FROM programmings p, commissions c 
	WHERE p.commission_id = c.id AND extract(year FROM c.date) = $1 GROUP BY 1) pr
	ON op.id = pr.physical_op_id
	LEFT OUTER JOIN
	(SELECT p.physical_op_id, SUM(value) AS value FROM pre_programmings p, commissions c 
	WHERE p.commission_id = c.id AND extract(year FROM c.date) = $1 GROUP BY 1) ppr
	ON op.id = ppr.physical_op_id
	LEFT OUTER JOIN
	(SELECT f.physical_op_id, value FROM prev_commitment f WHERE year = $1) pc
	ON op.id = pc.physical_op_id
	WHERE pr.value NOTNULL or (pc.value NOTNULL AND pc.value <> 0)
	ORDER BY op.number`, year)
	if err != nil {
		return err
	}
	var r ProgrammingAndPrevision
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Number, &r.Name, &r.Programmings, &r.PreProgrammings,
			&r.Prevision, &r.CategoryName, &r.ChapterCode); err != nil {
			return err
		}
		p.ProgrammingAndPrevisions = append(p.ProgrammingAndPrevisions, r)
	}
	err = rows.Err()
	if len(p.ProgrammingAndPrevisions) == 0 {
		p.ProgrammingAndPrevisions = []ProgrammingAndPrevision{}
	}
	return err
}
