package models

import (
	"database/sql"
	"strconv"
)

// ScenarioActionPayment is used to decode a line of the dedicated line.
type ScenarioActionPayment struct {
	Chapter     NullString  `json:"chapter"`
	Sector      NullString  `json:"sector"`
	Subfunction NullString  `json:"subfunction"`
	Program     NullString  `json:"program"`
	Action      NullString  `json:"action"`
	ActionName  NullString  `json:"action_name"`
	Y1          NullFloat64 `json:"y1"`
	Y2          NullFloat64 `json:"y2"`
	Y3          NullFloat64 `json:"y3"`
}

// ScenarioActionPayments embeddes an array of ScenarioPayment.
type ScenarioActionPayments struct {
	ScenarioActionPayments []ScenarioActionPayment `json:"ScenarioPaymentPerBudgetAction"`
}

// ScenarioStatActionPayments embeddes an array of ScenarioPayment.
type ScenarioStatActionPayments struct {
	ScenarioStatActionPayments []ScenarioActionPayment `json:"ScenarioStatisticalPaymentPerBudgetAction"`
}

// GetAll populates ScenarioActionPayments calculating the payment previsions
// of the scenario whose ID is given since firstYear.
func (s *ScenarioActionPayments) GetAll(firstYear int64, sID int64, ptID int64, db *sql.DB) (err error) {
	sy := strconv.FormatInt(firstYear, 10)
	ssID := strconv.FormatInt(sID, 10)
	sptID := strconv.FormatInt(ptID, 10)
	rows, err := db.Query(`SELECT b.chapter, b.sector, b.subfunction, b.program, 
	b.action, b.action_name, SUM(y1)*0.01 AS y1, SUM(y2)*0.01 AS y2, SUM(y3)*0.01 AS y3 FROM
(
(SELECT op.budget_action_id AS action_id, SUM(ct.y1) AS y1, SUM(ct.y2) AS y2, SUM(ct.y3) AS y3 FROM
crosstab(
	'WITH pr AS (SELECT*FROM payment_ratios WHERE payment_types_id=` + sptID + `),
			pp AS (SELECT physical_op_id, year, value FROM prev_payment
					WHERE value NOTNULL AND value<>0 AND year>=` + sy + ` AND year<=` + sy + `+2),
			pp_idx AS (SELECT physical_op_id, year FROM pp),
			fc_sum AS (SELECT physical_op_id, EXTRACT(year FROM date)::integer AS year, 
					SUM(value) AS value FROM financial_commitment 
					WHERE EXTRACT(year FROM date)<` + sy + `-1 GROUP BY 1,2),
			fc AS (SELECT fc_sum.physical_op_id, fc_sum.year+pr.index AS year, 
					fc_sum.value*pr.ratio AS value FROM fc_sum, pr
					WHERE fc_sum.year+pr.index>=` + sy + ` AND fc_sum.year+pr.index<=` + sy + `+2),
			fc_filtered AS (SELECT*FROM fc WHERE fc.physical_op_id NOTNULL 
					AND (fc.physical_op_id, fc.year) NOT IN (SELECT*FROM pp_idx)),
			pg_year AS (SELECT physical_op_id, year, SUM(value) AS value FROM programmings
					WHERE year=` + sy + `-1 GROUP BY 1,2),
			pg AS (SELECT pg_year.physical_op_id, pg_year.year+pr.index AS year, 
					pg_year.value*pr.ratio AS value FROM pg_year, pr
					WHERE pg_year.year+pr.index>=` + sy + ` AND pg_year.year+pr.index<=` + sy + `+2),
			pg_filtered AS (SELECT*FROM pg WHERE (pg.physical_op_id, pg.year) NOT IN (SELECT*FROM pp_idx)),
			sc AS (SELECT s.physical_op_id, p.year+s.offset AS year, p.value 
					FROM scenario_offset s, prev_commitment p
					WHERE s.scenario_id=` + ssID + ` AND s.physical_op_id = p.physical_op_id),
			pc AS (SELECT sc.physical_op_id, sc.year+pr.index AS year, sc.value*pr.ratio AS value
					FROM sc, pr WHERE sc.year+pr.index>=` + sy + ` AND sc.year+pr.index<=` + sy + `+2),
			pc_filtered AS (SELECT*FROM pc WHERE (pc.physical_op_id, pc.year) NOT IN (SELECT*FROM pp_idx))
	SELECT*FROM
	(SELECT*FROM pp
	UNION ALL
	SELECT physical_op_id, year, SUM(value) AS value FROM 
			(SELECT*FROM fc_filtered UNION ALL SELECT*FROM pg_filtered 
				UNION ALL SELECT*FROM pc_filtered) q1
			GROUP BY 1,2) q2 ORDER BY 1,2',
	'SELECT m FROM generate_series(` + sy + `, ` + sy + `+2) AS m') 
AS ct(physical_op_id integer, y1 numeric, y2 numeric, y3 numeric)
LEFT JOIN physical_op op ON op.id=ct.physical_op_id
GROUP BY 1
)
UNION ALL
(SELECT action_id, SUM(y1) AS y1, SUM(y2) AS y2, SUM(y3) AS y3 FROM
crosstab(
	'WITH pr AS (SELECT*FROM payment_ratios WHERE payment_types_id=` + sptID + `),
			unlinked_fc_sum AS (SELECT action_id, EXTRACT(year FROM date)::integer AS year, 
					SUM(value) AS value
					FROM financial_commitment
					WHERE EXTRACT(year FROM date)<` + sy + `-1 AND physical_op_id ISNULL GROUP BY 1,2),
			unlinked_fc AS (SELECT ufs.action_id, ufs.year+pr.index AS year, ufs.value*pr.ratio AS value
					FROM unlinked_fc_sum ufs, pr 
					WHERE ufs.year+pr.index>=` + sy + ` AND ufs.year+pr.index<=` + sy + `+2)
	SELECT action_id, year, SUM(value) FROM unlinked_fc GROUP BY 1,2 ORDER BY 1,2',
	'SELECT m FROM generate_series(` + sy + `, ` + sy + `+2) AS m')
AS (action_id integer, y1 numeric, y2 numeric, y3 numeric)
GROUP BY 1)
) cq_union
LEFT JOIN (SELECT ba.id, bc.code AS chapter, bs.code AS sector, 
bp.code_function || COALESCE(bp.code_subfunction, '') AS subfunction,
bp.code_contract || bp.code_function || bp.code_number as program,
bp.code_contract || bp.code_function || bp.code_number || ba.code as action,
ba.name AS action_name
FROM budget_chapter bc, budget_program bp, budget_action ba, budget_sector bs
WHERE ba.program_id = bp.id AND bp.chapter_id = bc.id AND ba.sector_id = bs.id) b
ON cq_union.action_id = b.id
GROUP BY 1,2,3,4,5,6
ORDER BY 1,2,3,4,5,6`)
	if err != nil {
		return err
	}
	var r ScenarioActionPayment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Chapter, &r.Sector, &r.Subfunction, &r.Program, &r.Action,
			&r.ActionName, &r.Y1, &r.Y2, &r.Y3); err != nil {
			return err
		}
		s.ScenarioActionPayments = append(s.ScenarioActionPayments, r)
	}
	err = rows.Err()
	return err
}

// GetAll populates ScenarioStatActionPayments calculating the payment previsions
// of the scenario whose ID is given since firstYear using a pure statistical approach.
func (s *ScenarioStatActionPayments) GetAll(firstYear int64, sID int64, ptID int64, db *sql.DB) (err error) {
	sy := strconv.FormatInt(firstYear, 10)
	ssID := strconv.FormatInt(sID, 10)
	sptID := strconv.FormatInt(ptID, 10)
	if err != nil {
		return err
	}
	rows, err := db.Query(`SELECT b.chapter, b.sector, b.subfunction, b.program, 
	b.action, b.action_name, SUM(y1)*0.01 AS y1, SUM(y2)*0.01 AS y2,
	SUM(y3)*0.01 AS y3 FROM
(
(SELECT op.budget_action_id AS action_id, SUM(ct.y1) AS y1, SUM(ct.y2) AS y2, 
	SUM(ct.y3) AS y3 FROM
crosstab(
	'WITH pr AS (SELECT*FROM payment_ratios WHERE payment_types_id=` + sptID + `),
		fc_sum AS (SELECT physical_op_id, EXTRACT(year FROM date)::integer AS year, 
			SUM(value) AS value FROM financial_commitment
			WHERE EXTRACT(year FROM date)<` + sy + `-1 GROUP BY 1,2),
		fc AS (SELECT fc_sum.physical_op_id, fc_sum.year+pr.index AS year, 
			fc_sum.value*pr.ratio AS value FROM fc_sum, pr
			WHERE fc_sum.year+pr.index>=` + sy + ` AND fc_sum.year+pr.index<=` + sy + `+2),
	fc_filtered AS (SELECT*FROM fc WHERE fc.physical_op_id IS NOT NULL),
		pg_year AS (SELECT physical_op_id, year, SUM(value) AS value FROM programmings
			WHERE year = ` + sy + ` - 1 GROUP BY 1,2),
		pg AS (SELECT pg_year.physical_op_id, pg_year.year+pr.index AS year, 
			pg_year.value*pr.ratio AS value  FROM pg_year, pr 
			WHERE pg_year.year+pr.index>=` + sy + ` AND pg_year.year+pr.index<=` + sy + `+2),
		sc AS (SELECT s.physical_op_id, p.year+s.offset AS year, p.value 
			FROM scenario_offset s, prev_commitment p 
			WHERE s.scenario_id = ` + ssID + ` AND s.physical_op_id = p.physical_op_id),
		pc AS (SELECT sc.physical_op_id, sc.year+pr.index AS year, sc.value*pr.ratio AS value 
			FROM sc, pr WHERE sc.year+pr.index>=` + sy + ` AND sc.year+pr.index<=` + sy + `+2)
	SELECT physical_op_id, year, SUM(value) AS value FROM 
			(SELECT*FROM fc_filtered UNION ALL SELECT*FROM pg UNION ALL SELECT*FROM pc) q1 
			GROUP BY 1,2 ORDER BY 1,2',
	'SELECT m FROM generate_series(` + sy + `, ` + sy + `+2) AS m') 
	AS ct(physical_op_id integer, y1 numeric, y2 numeric, y3 numeric)
LEFT JOIN physical_op op ON op.id = ct.physical_op_id
GROUP BY 1
)
UNION ALL
(SELECT action_id, SUM(y1) AS y1, SUM(y2) AS y2, SUM(y3) AS y3 FROM
crosstab(
	'WITH pr AS (SELECT*FROM payment_ratios WHERE payment_types_id = ` + sptID + `),
		unlinked_fc_sum AS (SELECT action_id, EXTRACT(year FROM date)::integer AS year, 
			SUM(value) AS value FROM financial_commitment 
			WHERE EXTRACT(year FROM date)<` + sy + `-1 AND physical_op_id ISNULL GROUP BY 1,2),
		unlinked_fc AS (SELECT ufs.action_id, ufs.year+pr.index AS year,
				ufs.value*pr.ratio AS value FROM unlinked_fc_sum ufs, pr 
			WHERE ufs.year+pr.index>=` + sy + ` AND ufs.year+pr.index<=` + sy + `+2)
	SELECT action_id, year, SUM(value) FROM unlinked_fc GROUP BY 1,2 ORDER BY 1,2',
	'SELECT m FROM generate_series(` + sy + `, ` + sy + `+2) AS m')
AS (action_id integer, y1 numeric, y2 numeric, y3 numeric)
GROUP BY 1)
) cq_union
LEFT JOIN (SELECT ba.id, bc.code AS chapter, bs.code AS sector, 
bp.code_function || COALESCE(bp.code_subfunction, '') AS subfunction,
bp.code_contract || bp.code_function || bp.code_number as program,
bp.code_contract || bp.code_function || bp.code_number || ba.code as action,
ba.name AS action_name FROM 
budget_chapter bc, budget_program bp, budget_action ba, budget_sector bs
WHERE ba.program_id=bp.id AND bp.chapter_id=bc.id AND ba.sector_id=bs.id) b
ON cq_union.action_id=b.id
GROUP BY 1,2,3,4,5,6
ORDER BY 1,2,3,4,5,6`)
	var r ScenarioActionPayment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Chapter, &r.Sector, &r.Subfunction, &r.Program, &r.Action,
			&r.ActionName, &r.Y1, &r.Y2, &r.Y3); err != nil {
			return err
		}
		s.ScenarioStatActionPayments = append(s.ScenarioStatActionPayments, r)
	}
	err = rows.Err()
	return err
}
