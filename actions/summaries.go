package actions

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// GetMultiannualProgrammation handles theget request to fetch multiannual programmation.
func GetMultiannualProgrammation(ctx iris.Context) {
	y1, err := ctx.URLParamInt64("y1")
	if err != nil {
		y1 = int64(time.Now().Year())
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	var lastYear int64
	if err = db.DB().QueryRow("select max(year) from prev_commitment").Scan(&lastYear); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Programmation pluriannuelle, requête max year : " + err.Error()})
		return
	}
	if lastYear < y1+4 {
		lastYear = y1 + 4
	}
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS tablefunc").Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Programmation pluriannuelle, tablefunc : " + err.Error()})
		return
	}
	var columnNames, typesNames, jsonNames []string
	for i := y1; i <= lastYear; i++ {
		year := strconv.FormatInt(i-y1, 10)
		columnNames = append(columnNames, `"`+strconv.FormatInt(i, 10)+`" AS y`+year)
		typesNames = append(typesNames, `"`+strconv.FormatInt(i, 10)+`" VARCHAR`)
		jsonNames = append(jsonNames, `'y`+year+`', q.y`+year)
	}

	qry := `SELECT json_build_object('number',q.number,'name',q.name,'step_name', q.step_name, 'category_name',
	 q.category_name,` + strings.Join(jsonNames, ",") + `) FROM
	(SELECT op.number, op.name, s.name as step_name, cat.name as category_name, ` + strings.Join(columnNames, ",") + ` FROM 
	crosstab('SELECT physical_op_id, year, row_to_json((SELECT d FROM (SELECT value, total_value, state_ratio) d)) 
						FROM prev_commitment ORDER BY 1,2', 
					'SELECT m FROM generate_series(` + strconv.FormatInt(y1, 10) + `,` + strconv.FormatInt(lastYear, 10) + `) AS m') AS
					(op_id INTEGER, ` + strings.Join(typesNames, ",") + `)
	JOIN physical_op op ON op.id = op_id
	LEFT OUTER JOIN step s ON op.step_id = s.id
	LEFT OUTER JOIN category cat ON op.category_id = cat.id) q`

	lines, line := []string{}, ""
	rows, err := db.DB().Query(qry)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Programmation pluriannuelle, requête finale : " + err.Error()})
		return
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&line); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Programmation pluriannuelle, lecture des lignes : " + err.Error()})
			return
		}
		lines = append(lines, line)
	}

	resp := `{"MultiannualProgrammation":[` + strings.Join(lines, ",") + `]}`

	ctx.StatusCode(http.StatusOK)
	ctx.ContentType("application/json")
	ctx.Write([]byte(resp))
}

// annualProg is used to decode one row of the annual programmation query
type annualProg struct {
	OperationNumber   models.NullString  `json:"operation_number" gorm:"column:operation_number"`
	Name              models.NullString  `json:"name" gorm:"column:name"`
	StepName          models.NullString  `json:"step_name" gorm:"column:step_name"`
	CategoryName      models.NullString  `json:"category_name" gorm:"column:category_name"`
	Date              time.Time          `json:"date" gorm:"column:date"`
	Programmings      models.NullInt64   `json:"programmings" gorm:"column:programmings"`
	TotalProgrammings models.NullInt64   `json:"total_programmings" gorm:"column:total_programmings"`
	StateRatio        models.NullFloat64 `json:"state_ratio" gorm:"column:state_ratio"`
	Commitment        models.NullInt64   `json:"commitment" gorm:"column:commitment"`
	Pendings          models.NullInt64   `json:"pendings" gorm:"column:pendings"`
}

// annualProgResp embeddes an array of annualProg for the annual programmation response.
type annualProgResp struct {
	AnnualProgrammation []annualProg       `json:"AnnualProgrammation"`
	ImportLog           []models.ImportLog `json:"ImportLog"`
}

// GetAnnualProgrammation handles the get request to fetch datas comparing programmation, commitments and pending commitments.
func GetAnnualProgrammation(ctx iris.Context) {
	year, err := ctx.URLParamInt("year")
	if err != nil {
		year = time.Now().Year()
	}
	db := ctx.Values().Get("db").(*gorm.DB)
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
	rows, err := db.DB().Query(qry, year)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Programmation annuelle, requête : " + err.Error()})
		return
	}
	defer rows.Close()
	resp, p := annualProgResp{}, annualProg{}
	for rows.Next() {
		if err = rows.Scan(&p.OperationNumber, &p.Name, &p.StepName, &p.CategoryName, &p.Date, &p.Programmings,
			&p.TotalProgrammings, &p.StateRatio, &p.Commitment, &p.Pendings); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Programmation annuelle, lecture de ligne : " + err.Error()})
			return
		}
		resp.AnnualProgrammation = append(resp.AnnualProgrammation, p)
	}
	if err = db.Find(&resp.ImportLog).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Programmation annuelle, import logs : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetProgrammingAndPrevisions handles the get request to compare precisely programmation and previsions.
func GetProgrammingAndPrevisions(ctx iris.Context) {
	year, err := ctx.URLParamInt64("y1")
	if err != nil {
		year = int64(time.Now().Year())
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	qry := `SELECT json_build_object('number',op.number, 'name',op.name,'programmings', pr.value , 'prevision', pc.value ) FROM physical_op op
	LEFT OUTER JOIN
	(SELECT p.physical_op_id, SUM(value) AS value FROM programmings p, commissions c 
	WHERE p.commission_id = c.id AND extract(year FROM c.date) = $1 GROUP BY 1) pr
	ON op.id = pr.physical_op_id
	LEFT OUTER JOIN
	(SELECT f.physical_op_id, value FROM prev_commitment f WHERE year = $1) pc
	ON op.id = pc.physical_op_id
	WHERE pr.value NOTNULL or (pc.value NOTNULL AND pc.value <> 0)
	ORDER BY op.number`
	lines, line := []string{}, ""
	rows, err := db.DB().Query(qry, year)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Comparaison programmation prévision, requête : " + err.Error()})
		return
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&line); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Comparaison programmation prévision, lecture des lignes : " + err.Error()})
			return
		}
		lines = append(lines, line)
	}

	resp := `{"ProgrammingsPrevision":[` + strings.Join(lines, ",") + `]}`

	ctx.StatusCode(http.StatusOK)
	ctx.ContentType("application/json")
	ctx.Write([]byte(resp))
}

// actionProg is used to decode results of the programmation by budget action query.
type actionProg struct {
	ActionCode models.NullString `json:"action_code"`
	ActionName models.NullString `json:"action_name"`
	Value      int64             `json:"value"`
}

type actionProgResp struct {
	BudgetProgrammation []actionProg `json:"BudgetProgrammation"`
}

// GetActionProgrammation handles the get request to fetch the programmation by budget actions.
func GetActionProgrammation(ctx iris.Context) {
	year, err := ctx.URLParamInt64("y1")
	if err != nil {
		year = int64(time.Now().Year())
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	qry := `SELECT b.action_code, b.name AS action_name, SUM(p.value) AS value FROM physical_op op
	JOIN programmings p ON p.physical_op_id = op.id 
	LEFT OUTER JOIN
	(SELECT ba.id, bp.code_contract||bp.code_function||bp.code_number||COALESCE(bp.code_subfunction,'')||ba.code as action_code, ba.name
		FROM budget_program bp, budget_action ba
		WHERE ba.program_id = bp.id) b
		ON op.budget_action_id = b.id
	WHERE p.year = ?
	GROUP BY 1,2 ORDER BY substring(b.action_code from 2), substring(b.action_code for 1)`
	rows, err := db.Raw(qry, year).Rows()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Programmation par action, requête : " + err.Error()})
		return
	}
	defer rows.Close()
	resp, p := actionProgResp{}, actionProg{}
	for rows.Next() {
		if err = db.ScanRows(rows, &p); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Programmation par action, lecture de ligne : " + err.Error()})
			return
		}
		resp.BudgetProgrammation = append(resp.BudgetProgrammation, p)
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetActionCommitment handles the get request to fetch prevision of payment by budget actions.
func GetActionCommitment(ctx iris.Context) {
	y1, err := ctx.URLParamInt64("FirstYear")
	if err != nil {
		y1 = int64(time.Now().Year()) + 1
	}
	y2 := y1 + 1
	y3 := y1 + 2
	py := y1 - 1
	sy1 := strconv.FormatInt(y1, 10)
	sy2 := strconv.FormatInt(y2, 10)
	sy3 := strconv.FormatInt(y3, 10)
	spy := strconv.FormatInt(py, 10)
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = db.Exec("CREATE EXTENSION IF NOT EXISTS tablefunc").Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Engagement par action, tablefunc : " + err.Error()})
		return
	}
	qry := `WITH budget as (SELECT ba.id, bc.code AS chapter, bs.code AS sector, 
		bp.code_function || COALESCE(bp.code_subfunction, '') AS subfunction,
		bp.code_contract || bp.code_function || bp.code_number as program,
		bp.code_contract || bp.code_function || bp.code_number || ba.code as action, 
		ba.name AS action_name 
	FROM budget_chapter bc, budget_program bp, budget_action ba, budget_sector bs
	WHERE ba.program_id = bp.id AND bp.chapter_id = bc.id AND ba.sector_id = bs.id) 
	SELECT json_build_object('chapter', q.chapter, 'sector', q.sector, 'subfunction', q.subfunction,
	'program', q.program, 'action', q.action, 'action_name',q.action_name, 'y` + spy + `', q.y` + spy + `, 'y` +
		sy1 + `', q.y` + sy1 + `, 'y` + sy2 + `', q.y` + sy2 + `,'y` + sy3 + `',q.y` + sy3 + `) FROM
(SELECT budget.chapter, budget.sector, budget.subfunction, budget.program, budget.action, budget.action_name,
SUM(y` + spy + `) AS y` + spy + `, SUM(tot.y` + sy1 + `) AS y` + sy1 + `, SUM(tot.y` + sy2 + `) AS y` + sy2 + `, 
SUM(tot.y` + sy3 + `) AS y` + sy3 + `
FROM 
(SELECT *, NULL as y` + spy + ` 
FROM crosstab('SELECT op.budget_action_id, pc.year, SUM(pc.value) * 0.01 
		FROM
			(SELECT * FROM prev_commitment WHERE year >= ` + sy1 +
		` AND year <= ` + sy3 + `) pc, physical_op op 
		WHERE pc.physical_op_id = op.id GROUP BY 1,2 ORDER BY 1,2', 
		'SELECT m FROM generate_series(` + sy1 + `, ` + sy3 + `) AS m')
AS (budget_action_id INTEGER, y` + sy1 + ` NUMERIC, y` + sy2 + ` NUMERIC, y` + sy3 + ` NUMERIC)
UNION ALL 
SELECT op.budget_action_id, NULL as y` + sy1 + `, NULL as y` + sy2 +
		`, NULL as y` + sy3 + `, SUM(pg.value) * 0.01 AS y` + spy + `
FROM programmings pg, physical_op op
WHERE pg.year = ` + spy + ` AND pg.physical_op_id = op.id GROUP BY 1) tot, budget
WHERE tot.budget_action_id = budget.id
GROUP BY 1,2,3,4,5,6 ORDER BY 1,2,3,4,5) q`
	rows, err := db.DB().Query(qry)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Engagement par action, requête : " + err.Error()})
		return
	}
	defer rows.Close()
	lines, line := []string{}, ""
	for rows.Next() {
		if err = rows.Scan(&line); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Engagement par action, lecture de ligne : " + err.Error()})
			return
		}
		lines = append(lines, line)
	}
	resp := `{"CommitmentPerBudgetAction":[` + strings.Join(lines, ",") + `]}`

	ctx.StatusCode(http.StatusOK)
	ctx.ContentType("application/json")
	ctx.Write([]byte(resp))
}

// GetDetailedActionCommitment handles the get request to have detailed commitment per budget actions.
func GetDetailedActionCommitment(ctx iris.Context) {
	y1, err := ctx.URLParamInt64("FirstYear")
	if err != nil {
		y1 = int64(time.Now().Year()) + 1
	}
	y2 := y1 + 1
	y3 := y1 + 2
	py := y1 - 1
	sy1 := strconv.FormatInt(y1, 10)
	sy2 := strconv.FormatInt(y2, 10)
	sy3 := strconv.FormatInt(y3, 10)
	spy := strconv.FormatInt(py, 10)
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = db.Exec("CREATE EXTENSION IF NOT EXISTS tablefunc").Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Engagement détaillé par action, tablefunc : " + err.Error()})
		return
	}
	qry := `SELECT json_build_object('chapter', q.chapter, 'sector', q.sector, 'subfunction', q.subfunction,
	'program', q.program, 'action', q.action, 'action_name',q.action_name, 'number', q.number, 'name', q.name, 'y` + spy + `', q.y` + spy + `, 'y` +
		sy1 + `', q.y` + sy1 + `, 'y` + sy2 + `', q.y` + sy2 + `,'y` + sy3 + `',q.y` + sy3 + `) FROM
		(SELECT budget.chapter, budget.sector, budget.subfunction, budget.program, budget.action, 
	budget.action_name, op.number, op.name, pg.value AS y` + spy + `, ct.y` + sy1 + `, ct.y` + sy2 + `, ct.y` + sy3 + ` FROM 
	physical_op op
	LEFT OUTER JOIN (SELECT * FROM crosstab('SELECT pc.physical_op_id, pc.year, pc.value * 0.01 FROM 
		(SELECT * FROM prev_commitment WHERE year >= ` + sy1 + ` AND year <=` + sy3 + `) pc ORDER BY 1,2',
'SELECT m FROM generate_series(` + sy1 + `,` + sy3 + `) AS m') AS (physical_op_id INTEGER, y` + sy1 + ` NUMERIC, y` + sy2 + ` NUMERIC, y` + sy3 + ` NUMERIC)) ct
ON ct.physical_op_id = op.id 
LEFT OUTER JOIN (SELECT physical_op_id, SUM(value) * 0.01 AS value FROM programmings WHERE year = ` + spy + ` GROUP BY 1) pg ON pg.physical_op_id = op.id
LEFT OUTER JOIN 
(SELECT ba.id, bc.code AS chapter, bs.code AS sector, bp.code_function || COALESCE(bp.code_subfunction, '') AS subfunction,
			bp.code_contract || bp.code_function || bp.code_number as program,
			bp.code_contract || bp.code_function || bp.code_number || ba.code as action, ba.name AS action_name FROM 
					budget_chapter bc, budget_program bp, budget_action ba, budget_sector bs
					WHERE ba.program_id = bp.id AND bp.chapter_id = bc.id AND ba.sector_id = bs.id) AS budget
ON op.budget_action_id = budget.id
WHERE pg.value IS NOT NULL OR (ct.y` + sy1 + ` <> 0 AND ct.y` + sy1 + ` IS NOT NULL) OR (ct.y` + sy2 + ` <> 0 AND ct.y` + sy2 + ` IS NOT NULL) OR (ct.y` + sy3 + ` <> 0 AND ct.y` + sy3 + ` IS NOT NULL)
ORDER BY 1, 2, 3, 4, 5) q`
	rows, err := db.DB().Query(qry)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Engagement détaillé par action, requête : " + err.Error()})
		return
	}
	defer rows.Close()
	lines, line := []string{}, ""
	for rows.Next() {
		if err = rows.Scan(&line); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Engagement détaillé par action, lecture de ligne : " + err.Error()})
			return
		}
		lines = append(lines, line)
	}
	resp := `{"DetailedCommitmentPerBudgetAction":[` + strings.Join(lines, ",") + `]}`

	ctx.StatusCode(http.StatusOK)
	ctx.ContentType("application/json")
	ctx.Write([]byte(resp))
}

// GetDetailedActionPayment handles the get request to get payment prevision by physical operation.
func GetDetailedActionPayment(ctx iris.Context) {
	y1, err := ctx.URLParamInt64("FirstYear")
	if err != nil {
		y1 = int64(time.Now().Year()) + 1
	}
	dID, err := ctx.URLParamInt64("DefaultPaymentTypeId")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiement détaillé par action, décodage : " + err.Error()})
		return
	}
	y2 := y1 + 1
	y3 := y1 + 2
	sy1 := strconv.FormatInt(y1, 10)
	sy2 := strconv.FormatInt(y2, 10)
	sy3 := strconv.FormatInt(y3, 10)
	sdID := strconv.FormatInt(dID, 10)
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = db.Exec("CREATE EXTENSION IF NOT EXISTS tablefunc").Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Crédits par action, tablefunc : " + err.Error()})
		return
	}
	qry := `SELECT json_build_object('chapter', q.chapter, 'sector', q.sector, 'subfunction', q.subfunction,
	'program', q.program, 'action', q.action, 'action_name',q.action_name, 'number', q.number, 'name', q.name, 'y` +
		sy1 + `', q.y` + sy1 + `, 'y` + sy2 + `', q.y` + sy2 + `,'y` + sy3 + `',q.y` + sy3 + `) FROM
		(SELECT b.chapter, b.sector, b.subfunction, b.program, b.action, b.action_name, cq_union.number, cq_union.name, cq_union.y` + sy1 + `, cq_union.y` + sy2 + `,cq_union.y` + sy3 + ` FROM
		(
		 (SELECT op.budget_action_id AS action_id, op.number, op.name, SUM(ct.y` + sy1 + `) * 0.01 AS y` + sy1 + `, SUM(ct.y` + sy2 + `) * 0.01 AS y` + sy2 + `, SUM(ct.y` + sy3 + `) * 0.01 AS y` + sy3 + ` FROM
				crosstab(
						'WITH pr AS (SELECT * FROM payment_ratios WHERE payment_types_id = ` + sdID + `),
								pp AS (SELECT physical_op_id, year, value FROM prev_payment WHERE value IS NOT NULL AND value <> 0 AND year>= ` + sy1 + ` AND year <=` + sy3 + `),
								pp_idx AS (SELECT physical_op_id, year FROM pp),
								fc_sum AS (SELECT physical_op_id, EXTRACT(year FROM date)::integer AS year, SUM(value) AS value FROM financial_commitment WHERE EXTRACT(year FROM date) <` + sy1 + ` - 1 GROUP BY 1,2),
								fc AS (SELECT fc_sum.physical_op_id, fc_sum.year + pr.index AS year, fc_sum.value * pr.ratio AS value FROM fc_sum, pr WHERE fc_sum.year + pr.index >= ` + sy1 + ` AND fc_sum.year + pr.index <= ` + sy3 + `),
								fc_filtered AS (SELECT * FROM fc WHERE fc.physical_op_id IS NOT NULL AND (fc.physical_op_id, fc.year) NOT IN (SELECT * FROM pp_idx)),
								pg_year AS (SELECT physical_op_id, year, SUM(value) AS value FROM programmings WHERE year = ` + sy1 + ` - 1 GROUP BY 1,2),
								pg AS (SELECT pg_year.physical_op_id, pg_year.year + pr.index AS year, pg_year.value * pr.ratio AS value FROM pg_year, pr WHERE pg_year.year + pr.index >= ` + sy1 + ` AND pg_year.year + pr.index <= ` + sy3 + `),
								pg_filtered AS (SELECT * FROM pg WHERE (pg.physical_op_id, pg.year) NOT IN (SELECT * FROM pp_idx)),
								pc AS (SELECT p.physical_op_id, p.year + pr.index AS year, p.value * pr.ratio AS value FROM prev_commitment p, pr WHERE p.year + pr.index >= ` + sy1 + ` AND p.year + pr.index <= ` + sy3 + `),
								pc_filtered AS (SELECT * FROM pc WHERE (pc.physical_op_id, pc.year) NOT IN (SELECT * FROM pp_idx))
						SELECT * FROM
						(SELECT * FROM pp
						UNION ALL
						SELECT physical_op_id, year, SUM(value) AS value FROM 
								(SELECT * FROM fc_filtered UNION ALL SELECT * FROM pg_filtered UNION ALL SELECT * FROM pc_filtered)q1 
								GROUP BY 1,2) q2 ORDER BY 1,2',
						'SELECT m FROM generate_series(` + sy1 + `, ` + sy3 + `) AS m') 
						AS ct(physical_op_id integer, y` + sy1 + ` numeric, y` + sy2 + ` numeric, y` + sy3 + ` numeric)
				LEFT JOIN physical_op op ON op.id = ct.physical_op_id
		 GROUP BY 1,2,3
		 )
		 UNION ALL
		 (SELECT action_id, NULL AS number, NULL AS name, SUM(y` + sy1 + `) * 0.01 AS y` + sy1 + `, SUM(y` + sy2 + `) * 0.01 AS y` + sy2 + `, SUM(y` + sy3 + `) * 0.01 AS y` + sy3 + ` FROM
				crosstab('
						WITH pr AS (SELECT * FROM payment_ratios WHERE payment_types_id = ` + sdID + `),
								unlinked_fc_sum AS (SELECT action_id, EXTRACT(year FROM date)::integer AS year, SUM(value) AS value FROM financial_commitment WHERE EXTRACT(year FROM date) <` + sy1 + ` - 1 AND physical_op_id IS NULL GROUP BY 1,2),
								unlinked_fc AS (SELECT unlinked_fc_sum.action_id, unlinked_fc_sum.year + pr.index AS year, unlinked_fc_sum.value * pr.ratio AS value FROM unlinked_fc_sum, pr WHERE unlinked_fc_sum.year + pr.index >= ` + sy1 + ` AND unlinked_fc_sum.year + pr.index <= ` + sy3 + `)
						SELECT action_id, year, SUM(value) FROM unlinked_fc GROUP BY 1,2 ORDER BY 1,2',
						'SELECT m FROM generate_series(` + sy1 + `, ` + sy3 + `) AS m')
				AS (action_id integer, y` + sy1 + ` numeric, y` + sy2 + ` numeric, y` + sy3 + ` numeric)
		 GROUP BY 1, 2, 3)
		) cq_union
		LEFT JOIN (SELECT ba.id, bc.code AS chapter, bs.code AS sector, bp.code_function || COALESCE(bp.code_subfunction, '') AS subfunction,
				bp.code_contract || bp.code_function || bp.code_number as program,
				bp.code_contract || bp.code_function || bp.code_number || ba.code as action, ba.name AS action_name FROM 
				budget_chapter bc, budget_program bp, budget_action ba, budget_sector bs
				WHERE ba.program_id = bp.id AND bp.chapter_id = bc.id AND ba.sector_id = bs.id) b
		ON cq_union.action_id = b.id
		ORDER BY 1,2,3,4,5,6,7,8) q`
	rows, err := db.DB().Query(qry)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiement détaillé par action, requête : " + err.Error()})
		return
	}
	defer rows.Close()
	lines, line := []string{}, ""
	for rows.Next() {
		if err = rows.Scan(&line); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Paiement détaillé par action, lecture de ligne : " + err.Error()})
			return
		}
		lines = append(lines, line)
	}
	resp := `{"DetailedPaymentPerBudgetAction":[` + strings.Join(lines, ",") + `]}`

	ctx.StatusCode(http.StatusOK)
	ctx.ContentType("application/json")
	ctx.Write([]byte(resp))
}

// GetActionPayment handles the get request to get payment prevision by budget action.
func GetActionPayment(ctx iris.Context) {
	y1, err := ctx.URLParamInt64("FirstYear")
	if err != nil {
		y1 = int64(time.Now().Year()) + 1
	}
	dID, err := ctx.URLParamInt64("DefaultPaymentTypeId")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiement détaillé par action, décodage : " + err.Error()})
		return
	}
	y2 := y1 + 1
	y3 := y1 + 2
	sy1 := strconv.FormatInt(y1, 10)
	sy2 := strconv.FormatInt(y2, 10)
	sy3 := strconv.FormatInt(y3, 10)
	sdID := strconv.FormatInt(dID, 10)
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = db.Exec("CREATE EXTENSION IF NOT EXISTS tablefunc").Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Crédits par action, tablefunc : " + err.Error()})
		return
	}
	qry := `SELECT json_build_object('chapter', q.chapter, 'sector', q.sector, 'subfunction', q.subfunction,
	'program', q.program, 'action', q.action, 'action_name',q.action_name, 'y` +
		sy1 + `', q.y` + sy1 + `, 'y` + sy2 + `', q.y` + sy2 + `,'y` + sy3 + `',q.y` + sy3 + `) FROM
		(SELECT b.chapter, b.sector, b.subfunction, b.program, b.action, b.action_name, 
			SUM(y` + sy1 + `) * 0.01 AS y` + sy1 + `, SUM(y` + sy2 + `) * 0.01 AS y` + sy2 + `, SUM(y` + sy3 + `) * 0.01 AS y` + sy3 + ` FROM
(
(SELECT op.budget_action_id AS action_id, SUM(ct.y` + sy1 + `) AS y` + sy1 + `, SUM(ct.y` + sy2 + `) AS y` + sy2 + `, SUM(ct.y` + sy3 + `) AS y` + sy3 + ` FROM
crosstab(
'WITH pr AS (SELECT * FROM payment_ratios WHERE payment_types_id = ` + sdID + `),
	pp AS (SELECT physical_op_id, year, value FROM prev_payment WHERE value IS NOT NULL AND value <> 0 AND year>= ` + sy1 + ` AND year <=` + sy3 + `),
	pp_idx AS (SELECT physical_op_id, year FROM pp),
	fc_sum AS (SELECT physical_op_id, EXTRACT(year FROM date)::integer AS year, SUM(value) AS value FROM financial_commitment WHERE EXTRACT(year FROM date) <` + sy1 + ` - 1 GROUP BY 1,2),
	fc AS (SELECT fc_sum.physical_op_id, fc_sum.year + pr.index AS year, fc_sum.value * pr.ratio AS value FROM fc_sum, pr WHERE fc_sum.year + pr.index >= ` + sy1 + ` AND fc_sum.year + pr.index <= ` + sy3 + `),
	fc_filtered AS (SELECT * FROM fc WHERE fc.physical_op_id IS NOT NULL AND (fc.physical_op_id, fc.year) NOT IN (SELECT * FROM pp_idx)),
	pg_year AS (SELECT physical_op_id, year, SUM(value) AS value FROM programmings WHERE year = ` + sy1 + ` - 1 GROUP BY 1,2),
	pg AS (SELECT pg_year.physical_op_id, pg_year.year + pr.index AS year, pg_year.value * pr.ratio AS value FROM pg_year, pr WHERE pg_year.year + pr.index >= ` + sy1 + ` AND pg_year.year + pr.index <= ` + sy3 + `),
	pg_filtered AS (SELECT * FROM pg WHERE (pg.physical_op_id, pg.year) NOT IN (SELECT * FROM pp_idx)),
	pc AS (SELECT p.physical_op_id, p.year + pr.index AS year, p.value * pr.ratio AS value FROM prev_commitment p, pr WHERE p.year + pr.index >= ` + sy1 + ` AND p.year + pr.index <= ` + sy3 + `),
	pc_filtered AS (SELECT * FROM pc WHERE (pc.physical_op_id, pc.year) NOT IN (SELECT * FROM pp_idx))
SELECT * FROM
(SELECT * FROM pp
UNION ALL
SELECT physical_op_id, year, SUM(value) AS value FROM 
	(SELECT * FROM fc_filtered UNION ALL SELECT * FROM pg_filtered UNION ALL SELECT * FROM pc_filtered)q1 
	GROUP BY 1,2) q2 ORDER BY 1,2',
'SELECT m FROM generate_series(` + sy1 + `, ` + sy3 + `) AS m') 
AS ct(physical_op_id integer, y` + sy1 + ` numeric, y` + sy2 + ` numeric, y` + sy3 + ` numeric)
LEFT JOIN physical_op op ON op.id = ct.physical_op_id
GROUP BY 1
)
UNION ALL
(SELECT action_id, SUM(y` + sy1 + `) * 0.01 AS y` + sy1 + `, SUM(y` + sy2 + `) * 0.01 AS y` + sy2 + `, SUM(y` + sy3 + `) * 0.01 AS y` + sy3 + ` FROM
crosstab('
WITH pr AS (SELECT * FROM payment_ratios WHERE payment_types_id = ` + sdID + `),
	unlinked_fc_sum AS (SELECT action_id, EXTRACT(year FROM date)::integer AS year, SUM(value) AS value FROM financial_commitment WHERE EXTRACT(year FROM date) <` + sy1 + ` - 1 AND physical_op_id IS NULL GROUP BY 1,2),
	unlinked_fc AS (SELECT unlinked_fc_sum.action_id, unlinked_fc_sum.year + pr.index AS year, unlinked_fc_sum.value * pr.ratio AS value FROM unlinked_fc_sum, pr WHERE unlinked_fc_sum.year + pr.index >= ` + sy1 + ` AND unlinked_fc_sum.year + pr.index <= ` + sy3 + `)
SELECT action_id, year, SUM(value) FROM unlinked_fc GROUP BY 1,2 ORDER BY 1,2',
'SELECT m FROM generate_series(` + sy1 + `, ` + sy3 + `) AS m')
AS (action_id integer, y` + sy1 + ` numeric, y` + sy2 + ` numeric, y` + sy3 + ` numeric)
GROUP BY 1)
) cq_union
LEFT JOIN (SELECT ba.id, bc.code AS chapter, bs.code AS sector, bp.code_function || COALESCE(bp.code_subfunction, '') AS subfunction,
bp.code_contract || bp.code_function || bp.code_number as program,
bp.code_contract || bp.code_function || bp.code_number || ba.code as action, ba.name AS action_name FROM 
budget_chapter bc, budget_program bp, budget_action ba, budget_sector bs
WHERE ba.program_id = bp.id AND bp.chapter_id = bc.id AND ba.sector_id = bs.id) b
ON cq_union.action_id = b.id
GROUP BY 1,2,3,4,5,6
ORDER BY 1,2,3,4,5,6) q`
	rows, err := db.DB().Query(qry)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiement par action, requête : " + err.Error()})
		return
	}
	defer rows.Close()
	lines, line := []string{}, ""
	for rows.Next() {
		if err = rows.Scan(&line); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Paiement par action, lecture de ligne : " + err.Error()})
			return
		}
		lines = append(lines, line)
	}
	resp := `{"PaymentPerBudgetAction":[` + strings.Join(lines, ",") + `]}`

	ctx.StatusCode(http.StatusOK)
	ctx.ContentType("application/json")
	ctx.Write([]byte(resp))
}
