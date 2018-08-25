package actions

import (
	"net/http"
	"time"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// prog is used to scan the programmation query results.
type prog struct {
	ID                  models.NullInt64   `json:"id"`
	Value               models.NullInt64   `json:"value"`
	TotalValue          models.NullInt64   `json:"total_value"`
	StateRatio          models.NullFloat64 `json:"state_ratio"`
	PhysicalOpID        int64              `json:"physical_op_id"`
	CommissionID        models.NullInt64   `json:"commission_id"`
	OpNumber            string             `json:"op_number"`
	OpName              string             `json:"op_name"`
	Prevision           models.NullInt64   `json:"prevision"`
	TotalPrevision      models.NullInt64   `json:"total_prevision"`
	StateRatioPrevision models.NullFloat64 `json:"state_ratio_prevision"`
	PreProgValue        models.NullInt64   `json:"pre_prog_value"`
	PreProgTotalValue   models.NullInt64   `json:"pre_prog_total_value"`
	PreProgStateRatio   models.NullFloat64 `json:"pre_prog_state_ratio"`
	PreProgDescript     models.NullString  `json:"pre_prog_descript"`
	PlanName            models.NullString  `json:"plan_name"`
	PlanLineName        models.NullString  `json:"plan_line_name"`
	PlanLineValue       models.NullInt64   `json:"plan_line_value"`
	PlanLineTotalValue  models.NullInt64   `json:"plan_line_total_value"`
}

// progResp is used to embed an array of prog.
type progResp struct {
	Programmings []prog `json:"Programmings"`
}

// getProg fetches the datas for the user and year and embeddes it in a preProgResp
func getProg(ctx iris.Context, errPrefix string, year int64) (*progResp, error) {
	db := ctx.Values().Get("db").(*gorm.DB)
	rows, err := db.Raw(`SELECT pr.id, pr.value, pr.total_value, pr.state_ratio, op.id AS physical_op_id, 
			pr.commission_id, op.number as op_number, op.name as op_name, pc.value as prevision, 
			pc.total_value as total_prevision, pc.state_ratio as state_ratio_prevision,
			pp.value AS pre_prog_value, pp.total_value AS pre_prog_total_value,
			pp.state_ratio AS pre_prog_state_ratio, pp.descript AS pre_prog_descript, pl.plan_name, 
			pl.plan_line_name, pl.plan_line_value, pl.plan_line_total_value
		FROM physical_op op
		LEFT OUTER JOIN (SELECT pl.id, pl.name AS plan_line_name, pl.value AS plan_line_value, 
				pl.total_value AS plan_line_total_value, p.name AS plan_name 
			FROM plan_line pl, plan p WHERE pl.plan_id = p.id) pl ON op.plan_line_id = pl.id
		LEFT OUTER JOIN (SELECT * FROM programmings WHERE year=?) pr ON pr.physical_op_id = op.id
		LEFT OUTER JOIN (SELECT * FROM prev_commitment WHERE year=?) pc ON pc.physical_op_id = op.id
		LEFT OUTER JOIN (SELECT * FROM pre_programmings WHERE year=?) pp ON op.id = pp.physical_op_id`,
		year, year, year).Rows()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{errPrefix + ", erreur de requête : " + err.Error()})
		return nil, err
	}
	defer rows.Close()
	pp, p := progResp{}, prog{}
	for rows.Next() {
		db.ScanRows(rows, &p)
		pp.Programmings = append(pp.Programmings, p)
	}
	db.Commit()
	return &pp, nil
}

// GetProgrammings handle the get request to fetch the programming of a year.
func GetProgrammings(ctx iris.Context) {
	year, err := ctx.URLParamInt64("year")
	if err != nil {
		year = int64(time.Now().Year())
	}
	resp, err := getProg(ctx, "Programmation annuelle", year)
	if err != nil {
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

//progYearResp embeddes the array of years for the response
type progYearResp struct {
	ProgrammingsYear []int64 `json:"ProgrammingsYear"`
}

// GetProgrammingsYear handles the get request to get years with available programmation
func GetProgrammingsYear(ctx iris.Context) {
	resp, db := progYearResp{}, ctx.Values().Get("db").(*gorm.DB)
	rows, err := db.Raw("select distinct year from programmings").Rows()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Années de programmation, select : " + err.Error()})
		return
	}
	defer rows.Close()
	var year int64
	for rows.Next() {
		rows.Scan(&year)
		resp.ProgrammingsYear = append(resp.ProgrammingsYear, year)
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// progReq is used to decode sent programming.
type progReq struct {
	Value        int64              `json:"value"`
	PhysicalOpID int64              `json:"physical_op_id"`
	CommissionID int64              `json:"commission_id"`
	Year         int64              `json:"year"`
	TotalValue   models.NullInt64   `json:"total_value"`
	StateRatio   models.NullFloat64 `json:"state_ratio"`
}

// batchProgReq embeddes the array of progReq.
type batchProgReq struct {
	Programmings []progReq `json:"Programmings"`
	Year         int64     `json:"year"`
}

// BatchProgrammings handles the post request containing a full programmation for the current year.
func BatchProgrammings(ctx iris.Context) {
	req := batchProgReq{}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch programmation, décodage impossible : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	tx, err := db.DB().Begin()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch programmation, erreur transaction  : " + err.Error()})
		return
	}
	if _, err := tx.Exec("DELETE from programmings WHERE year = $1", req.Year); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch programmation, erreur de delete : " + err.Error()})
		tx.Rollback()
		return
	}
	stmt, err := tx.Prepare(`INSERT INTO programmings (value, physical_op_id, commission_id, year, 
		total_value, state_ratio) VALUES ($1,$2,$3,$4,$5,$6)`)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch programmation, erreur de statement : " + err.Error()})
		tx.Rollback()
		return
	}
	for _, p := range req.Programmings {
		if _, err := stmt.Exec(p.Value, p.PhysicalOpID, p.CommissionID, p.Year, p.TotalValue, p.StateRatio); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Batch programmation, erreur de insert : " + err.Error()})
			tx.Rollback()
			return
		}
	}
	tx.Commit()
	resp, err := getProg(ctx, "Batch programmation", req.Year)
	if err != nil {
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(*resp)
}
