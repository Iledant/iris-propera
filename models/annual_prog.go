package models

import (
	"database/sql"
	"time"
)

// AnnualProgLine is used to decode one row of the annual programmation query.
type AnnualProgLine struct {
	OperationNumber   NullString  `json:"operation_number"`
	Name              NullString  `json:"name"`
	StepName          NullString  `json:"step_name"`
	CategoryName      NullString  `json:"category_name"`
	Date              time.Time   `json:"date"`
	Programmings      NullInt64   `json:"programmings"`
	TotalProgrammings NullInt64   `json:"total_programmings"`
	StateRatio        NullFloat64 `json:"state_ratio"`
	Commitment        NullInt64   `json:"commitment"`
	Pendings          NullInt64   `json:"pendings"`
}

// AnnualProgrammation embeddes an array of ArrayProgLine for json export.
type AnnualProgrammation struct {
	AnnualProgrammation []AnnualProgLine `json:"AnnualProgrammation"`
}

// GetAll fetches annual programmation of the given year from database.
func (a *AnnualProgrammation) GetAll(year int, db *sql.DB) (err error) {
	qry := `WITH dates AS (
		SELECT DISTINCT date FROM financial_commitment WHERE DATE_PART('YEAR', date) = $1
		UNION
		SELECT DISTINCT c.date FROM programmings p, commissions c
			WHERE p.commission_id = c.id AND  p.year = $1
		UNION
		SELECT DISTINCT commission_date AS date FROM pending_commitments
			WHERE DATE_PART('YEAR', commission_date) = $1)
			SELECT q.operation_number::varchar, q.name::varchar, q.step_name::varchar,
			q.category_name::varchar, q.date, q.programmings::bigint,
			q.total_programmings::bigint, q.state_ratio::double precision,
			q.commitment::bigint, q.pendings::bigint FROM 
(	SELECT op.number AS operation_number, op.name, op.step_name, op.category_name, op.date, pr.value AS programmings,
			pr.total_value AS total_programmings, pr.state_ratio, fc.value AS commitment, 
			pe.proposed_value AS pendings FROM
 (SELECT op.id, op.name, op.number, step.name AS step_name, category.name AS category_name, dates.date
			 FROM physical_op op
			 CROSS JOIN dates
			 LEFT OUTER JOIN step ON op.step_id = step.id
			 LEFT OUTER JOIN category ON op.category_id = category.id) op
		LEFT JOIN
		(SELECT p.physical_op_id, SUM(p.value) AS value, SUM(p.total_value) AS total_value, p.state_ratio, c.date 
			FROM programmings p, commissions c
			WHERE p.commission_id = c.id GROUP BY 1,4,5) pr
		ON pr.date = op.date AND pr.physical_op_id = op.id
		LEFT JOIN 
		(SELECT SUM(value) AS value, physical_op_id, financial_commitment.date, null AS total_value,
						null as state_ratio FROM financial_commitment GROUP BY 2,3) fc
		ON fc.physical_op_id = op.id AND fc.date=op.date
		LEFT JOIN 
		(SELECT SUM(proposed_value) AS proposed_value, physical_op_id, commission_date AS date, 
						null AS total_value, null as state_ratio FROM pending_commitments GROUP BY 2,3) pe
		ON pe.physical_op_id = op.id AND pe.date=op.date
		WHERE pr.value NOTNULL OR fc.value NOTNULL OR pe.proposed_value NOTNULL
	UNION ALL
	SELECT NULL as operation_number, fc.name AS name, NULL as step_name, NULL as category_name, fc.date,
				 NULL AS programmings, NULL as total_programmings, NULL as state_ratio, fc.value AS commitment,
				 NULL AS pendings
		FROM financial_commitment fc
		WHERE fc.physical_op_id ISNULL AND DATE_PART('YEAR',fc.Date)= $1
	UNION ALL
		SELECT NULL as operation_number, pe.name AS name, NULL as step_name, NULL as category_name, 
					 pe.commission_date AS date, NULL AS programmings, NULL as total_programmings, NULL as state_ratio,
					 NULL AS commitment, pe.proposed_value AS pendings
			FROM pending_commitments pe
			WHERE pe.physical_op_id ISNULL AND DATE_PART('YEAR', pe.commission_date) = $1
			ORDER BY 3, 1) q`
	rows, err := db.Query(qry, year)
	if err != nil {
		return err
	}
	var r AnnualProgLine
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.OperationNumber, &r.Name, &r.StepName, &r.CategoryName,
			&r.Date, &r.Programmings, &r.TotalProgrammings, &r.StateRatio, &r.Commitment,
			&r.Pendings); err != nil {
			return err
		}
		a.AnnualProgrammation = append(a.AnnualProgrammation, r)
	}
	err = rows.Err()
	return err
}
