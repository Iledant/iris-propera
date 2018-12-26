package models

import "database/sql"

// CurrentYearPrevPayment is used to decode one row of dedicated query.
type CurrentYearPrevPayment struct {
	Chapter     NullInt64   `json:"chapter"`
	Sector      NullString  `json:"sector"`
	SubFunction NullString  `json:"subfunction"`
	Program     NullString  `json:"program"`
	Action      NullString  `json:"action"`
	ActionName  NullString  `json:"action_name"`
	Prevision   NullFloat64 `json:"prevision"`
	Payment     NullFloat64 `json:"payment"`
}

// CurrentYearPrevPayments embeddes an array of CurrentYearPrevPayment.
type CurrentYearPrevPayments struct {
	CurrentYearPrevPayments []CurrentYearPrevPayment `json:"StatisticalCurrentYearPaymentPerAction"`
}

// GetAll calculates the CurrentYearPrevPayments of the given year
// using payment types whose ID is given.
func (c *CurrentYearPrevPayments) GetAll(year int64, ptID int64, db *sql.DB) (err error) {
	rows, err := db.Query(`WITH pr AS (SELECT * FROM payment_ratios WHERE payment_types_id=$1),
		fc_sum AS (SELECT physical_op_id, EXTRACT(year FROM date)::integer AS year, 
			SUM(value) AS value FROM financial_commitment
			WHERE EXTRACT(year FROM date)<$2 GROUP BY 1,2),
		fc AS (SELECT fc_sum.physical_op_id, fc_sum.year+pr.index AS year, 
			fc_sum.value*pr.ratio AS value FROM fc_sum, pr WHERE fc_sum.year+pr.index=$2),
		fc_filtered AS (SELECT * FROM fc WHERE fc.physical_op_id NOTNULL),
		pg_year AS (SELECT physical_op_id, year, SUM(value) AS value FROM programmings 
			WHERE year=$2 GROUP BY 1,2),
		pg AS (SELECT pg_year.physical_op_id, pg_year.year+pr.index AS year, 
			pg_year.value*pr.ratio AS value FROM pg_year, pr WHERE pg_year.year+pr.index=$2),
		unlinked_fc_sum AS (SELECT action_id, EXTRACT(year FROM date)::integer AS year, 
			SUM(value) AS value FROM financial_commitment 
			WHERE EXTRACT(year FROM date)<$2 AND physical_op_id ISNULL GROUP BY 1,2),
		unlinked_fc AS (SELECT ufs.action_id, ufs.year+pr.index AS year, 
			ufs.value*pr.ratio AS value FROM unlinked_fc_sum ufs, pr 
			WHERE ufs.year+pr.index=$2)
	SELECT b.chapter, b.sector, b.subfunction, b.program, b.action, b.action_name, 
		SUM(prev)*0.01 AS prev, SUM(yp.value)*0.01 AS payment FROM
	(
		(SELECT op.budget_action_id AS action_id, SUM(q2.prev) AS prev FROM
			(SELECT physical_op_id, SUM(q1.value) AS prev FROM 
				(SELECT * FROM fc_filtered UNION ALL SELECT * FROM pg) q1 
			WHERE year=$2 GROUP BY 1) q2
		LEFT JOIN physical_op op ON op.id=q2.physical_op_id
		GROUP BY 1)
		UNION ALL
		SELECT action_id, SUM(prev) AS prev FROM
			(SELECT action_id, SUM(value) AS prev FROM unlinked_fc WHERE year=$2 GROUP BY 1) q4
		GROUP BY 1
	) cq_union
	FULL OUTER JOIN   
	(SELECT f.action_id, SUM(p.value)::bigint as value
		FROM financial_commitment f, payment p 
		WHERE p.financial_commitment_id=f.id AND EXTRACT(year FROM p.date)=$2 GROUP BY 1) yp
	ON cq_union.action_id = yp.action_id
	LEFT JOIN (SELECT ba.id, bc.code AS chapter, bs.code AS sector, 
		bp.code_function || COALESCE(bp.code_subfunction, '') AS subfunction,
		bp.code_contract || bp.code_function || bp.code_number as program,
		bp.code_contract || bp.code_function || bp.code_number || ba.code as action,
		ba.name AS action_name FROM 
			budget_chapter bc, budget_program bp, budget_action ba, budget_sector bs
		WHERE ba.program_id=bp.id AND bp.chapter_id=bc.id AND ba.sector_id=bs.id) b
	ON cq_union.action_id = b.id
	GROUP BY 1,2,3,4,5,6
	ORDER BY 1,2,3,4,5,6`, ptID, year)
	if err != nil {
		return err
	}
	var r CurrentYearPrevPayment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Chapter, &r.Sector, &r.SubFunction, &r.Program,
			&r.Action, &r.ActionName, &r.Prevision, &r.Payment); err != nil {
			return err
		}
		c.CurrentYearPrevPayments = append(c.CurrentYearPrevPayments, r)
	}
	err = rows.Err()
	return err
}
