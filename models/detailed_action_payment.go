package models

import (
	"database/sql"
	"strconv"
)

// DetailedActionPayment is used to decode a line of dedicated query.
type DetailedActionPayment struct {
	Chapter     NullInt64   `json:"chapter"`
	Sector      NullString  `json:"sector"`
	Subfunction NullString  `json:"subfunction"`
	Program     NullString  `json:"program"`
	Action      NullString  `json:"action"`
	ActionName  NullString  `json:"action_name"`
	Number      string      `json:"number"`
	Name        string      `json:"name"`
	Y1          NullFloat64 `json:"y1"`
	Y2          NullFloat64 `json:"y2"`
	Y3          NullFloat64 `json:"y3"`
}

// DetailedActionPayments embeddes an array of DetailedActionPayment for json export.
type DetailedActionPayments struct {
	DetailedActionPayments []DetailedActionPayment `json:"DetailedPaymentPerBudgetAction"`
}

// StatDetailedActionPayments embeddes an array of DetailedActionPayment for json export.
type StatDetailedActionPayments struct {
	DetailedActionPayments []DetailedActionPayment `json:"StatisticalDetailedPaymentPerBudgetAction"`
}

// GetAll fetches payments previsions per budget actions since given year and using
// given payment types from database.
func (d *DetailedActionPayments) GetAll(year int64, ptID int64, db *sql.DB) (err error) {
	sy := strconv.FormatInt(year, 10)
	sptID := strconv.FormatInt(ptID, 10)
	rows, err := db.Query(`SELECT b.chapter, b.sector, b.subfunction, b.program, b.action, 
	b.action_name, cq_union.number, cq_union.name, cq_union.y1, cq_union.y2, cq_union.y3 FROM
	(
	 (SELECT op.budget_action_id AS action_id, op.number, op.name, SUM(ct.y1)*0.01 AS y1,
	 	 SUM(ct.y2)*0.01 AS y2, SUM(ct.y3)*0.01 AS y3 FROM
			crosstab(
				'WITH pr AS (SELECT * FROM payment_ratios WHERE payment_types_id=` + sptID + `),
					pp AS (SELECT physical_op_id, year, value FROM prev_payment 
						WHERE value NOTNULL AND value<>0 AND year>=` + sy + ` AND year<=` + sy + `+2),
					pp_idx AS (SELECT physical_op_id, year FROM pp),
					fc_sum AS (SELECT physical_op_id, EXTRACT(year FROM date)::integer AS year, 
						SUM(value) AS value FROM financial_commitment 
						WHERE EXTRACT(year FROM date)<` + sy + `-1 GROUP BY 1,2),
					fc AS (SELECT fc_sum.physical_op_id, fc_sum.year+pr.index AS year, 
						fc_sum.value*pr.ratio AS value FROM fc_sum, pr 
						WHERE fc_sum.year+pr.index>=` + sy + ` AND fc_sum.year+pr.index<=` + sy + `+2),
					fc_filtered AS (SELECT * FROM fc WHERE fc.physical_op_id IS NOT NULL 
						AND (fc.physical_op_id, fc.year) NOT IN (SELECT * FROM pp_idx)),
					pg_year AS (SELECT physical_op_id, year, SUM(value) AS value 
						FROM programmings WHERE year = ` + sy + ` - 1 GROUP BY 1,2),
					pg AS (SELECT pg_year.physical_op_id, pg_year.year+pr.index AS year, 
						pg_year.value*pr.ratio AS value FROM pg_year, pr 
						WHERE pg_year.year+pr.index>=` + sy + ` AND pg_year.year+pr.index<=` + sy + `+2),
					pg_filtered AS (SELECT * FROM pg 
							WHERE (pg.physical_op_id, pg.year) NOT IN (SELECT * FROM pp_idx)),
					pc AS (SELECT p.physical_op_id, p.year+pr.index AS year, p.value*pr.ratio AS value 
						FROM prev_commitment p, pr WHERE p.year+pr.index>=` + sy + ` 
						AND p.year+pr.index<=` + sy + `+2),
					pc_filtered AS (SELECT * FROM pc WHERE (pc.physical_op_id, pc.year) 
						NOT IN (SELECT * FROM pp_idx))
				SELECT * FROM
				(SELECT * FROM pp
				UNION ALL
				SELECT physical_op_id, year, SUM(value) AS value FROM 
					(SELECT * FROM fc_filtered UNION ALL SELECT * FROM pg_filtered
					UNION ALL
					SELECT * FROM pc_filtered) q1
				GROUP BY 1,2) q2 ORDER BY 1,2',
				'SELECT m FROM generate_series(` + sy + `, ` + sy + `+2) AS m') 
			AS ct(physical_op_id integer, y1 numeric, y2 numeric, y3 numeric)
		LEFT JOIN physical_op op ON op.id = ct.physical_op_id
	 	GROUP BY 1,2,3)
	  UNION ALL
		(SELECT action_id, NULL AS number, NULL AS name, SUM(y1)*0.01 AS y1, 
		SUM(y2)*0.01 AS y2, SUM(y3)*0.01 AS y3 FROM
			crosstab(
					'WITH pr AS (SELECT * FROM payment_ratios WHERE payment_types_id=` + sptID + `),
						unlinked_fc_sum AS (SELECT action_id, EXTRACT(year FROM date)::integer AS year, 
							SUM(value) AS value FROM financial_commitment WHERE EXTRACT(year FROM date) <` + sy + `-1
							AND physical_op_id IS NULL GROUP BY 1,2),
						unlinked_fc AS (SELECT ufs.action_id, ufs.year+pr.index AS year, 
							ufs.value*pr.ratio AS value FROM unlinked_fc_sum ufs, pr 
							WHERE ufs.year+pr.index>=` + sy + ` AND ufs.year+pr.index<=` + sy + `+2)
					SELECT action_id, year, SUM(value) FROM unlinked_fc GROUP BY 1,2 ORDER BY 1,2',
					'SELECT m FROM generate_series(` + sy + `, ` + sy + ` + 2) AS m')
			AS (action_id integer, y1 numeric, y2 numeric, y3 numeric)
	 GROUP BY 1, 2, 3)
	) cq_union
	LEFT JOIN (SELECT ba.id, bc.code AS chapter, bs.code AS sector, 
		bp.code_function || COALESCE(bp.code_subfunction, '') AS subfunction,
		bp.code_contract || bp.code_function || bp.code_number as program,
		bp.code_contract || bp.code_function || bp.code_number || ba.code as action, 
		ba.name AS action_name FROM 
		budget_chapter bc, budget_program bp, budget_action ba, budget_sector bs
		WHERE ba.program_id = bp.id AND bp.chapter_id = bc.id AND ba.sector_id = bs.id) b
	ON cq_union.action_id = b.id
	ORDER BY 1,2,3,4,5,6,7,8`)
	if err != nil {
		return err
	}
	var r DetailedActionPayment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Chapter, &r.Sector, &r.Subfunction, &r.Program, &r.Action, &r.ActionName,
			&r.Number, &r.Name, &r.Y1, &r.Y2, &r.Y3); err != nil {
			return err
		}
		d.DetailedActionPayments = append(d.DetailedActionPayments, r)
	}
	err = rows.Err()
	return err
}

// GetAll fetches payments previsions per budget actions since given year and using
// given payment types from database with a pure statistical approach i.e. without
//taking payment prevision datas into account.
func (d *StatDetailedActionPayments) GetAll(year int64, ptID int64, db *sql.DB) (err error) {
	sy := strconv.FormatInt(year, 10)
	sptID := strconv.FormatInt(ptID, 10)
	rows, err := db.Query(`SELECT b.chapter, b.sector, b.subfunction, b.program, b.action,
		b.action_name, cq_union.number, cq_union.name, cq_union.y1, cq_union.y2, cq_union.y3 
	FROM (
		(SELECT op.budget_action_id AS action_id, op.number, op.name, SUM(ct.y1)*0.01 AS y1,
			SUM(ct.y2)*0.01 AS y2, SUM(ct.y3)*0.01 AS y3 FROM
			crosstab(
				'WITH pr AS (SELECT * FROM payment_ratios WHERE payment_types_id=` + sptID + `),
						fc_sum AS (SELECT physical_op_id, EXTRACT(year FROM date)::integer AS year, 
						SUM(value) AS value FROM financial_commitment 
						WHERE EXTRACT(year FROM date) <` + sy + ` - 1 GROUP BY 1,2),
						fc AS (SELECT fc_sum.physical_op_id, fc_sum.year+pr.index AS year, 
							fc_sum.value*pr.ratio AS value FROM fc_sum, pr 
							WHERE fc_sum.year+pr.index>=` + sy + ` AND fc_sum.year+pr.index<=` + sy + `+2),
						fc_filtered AS (SELECT * FROM fc WHERE fc.physical_op_id NOTNULL),
						pg_year AS (SELECT physical_op_id, year, SUM(value) AS value 
							FROM programmings WHERE year=` + sy + `-1 GROUP BY 1,2),
						pg AS (SELECT pg_year.physical_op_id, pg_year.year+pr.index AS year, 
							pg_year.value*pr.ratio AS value FROM pg_year, pr 
							WHERE pg_year.year+pr.index>=` + sy + ` AND pg_year.year+pr.index<=` + sy + `+2),
						pc AS (SELECT p.physical_op_id, p.year+pr.index AS year,
							p.value*pr.ratio AS value FROM prev_commitment p, pr
							WHERE p.year+pr.index>=` + sy + ` AND p.year+pr.index<=` + sy + `+2)
				SELECT physical_op_id, year, SUM(value) AS value FROM 
					(SELECT * FROM fc_filtered UNION ALL SELECT * FROM pg UNION ALL SELECT * FROM pc) q1 
					GROUP BY 1,2 ORDER BY 1,2',
				'SELECT m FROM generate_series(` + sy + `, ` + sy + `+2) AS m') 
			AS ct(physical_op_id integer, y1 numeric, y2 numeric, y3 numeric)
			LEFT JOIN physical_op op ON op.id=ct.physical_op_id
	 GROUP BY 1,2,3)
	 UNION ALL
	 (SELECT action_id, NULL AS number, NULL AS name, SUM(y1)*0.01 AS y1, 
	 	SUM(y2)*0.01 AS y2, SUM(y3)*0.01 AS y3 FROM
		crosstab(
			'WITH pr AS (SELECT * FROM payment_ratios WHERE payment_types_id=` + sptID + `),
				unlinked_fc_sum AS (SELECT action_id, EXTRACT(year FROM date)::integer AS year,
					SUM(value) AS value FROM financial_commitment 
					WHERE EXTRACT(year FROM date)<` + sy + `-1 AND physical_op_id ISNULL GROUP BY 1,2),
				unlinked_fc AS (SELECT ufs.action_id, ufs.year+pr.index AS year, 
					ufs.value*pr.ratio AS value FROM unlinked_fc_sum ufs, pr 
					WHERE ufs.year+pr.index>=` + sy + ` AND ufs.year+pr.index<=` + sy + `+2)
			SELECT action_id, year, SUM(value) FROM unlinked_fc GROUP BY 1,2 ORDER BY 1,2',
			'SELECT m FROM generate_series(` + sy + `, ` + sy + `+2) AS m')
		AS (action_id integer, y1 numeric, y2 numeric, y3 numeric)
	  GROUP BY 1, 2, 3)
	) cq_union
	LEFT JOIN (SELECT ba.id, bc.code AS chapter, bs.code AS sector, 
		bp.code_function || COALESCE(bp.code_subfunction, '') AS subfunction,
		bp.code_contract || bp.code_function || bp.code_number as program,
		bp.code_contract || bp.code_function || bp.code_number || ba.code as action,
		ba.name AS action_name FROM budget_chapter bc, budget_program bp, 
		budget_action ba, budget_sector bs
		WHERE ba.program_id = bp.id AND bp.chapter_id = bc.id AND ba.sector_id = bs.id) b
	ON cq_union.action_id = b.id
	ORDER BY 1,2,3,4,5,6,7,8;`)
	if err != nil {
		return err
	}
	var r DetailedActionPayment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Chapter, &r.Sector, &r.Subfunction, &r.Program, &r.Action, &r.ActionName,
			&r.Number, &r.Name, &r.Y1, &r.Y2, &r.Y3); err != nil {
			return err
		}
		d.DetailedActionPayments = append(d.DetailedActionPayments, r)
	}
	err = rows.Err()
	return err
}
