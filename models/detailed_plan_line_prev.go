package models

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// DetailedPlanLineAndPrevisions contains json formatted result of the
// dedicated query
type DetailedPlanLineAndPrevisions struct {
	DetailedPlanLineAndPrevisions json.RawMessage `json:"DetailedPlanLine"`
}

// GetAll populates the DetailedPlanLineAndPrevisions of the given plan
func (d *DetailedPlanLineAndPrevisions) GetAll(plan *Plan, db *sql.DB) (err error) {
	firstYear, lastYear, err := plan.GetFirstAndLastYear(db)
	if err != nil {
		return err
	}
	var pp, nn, cc, ll, jj []string
	for year := firstYear; year <= lastYear; year++ {
		sy := strconv.FormatInt(year, 10)
		pp = append(pp, `fc."`+sy+`"`)
		nn = append(nn, `NULL::bigint AS"`+sy+`"`)
		cc = append(cc, `"`+sy+`" bigint`)
		ll = append(ll, `"`+sy+`"`)
		jj = append(jj, `'`+sy+`', q."`+sy+`"`)
	}
	prevQry := strings.Join(pp, ",")
	nullQry := strings.Join(nn, ",")
	convertQry := strings.Join(cc, ",")
	colQry := strings.Join(ll, ",")
	jsonQry := strings.Join(jj, ",")
	actualYear := strconv.Itoa(time.Now().Year())
	var prevPart string
	if lastYear >= firstYear {
		prevPart = `
		UNION ALL
		SELECT op.number AS op_number, op.name as op_name, step.name AS step,
		category.name as category, NULL AS commitment_name, 
			NULL AS commitment_code, NULL AS commitment_date, NULL AS commitment_value, 
			NULL AS programmings_value, NULL AS programmings_date, ` + colQry + `, op.plan_line_id
		FROM crosstab (
			'SELECT physical_op_id, year, value FROM prev_commitment ORDER BY 1,2',
			'SELECT m FROM generate_series(` + strconv.FormatInt(firstYear, 10) + `, ` +
			strconv.FormatInt(lastYear, 10) + `) AS m')
			AS (physical_op_id INTEGER, ` + convertQry + `) 
		JOIN physical_op op ON physical_op_id = op.id 
		LEFT OUTER JOIN step ON op.step_id=step.id
		LEFT OUTER JOIN category ON op.category_id=category.id
			`
		prevQry = "," + prevQry
		jsonQry = "," + jsonQry
		nullQry = "," + nullQry
	}

	finalQry := `SELECT json_build_object('id', q.id, 'name', q.name, 
	'op_name', q.op_name, 'value', q.value, 'total_value', q.total_value,
	'op_number',q.op_number,'step',q.step,'category',q.category, 'commitment_name', q.commitment_name, 
	'commitment_code', q.commitment_code, 'commitment_date', q.commitment_date, 
	'commitment_value', q.commitment_value, 'programmings_value', q.programmings_value, 
	'programmings_date', q.programmings_date	` + jsonQry + `) FROM
	(SELECT pl.id, pl.name, pl.total_value, pl.value, fc.op_number, fc.op_name,fc.Step,fc.category,
		fc.commitment_name, fc.commitment_code, fc.commitment_date, fc.commitment_value,   
		fc.programmings_value, fc.programmings_date ` + prevQry + ` FROM plan_line pl
	LEFT OUTER JOIN 
	(SELECT op.number AS op_number, op.name as op_name, step.name AS step,
		category.name as category, f.name AS commitment_name, 
		f.iris_code AS commitment_code, f.date AS commitment_date, f.value AS commitment_value, 
		NULL AS programmings_value,NULL AS programmings_date ` + nullQry + `, f.plan_line_id 
		FROM financial_commitment f
		JOIN physical_op op ON f.physical_op_id = op.id 
		LEFT OUTER JOIN step ON op.step_id=step.id
		LEFT OUTER JOIN category ON op.category_id=category.id
	WHERE EXTRACT(year FROM f.date) < ` + actualYear + ` AND f.plan_line_id NOTNULL 
	UNION ALL
	SELECT op.number AS op_number, op.name as op_name,step.name AS step,
	category.name as category,  NULL AS commitment_name, 
		NULL AS commitment_code, NULL AS commitment_date, NULL AS commitment_value, 
		p.value AS programmings_value, c.date AS programmings_date ` + nullQry + `, 
		op.plan_line_id
	FROM programmings p
	JOIN commissions c ON c.id=p.commission_id
	JOIN physical_op op ON p.physical_op_id = op.id 
	LEFT OUTER JOIN step ON op.step_id=step.id
	LEFT OUTER JOIN category ON op.category_id=category.id
	WHERE p.year=` + actualYear + prevPart + ` 
	) fc ON fc.plan_line_id = pl.id
	WHERE pl.plan_id = ` + strconv.FormatInt(plan.ID, 10) + `
	ORDER BY 1,5,9,12) q`
	lines, line := []string{}, ""
	rows, err := db.Query(finalQry)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&line); err != nil {
			return err
		}
		lines = append(lines, line)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	d.DetailedPlanLineAndPrevisions = json.RawMessage("[" + strings.Join(lines, ",") + "]")
	return err
}
