package actions

import (
	"database/sql"
	"net/http"

	"github.com/kataras/iris"

	"github.com/Iledant/iris-propera/models"
)

// TODO : refactor model

type planLineResp struct {
	PlanLine []string `json:"PlanLine"`
}

// PlanLinesResp embeddes plan line and previsions and beneficiaries
// for JSON export.
type PlanLinesResp struct {
	models.Beneficiaries
	models.PlanLineAndPrevisions
}

// GetPlanLines handles the get request to have all plan lines of a plan.
func GetPlanLines(ctx iris.Context) {
	db := ctx.Values().Get("db").(*sql.DB)
	planID, err := ctx.Params().GetInt64("pID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Listes des lignes de plan, paramètre : " + err.Error()})
		return
	}
	plan := models.Plan{ID: planID}
	if err = plan.GetByID(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Listes des lignes de plan, requête plan : " + err.Error()})
		return
	}
	var resp PlanLinesResp
	if err = resp.PlanLineAndPrevisions.GetAll(&plan, 0, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des lignes de plan, requête de calcul : " + err.Error()})
		return
	}
	if err = resp.Beneficiaries.GetPlanAll(planID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des lignes de plan, récupération des bénéficiaires : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetDetailedPlanLines handles the get request to have all operation prevision by lines.
func GetDetailedPlanLines(ctx iris.Context) {
	db := ctx.Values().Get("db").(*sql.DB)
	planID, err := ctx.Params().GetInt64("pID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste détaillée des lignes de plan, paramètre : " + err.Error()})
		return
	}
	plan := models.Plan{ID: planID}
	if err = plan.GetByID(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste détaillée des lignes de plan, requête plan : " + err.Error()})
		return
	}
	var resp models.DetailedPlanLineAndPrevisions
	if err = resp.GetAll(&plan, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste détaillée des lignes de plan, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// planLineReq is used to decode a plan line with embedded ratios array
type planLineReq struct {
	Name       *string `json:"name"`
	Value      *int64  `json:"value"`
	TotalValue *int64  `json:"total_value"`
	Descript   *string `json:"descript"`
	models.PlanLineRatios
}

// CreatePlanLine handles the post request to create a plan line
func CreatePlanLine(ctx iris.Context) {
	db := ctx.Values().Get("db").(*sql.DB)
	planID, err := ctx.Params().GetInt64("pID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de ligne de plan, paramètre : " + err.Error()})
		return
	}
	plan := models.Plan{ID: planID}
	if err = plan.GetByID(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de ligne de plan, requête plan : " + err.Error()})
		return
	}
	var req planLineReq
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de ligne de plan, décodage : " + err.Error()})
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
	if err = planLine.Create(&req.PlanLineRatios, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de ligne de plan, requête : " + err.Error()})
		return
	}
	var pl models.PlanLineAndPrevisions
	if err = pl.GetAll(&plan, planLine.ID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de ligne de plan, requête get : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(pl)
}

// ModifyPlanLine handle the put request to modify a plan line.
func ModifyPlanLine(ctx iris.Context) {
	db := ctx.Values().Get("db").(*sql.DB)
	planID, err := ctx.Params().GetInt64("pID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de ligne de plan, paramètre pID : " + err.Error()})
		return
	}
	plan := models.Plan{ID: planID}
	if err = plan.GetByID(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de ligne de plan, requête plan : " + err.Error()})
		return
	}
	plID, err := ctx.Params().GetInt64("plID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de ligne de plan, paramètre plID : " + err.Error()})
		return
	}
	planLine := models.PlanLine{ID: plID}
	if err = planLine.GetByID(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de ligne de plan, requête getByID : " + err.Error()})
		return
	}
	req := planLineReq{}
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de ligne de plan, décodage : " + err.Error()})
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
	if err = planLine.Update(&req.PlanLineRatios, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de ligne de plan, requête : " + err.Error()})
		return
	}
	var pl models.PlanLineAndPrevisions
	if err = pl.GetAll(&plan, planLine.ID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de ligne de plan, requête get : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(pl)
}

// DeletePlanLine handle the delete request to remove a plan line.
func DeletePlanLine(ctx iris.Context) {
	db := ctx.Values().Get("db").(*sql.DB)
	plID, err := ctx.Params().GetInt64("plID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression de ligne de plan, paramètre : " + err.Error()})
		return
	}
	planLine := models.PlanLine{ID: plID}
	if err = planLine.Delete(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression de ligne de plan, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Ligne de plan supprimée"})
}

// BatchPlanLines handle the post request to import a batch of plan lines
// and their beneficiary ratios
func BatchPlanLines(ctx iris.Context) {
	pID, err := ctx.Params().GetInt64("pID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch lignes de plan, paramètre : " + err.Error()})
		return
	}
	var req models.PlanLineBatch
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Batch lignes de plan, décodage : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err = req.Save(pID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch lignes de plan, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Batch lignes de plan importé"})
}
