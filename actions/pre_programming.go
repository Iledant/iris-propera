package actions

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/Iledant/iris_propera/actions/queries"
	"github.com/jinzhu/gorm"

	"github.com/Iledant/iris_propera/models"
	"github.com/kataras/iris"
)

// preProg is used to scan the select pre programming query results
type preProg struct {
	PhysicalOpID        int64              `json:"physical_op_id"`
	PhysicalOpNumber    string             `json:"physical_op_number"`
	PhysicalOpName      string             `json:"physical_op_name"`
	PrevValue           models.NullInt64   `json:"prev_value"`
	PrevStateRatio      models.NullFloat64 `json:"prev_state_ratio"`
	PrevTotalValue      models.NullInt64   `json:"prev_total_value"`
	PrevDescript        models.NullString  `json:"prev_descript"`
	PreProgID           models.NullInt64   `json:"pre_prog_id"`
	PreProgValue        models.NullInt64   `json:"pre_prog_value"`
	PreProgYear         models.NullInt64   `json:"pre_prog_year"`
	PreProgCommissionID models.NullInt64   `json:"pre_prog_commission_id"`
	PreProgStateRatio   models.NullFloat64 `json:"pre_prog_state_ratio"`
	PreProgTotalValue   models.NullInt64   `json:"pre_prog_total_value"`
	PreProgDescript     models.NullString  `json:"pre_prog_descript"`
	PlanName            models.NullString  `json:"plan_name"`
	PlanLineName        models.NullString  `json:"plan_line_name"`
	PlanLineValue       models.NullInt64   `json:"plan_line_value"`
	PlanLineTotalValue  models.NullInt64   `json:"plan_line_total_value"`
}

// preProgResp embeddes the Preprogrammings response
type preProgResp struct {
	Pp []preProg `json:"PreProgrammings"`
}

// getUserRoleAndID fetch user role and ID with the token
func getUserRoleAndID(ctx iris.Context, errPrefix string, role *string, userID *int) error {
	u, err := bearerToUser(ctx)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{errPrefix + ", récupération token : " + err.Error()})
		return err
	}
	uID, err := strconv.Atoi(u.Subject)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{errPrefix + ", lecture ID user : " + err.Error()})
		return err
	}
	*role = u.Role
	*userID = uID
	return nil
}

// getPreProg fetches the datas for the user and year and embeddes it in a preProgResp
func getPreProg(ctx iris.Context, errPrefix string, role string, userID int, year int64) (*preProgResp, error) {
	db := ctx.Values().Get("db").(*gorm.DB)
	var rows *sql.Rows
	var err error
	tx := db.Begin()
	if role == models.AdminRole {
		rows, err = tx.Raw(queries.GetAdminPreProg, year, year).Rows()
	} else {
		rows, err = tx.Raw(queries.GetUserPreProg, userID, year, year).Rows()
	}
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{errPrefix + ", erreur de requête : " + err.Error()})
		tx.Rollback()
		return nil, err
	}
	defer rows.Close()
	pp, p := preProgResp{}, preProg{}
	for rows.Next() {
		tx.ScanRows(rows, &p)
		pp.Pp = append(pp.Pp, p)
	}
	tx.Commit()
	return &pp, nil
}

// GetPreProgrammings handle the get request to get all preprogrammings and all linked datas.
// The scope is all physical operations for ADMIN role or controlled operations for USER
func GetPreProgrammings(ctx iris.Context) {
	year, err := ctx.URLParamInt64("year")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste de la préprogrammation : " + err.Error()})
		return
	}

	role, userID := "", 0
	if err = getUserRoleAndID(ctx, "Liste de la préprogrammation", &role, &userID); err != nil {
		return
	}

	preProgResp, err := getPreProg(ctx, "Liste de la préprogrammation", role, userID, year)
	if err != nil {
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(*preProgResp)
}

// sentPreProg is used to parse one row of sent datas for pre programmings
type sentPreProg struct {
	PhysicalOpID        *int64   `json:"physical_op_id"`
	PreProgID           *int64   `json:"pre_prog_id"`
	PreProgYear         *int64   `json:"pre_prog_year"`
	PreProgValue        *int64   `json:"pre_prog_value"`
	PreProgCommissionID *int64   `json:"pre_prog_commission_id"`
	PreProgTotalValue   *int64   `json:"pre_prog_total_value"`
	PreProgStateRatio   *float64 `json:"pre_prog_state_ratio"`
}

// sentReq is used to decode the global data set
type preProgReq struct {
	Pps  []sentPreProg `json:"PreProgrammings"`
	Year int64         `json:"year"`
}

// BatchPreProgrammings sets the pre programmings replacing existing one
func BatchPreProgrammings(ctx iris.Context) {
	var req preProgReq
	err := ctx.ReadJSON(&req)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch préprogrammation, erreur de décodage : " + err.Error()})
		return
	}

	role, userID := "", 0
	if err = getUserRoleAndID(ctx, "Liste de la préprogrammation", &role, &userID); err != nil {
		return
	}

	db := ctx.Values().Get("db").(*gorm.DB)
	tx := db.Begin()

	if err = tx.Exec(queries.CreateTempPreProgTable).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch préprogrammation, erreur création table temporaire : " + err.Error()})
		tx.Rollback()
		return
	}

	for _, pp := range req.Pps {
		if err = tx.Exec(queries.InsertTempPreProg, pp.PreProgID, pp.PreProgYear, pp.PhysicalOpID,
			pp.PreProgCommissionID, pp.PreProgValue, pp.PreProgTotalValue, pp.PreProgStateRatio).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Batch préprogrammation, erreur insertion table temporaire : " + err.Error()})
			tx.Rollback()
			return
		}
	}

	if err = tx.Exec(queries.UpdatePreProgWithTemp).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch préprogrammation, erreur update : " + err.Error()})
		tx.Rollback()
		return
	}

	if role == models.AdminRole {
		if err = tx.Exec(queries.DelPreProgAdmin, req.Year).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Batch préprogrammation, erreur delete : " + err.Error()})
			tx.Rollback()
			return
		}
	} else {
		if err = tx.Exec(queries.DelPreProgUser, userID, req.Year).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Batch préprogrammation, erreur delete : " + err.Error()})
			tx.Rollback()
			return
		}
	}

	if err = tx.Exec(queries.InsertPreProg).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch préprogrammation, erreur insert : " + err.Error()})
		tx.Rollback()
		return
	}

	if err = tx.Exec(queries.DeleteTempPreProgTable).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch préprogrammation, erreur suppression table temporaire : " + err.Error()})
		tx.Rollback()
		return
	}

	tx.Commit()

	preProgResp, err := getPreProg(ctx, "Batch préprogrammation", role, userID, req.Year)
	if err != nil {
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(*preProgResp)
}
