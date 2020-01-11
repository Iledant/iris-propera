package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// PlanLineAndPrevisions is used to store the dedicated query results.
type PlanLineAndPrevisions struct {
	PlanLineAndPrevisions json.RawMessage `json:"PlanLine"`
}

// GetAll fetches plan line datas and previsions attached to a specific
// plan line or for all plan lines of a plan
func (p *PlanLineAndPrevisions) GetAll(plan *Plan, plID int64, db *sql.DB) (err error) {
	firstYear, lastYear, err := plan.GetFirstAndLastYear(db)
	if err != nil {
		return err
	}
	var whereQry, beneficiaryIdsQry string
	if plID != 0 {
		whereQry = "p.id=" + strconv.FormatInt(plID, 10) + " "
		beneficiaryIdsQry = `SELECT DISTINCT beneficiary_id FROM plan_line_ratios 
		WHERE plan_line_id=` + strconv.FormatInt(plID, 10) + " "
	} else {
		whereQry = "p.plan_id=" + strconv.FormatInt(plan.ID, 10) + " "
		beneficiaryIdsQry = `SELECT DISTINCT beneficiary_id FROM plan_line_ratios 
		WHERE plan_line_id IN (SELECT id FROM plan_line WHERE plan_id=` +
			strconv.FormatInt(plan.ID, 10) + ") "
	}
	bIDs, bID := []int64{}, int64(0)
	rows, err := db.Query(beneficiaryIdsQry)
	if err != nil {
		return fmt.Errorf("beneficiary select %v ", err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&bID); err != nil {
			return err
		}
		bIDs = append(bIDs, bID)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	benQry, jsonBenQry, benCrossQry := "", "", ""
	if len(bIDs) > 0 {
		bb, bc, jj := []string{}, []string{}, []string{}
		for _, bID := range bIDs {
			sbID := strconv.FormatInt(bID, 10)
			bb = append(bb, "b.b"+sbID)
			jj = append(jj, "'b"+sbID+"', q.b"+sbID)
			bc = append(bc, "b"+sbID+" double precision")
		}
		benQry = ", " + strings.Join(bb, ",")
		jsonBenQry = ", " + strings.Join(jj, ",")
		benCrossQry = strings.Join(bc, ",")
		benCrossQry = ` LEFT JOIN (SELECT * FROM crosstab('SELECT plan_line_id, 
		beneficiary_id, ratio FROM plan_line_ratios ORDER BY 1,2', '` +
			beneficiaryIdsQry + "') AS (plan_line_id INTEGER, " + benCrossQry +
			")) b ON b.plan_line_id = p.id"
	} else {
		benQry = ""
		jsonBenQry = ""
		benCrossQry = ""
	}
	pp, cc, jj := []string{}, []string{}, []string{}
	for year := firstYear; year <= lastYear; year++ {
		sy := strconv.FormatInt(year, 10)
		pp = append(pp, `prev."`+sy+`"`)
		cc = append(cc, `"`+sy+`" bigint`)
		jj = append(jj, `'`+sy+`', q."`+sy+`"`)
	}
	prevQry := strings.Join(pp, ",")
	convertQry := strings.Join(cc, ",")
	jsonQry := strings.Join(jj, ",")

	actualYear := strconv.Itoa(time.Now().Year())
	var prevPart string
	if lastYear >= firstYear {
		prevQry = "," + prevQry
		jsonQry = "," + jsonQry
		prevPart = `
		LEFT JOIN (SELECT * FROM 
			crosstab ('SELECT p.plan_line_id, c.year, SUM(c.value) FROM physical_op p, prev_commitment c 
									WHERE p.id = c.physical_op_id AND p.plan_line_id NOTNULL GROUP BY 1,2 ORDER BY 1,2',
								'SELECT m FROM generate_series( ` + strconv.FormatInt(firstYear, 10) + `, ` + strconv.FormatInt(lastYear, 10) + `) AS m')
				AS (plan_line_id INTEGER, ` + convertQry + `)) prev 
		ON prev.plan_line_id = p.id`
	}
	finalQry := `SELECT json_build_object('id',q.id,'name',q.name, 'descript', q.descript,'value', q.value, 
	'total_value', q.total_value, 'commitment', q.commitment, 'programmings', q.programmings` + jsonQry + jsonBenQry + ` ) FROM
	(SELECT p.id, p.name, p.descript, p.value, p.total_value, 
	CAST(fc.value AS bigint) AS commitment, CAST(pr.value AS bigint) AS programmings ` + prevQry + benQry + `
FROM plan_line p` + benCrossQry + `
LEFT JOIN (SELECT f.plan_line_id, SUM(f.value) AS value FROM financial_commitment f
						WHERE f.plan_line_id NOTNULL AND EXTRACT(year FROM f.date) < ` + actualYear + `
						GROUP BY 1) fc
	ON fc.plan_line_id = p.id
LEFT JOIN (SELECT op.plan_line_id, SUM(p.value) AS value FROM physical_op op, programmings p 
						WHERE p.physical_op_id = op.id AND p.year = ` + actualYear + ` GROUP BY 1) pr 
	ON pr.plan_line_id = p.id` + prevPart + `
WHERE ` + whereQry + ` ORDER BY 1) q`
	lines, line := []string{}, ""
	rows, err = db.Query(finalQry)
	if err != nil {
		return fmt.Errorf("select json %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&line); err != nil {
			return fmt.Errorf("scan %v", err)
		}
		lines = append(lines, line)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	if plID == 0 {
		p.PlanLineAndPrevisions = json.RawMessage("[" + strings.Join(lines, ",") + "]")
	} else {
		p.PlanLineAndPrevisions = json.RawMessage(strings.Join(lines, ","))
	}
	return err
}
