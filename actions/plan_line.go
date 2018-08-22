package actions

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/iris"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
)

type planLineResp struct {
	PlanLine []string `json:"PlanLine"`
}

func getFirstAndLastYear(plan models.Plan, db *gorm.DB) (firstYear int64, lastYear int64, err error) {
	firstYear = int64(time.Now().Year() + 1)
	if plan.FirstYear.Valid && plan.LastYear.Valid {
		if plan.FirstYear.Int64 > firstYear {
			firstYear = plan.FirstYear.Int64
		}
		lastYear = plan.LastYear.Int64
	} else {
		if err := db.DB().QueryRow("select max(year) from prev_commitment").Scan(&lastYear); err != nil {
			return 0, 0, err
		}
	}
	return firstYear, lastYear, nil
}

// getPlanLineAndPrevisons compute the query to get all informations including previsions of a plan line or all plan lines of a plan
// As the number of columns can't be known the query converts everything directly in JSON within postsgresql
func getPlanLineAndPrevisons(plan models.Plan, planLineID int64, db *gorm.DB) (*string, error) {
	firstYear, lastYear, err := getFirstAndLastYear(plan, db)
	if err != nil {
		return nil, err
	}

	var whereQry, beneficiaryIdsQry string
	if planLineID != 0 {
		whereQry = "p.id=" + strconv.FormatInt(planLineID, 10) + " "
		beneficiaryIdsQry = "SELECT DISTINCT beneficiary_id FROM plan_line_ratios WHERE plan_line_id = " +
			strconv.FormatInt(planLineID, 10) + " "
	} else {
		whereQry = "p.plan_id=" + strconv.FormatInt(plan.ID, 10) + " "
		beneficiaryIdsQry = "SELECT DISTINCT beneficiary_id FROM plan_line_ratios WHERE plan_line_id IN (SELECT id FROM plan_line WHERE plan_id=" +
			strconv.FormatInt(plan.ID, 10) + ") "
	}

	bIDs, bID := []int64{}, int64(0)
	rows, err := db.DB().Query(beneficiaryIdsQry)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&bID); err != nil {
			return nil, err
		}
		bIDs = append(bIDs, bID)
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
		benCrossQry = " LEFT JOIN (SELECT * FROM crosstab('SELECT plan_line_id, beneficiary_id, ratio FROM plan_line_ratios ORDER BY 1,2', '" +
			beneficiaryIdsQry + "') AS (plan_line_id INTEGER, " + benCrossQry + ")) b ON b.plan_line_id = p.id"
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
	finalQry := `SELECT json_build_object('id',q.id,'name',q.name, 'descript', q.descript,'value', q.value, 
	'total_value', q.total_value, 'commitment', q.commitment, 'programmings', q.programmings,` + jsonQry + jsonBenQry + ` ) FROM
	(SELECT p.id, p.name, p.descript, p.value, p.total_value, 
	CAST(fc.value AS bigint) AS commitment, CAST(pr.value AS bigint) AS programmings, ` + prevQry + benQry + `
FROM plan_line p` + benCrossQry + `
LEFT JOIN (SELECT f.plan_line_id, SUM(f.value) AS value FROM financial_commitment f
						WHERE f.plan_line_id NOTNULL AND EXTRACT(year FROM f.date) < ` + actualYear + `
						GROUP BY 1) fc
	ON fc.plan_line_id = p.id
LEFT JOIN (SELECT op.plan_line_id, SUM(p.value) AS value FROM physical_op op, programmings p 
						WHERE p.physical_op_id = op.id AND p.year = ` + actualYear + ` GROUP BY 1) pr 
	ON pr.plan_line_id = p.id
LEFT JOIN (SELECT * FROM 
	crosstab ('SELECT p.plan_line_id, c.year, SUM(c.value) FROM physical_op p, prev_commitment c 
							WHERE p.id = c.physical_op_id AND p.plan_line_id NOTNULL GROUP BY 1,2 ORDER BY 1,2',
						'SELECT m FROM generate_series( ` + strconv.FormatInt(firstYear, 10) + `, ` + strconv.FormatInt(lastYear, 10) + `) AS m')
		AS (plan_line_id INTEGER, ` + convertQry + `)) prev 
ON prev.plan_line_id = p.id
WHERE ` + whereQry + ` ORDER BY 1) q`

	lines, line := []string{}, ""
	rows, err = db.DB().Query(finalQry)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&line); err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}

	resp := strings.Join(lines, ",")
	if planLineID == 0 {
		resp = `[` + resp + `]`
	}
	return &resp, nil
}

// getPlan fetch params to get the ID of the plan and queries the database
func getPlan(ctx iris.Context, db *gorm.DB, errPrefix string) (plan models.Plan, err error) {
	planID, err := ctx.Params().GetInt64("pID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{errPrefix + ", identificateur du plan : " + err.Error()})
		return
	}

	if err := db.First(&plan, planID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{errPrefix + " : plan introuvable"})
			return plan, err
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{errPrefix + ", requête plan : " + err.Error()})
		return plan, err
	}

	return plan, nil
}

// GetPlanLines handles the get request to have all plan lines of a plan.
func GetPlanLines(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	plan, err := getPlan(ctx, db, "Liste des lignes de plan")
	if err != nil {
		return
	}

	resp, err := getPlanLineAndPrevisons(plan, 0, db)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des lignes de plan, requête de calcul : " + err.Error()})
		return
	}

	bb, b := []models.Beneficiary{}, models.Beneficiary{}
	rows, err := db.Raw(`SELECT * FROM beneficiary WHERE id IN 
	(SELECT DISTINCT beneficiary_id FROM plan_line_ratios WHERE plan_line_id IN 
		(SELECT id FROM plan_line WHERE plan_id=?))`, plan.ID).Rows()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des lignes de plan, récupération des bénéficiaires : " + err.Error()})
		return
	}
	defer rows.Close()
	for rows.Next() {
		if err = db.ScanRows(rows, &b); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Liste des lignes de plan, lecture des bénéficiaires : " + err.Error()})
			return
		}
		bb = append(bb, b)
	}

	jsonBb, err := json.Marshal(bb)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des lignes de plan, json des bénéficiaires : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.ContentType("application/json")
	r := append([]byte(`{"PlanLine":`+*resp+`,"Beneficiary":`), jsonBb...)
	r = append(r, '}')
	ctx.Write(r)
}

// GetDetailedPlanLines handles the get request to have all operation prevision by lines.
func GetDetailedPlanLines(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	plan, err := getPlan(ctx, db, "Liste détaillée des lignes de plan")
	if err != nil {
		return
	}

	firstYear, lastYear, err := getFirstAndLastYear(plan, db)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste détaillée des lignes de plan impossible de calculer les années : " + err.Error()})
		return
	}

	pp, nn, cc, ll, jj := []string{}, []string{}, []string{}, []string{}, []string{}
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

	finalQry := `SELECT json_build_object('id',q.id,'name',q.name,'commitment_name', q.commitment_name, 
	'commitment_code', q.commitment_code, 'commitment_date', q.commitment_date, 'commitment_value', q.commitment_value,
	'programmings_value', q.programmings_value, 'programmings_date', q.programmings_date,	` + jsonQry + `) FROM
	(SELECT pl.id, pl.name, pl.total_value, pl.value, fc.op_number, fc.op_name,
	fc.commitment_name, fc.commitment_code, fc.commitment_date, fc.commitment_value,   
	fc.programmings_value, fc.programmings_date, ` + prevQry + ` FROM plan_line pl
LEFT OUTER JOIN 
(SELECT op.number AS op_number, op.name as op_name, f.name AS commitment_name, f.iris_code AS commitment_code, 
			f.date AS commitment_date, f.value AS commitment_value, NULL AS programmings_value,NULL AS programmings_date,
	` + nullQry + `, f.plan_line_id 
FROM financial_commitment f, physical_op op 
WHERE EXTRACT(year FROM f.date) < ` + actualYear + ` AND f.plan_line_id NOTNULL AND f.physical_op_id = op.id
UNION ALL
SELECT op.number AS op_number, op.name as op_name, NULL AS commitment_name, NULL AS commitment_code, NULL AS commitment_date,
			NULL AS commitment_value, p.value AS programmings_value, c.date AS programmings_date, 
	` + nullQry + `, op.plan_line_id
FROM programmings p, physical_op op,commissions c 
WHERE p.year = ` + actualYear + ` AND c.id = p.commission_id AND op.id = p.physical_op_id
UNION ALL
SELECT op.number AS op_number, op.name as op_name, NULL AS commitment_name, NULL AS commitment_code, NULL AS commitment_date,
			NULL AS commitment_value, NULL AS programmings_value, NULL AS programmings_date, ` + colQry + `, op.plan_line_id
FROM crosstab ('SELECT physical_op_id, year, value FROM prev_commitment ORDER BY 1,2',
			'SELECT m FROM generate_series(` + strconv.FormatInt(firstYear, 10) + `, ` + strconv.FormatInt(lastYear, 10) + `) AS m')
	AS (physical_op_id INTEGER, ` + convertQry + `) , physical_op op
WHERE physical_op_id = op.id 
) fc ON fc.plan_line_id = pl.id
WHERE pl.plan_id = ` + strconv.FormatInt(plan.ID, 10) + `
ORDER BY 1,5,9,12) q`

	lines, line := []string{}, ""
	rows, err := db.DB().Query(finalQry)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste détaillée des lignes de plan requête finale : " + err.Error()})
		return
	}

	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&line); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Liste détaillée des lignes de plan lecture des lignes : " + err.Error()})
			return
		}
		lines = append(lines, line)
	}

	resp := `{"DetailedPlanLine":[` + strings.Join(lines, ",") + `]}`

	ctx.StatusCode(http.StatusOK)
	ctx.ContentType("application/json")
	ctx.Write([]byte(resp))
}

// ratioReq is used to decode a plan line ratio
type ratioReq struct {
	Ratio         float64 `json:"ratio"`
	PlanLineID    int64   `json:"plan_line_id"`
	BeneficiaryID int64   `json:"beneficiary_id"`
}

// planLineReq is used to decode a plan line with embedded ratios array
type planLineReq struct {
	Name       *string     `json:"name"`
	Value      *int64      `json:"value"`
	TotalValue *int64      `json:"total_value"`
	Descript   *string     `json:"descript"`
	Ratios     *[]ratioReq `json:"ratios"`
}

// setPlanLineRatios destroy existing ratios and add sent ones for a plan line
func setPlanLineRatios(tx *gorm.DB, plID int64, ratios *[]ratioReq) error {
	if ratios != nil {
		if err := tx.Exec("DELETE FROM plan_line_ratios WHERE plan_line_id = ?", plID).Error; err != nil {
			return err
		}
		for _, r := range *ratios {
			if err := tx.Exec("INSERT INTO plan_line_ratios (plan_line_id, beneficiary_id, ratio) VALUES (?,?,?)",
				plID, r.BeneficiaryID, r.Ratio).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// CreatePlanLine handles the post request to create a plan line
func CreatePlanLine(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	plan, err := getPlan(ctx, db, "Création de ligne de plan")
	if err != nil {
		return
	}

	req := planLineReq{}
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de ligne de plan impossible de décoder : " + err.Error()})
		return
	}

	if req.Name == nil || *req.Name == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création de ligne de plan, erreur de name"})
		return
	}

	if req.Value == nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création de ligne de plan, erreur de value"})
		return
	}
	tx := db.Begin()

	planLine := models.PlanLine{Name: *req.Name, Value: *req.Value, PlanID: plan.ID}
	if req.TotalValue != nil {
		planLine.TotalValue.Valid = true
		planLine.TotalValue.Int64 = *req.TotalValue
	} else {
		planLine.TotalValue.Valid = false
	}
	if req.Descript != nil {
		planLine.Descript.Valid = true
		planLine.Descript.String = *req.Descript
	} else {
		planLine.Descript.Valid = false
	}

	if err = tx.Create(&planLine).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de ligne de plan, erreur d'insertion : " + err.Error()})
		tx.Rollback()
		return
	}

	if err = setPlanLineRatios(tx, planLine.ID, req.Ratios); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de ligne de plan, erreur sur les ratios : " + err.Error()})
		tx.Rollback()
		return
	}

	tx.Commit()

	pl, err := getPlanLineAndPrevisons(plan, planLine.ID, db)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de ligne de plan, impossible de récupérer : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.ContentType("application/json")
	ctx.Write([]byte(`{"PlanLine":` + *pl + `}`))
}

// ModifyPlanLine handle the put request to modify a plan line.
func ModifyPlanLine(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	plan, err := getPlan(ctx, db, "Modification de ligne de plan")
	if err != nil {
		return
	}

	plID, err := ctx.Params().GetInt64("plID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de ligne de plan, lecture de l'identifiant : " + err.Error()})
		return
	}

	planLine := models.PlanLine{}
	if err = db.First(&planLine, plID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Modification de ligne de plan : introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de ligne de plan, récupération de la ligne : " + err.Error()})
		return
	}

	req := planLineReq{}
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de ligne de plan impossible de décoder : " + err.Error()})
		return
	}

	if req.Name != nil && *req.Name != "" {
		planLine.Name = *req.Name
	}

	if req.Value != nil {
		planLine.Value = *req.Value
	}

	if req.TotalValue != nil {
		planLine.TotalValue.Int64 = *req.TotalValue
		planLine.TotalValue.Valid = true
	}

	if req.Descript != nil {
		planLine.Descript.String = *req.Descript
		planLine.TotalValue.Valid = true
	}

	tx := db.Begin()

	if err = tx.Save(&planLine).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de ligne de plan, erreur d'update' : " + err.Error()})
		tx.Rollback()
		return
	}

	if err = setPlanLineRatios(tx, planLine.ID, req.Ratios); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de ligne de plan, erreur sur les ratios : " + err.Error()})
		tx.Rollback()
		return
	}

	tx.Commit()

	pl, err := getPlanLineAndPrevisons(plan, planLine.ID, db)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de ligne de plan, impossible de récupérer : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.ContentType("application/json")
	ctx.Write([]byte(`{"PlanLine":` + *pl + `}`))
}

// DeletePlanLine handle the delete request to remove a plan line.
func DeletePlanLine(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)

	plID, err := ctx.Params().GetInt64("plID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression de ligne de plan, lecture de l'identifiant : " + err.Error()})
		return
	}

	planLine := models.PlanLine{}
	if err = db.First(&planLine, plID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Suppression de ligne de plan : introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression de ligne de plan, récupération de la ligne : " + err.Error()})
		return
	}

	tx := db.Begin()

	if err = tx.Exec("DELETE FROM plan_line_ratios WHERE plan_line_id = ?", plID).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression de ligne de plan, erreur de suppression des ratios : " + err.Error()})
		tx.Rollback()
		return
	}

	if err = tx.Exec("DELETE FROM plan_line WHERE id = ?", plID).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression de ligne de plan, erreur de delete : " + err.Error()})
		tx.Rollback()
		return
	}

	tx.Commit()
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Ligne de plan supprimée"})
}

// TODO : implement batch with a variable number of columns
