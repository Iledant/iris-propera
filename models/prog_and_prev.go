package models

import "database/sql"

// ProgrammingAndPrevision is used to decode one line of dedicated query.
type ProgrammingAndPrevision struct {
	Name         string    `json:"name"`
	Number       string    `json:"number"`
	Programmings NullInt64 `json:"programmings"`
	Prevision    NullInt64 `json:"prevision"`
}

// ProgrammingAndPrevisions embeddes an array of ProgrammingAndPrevision for json export.
type ProgrammingAndPrevisions struct {
	ProgrammingAndPrevisions []ProgrammingAndPrevision `json:"ProgrammingsPrevision"`
}

// GetAll fetches programmation and commitments prevision for the given year
func (p *ProgrammingAndPrevisions) GetAll(year int64, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT op.number, op.name, pr.value as programmings, 
	  pc.value as prevision FROM physical_op op
	LEFT OUTER JOIN
	(SELECT p.physical_op_id, SUM(value) AS value FROM programmings p, commissions c 
	WHERE p.commission_id = c.id AND extract(year FROM c.date) = $1 GROUP BY 1) pr
	ON op.id = pr.physical_op_id
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
		if err = rows.Scan(&r.Number, &r.Name, &r.Programmings, &r.Prevision); err != nil {
			return err
		}
		p.ProgrammingAndPrevisions = append(p.ProgrammingAndPrevisions, r)
	}
	err = rows.Err()
	return err
}
