package actions

import (
	"database/sql"
	"net/http"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

// bsResp embeddes response for an single budget sector.
type bsResp struct {
	BudgetSector models.BudgetSector `json:"BudgetSector"`
}

// GetBudgetSectors handles request get all budget sectors.
func GetBudgetSectors(ctx iris.Context) {
	var resp models.BudgetSectors
	db := ctx.Values().Get("db").(*sql.DB)
	if err := resp.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des secteurs budgétaire, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// CreateBudgetSector handles request post request to create a new sector.
func CreateBudgetSector(ctx iris.Context) {
	var req models.BudgetSector
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'un secteur budgétaire, décodage : " + err.Error()})
		return
	}
	if err := req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'un secteur budgétaire : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := req.Create(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'un secteur budgétaire, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(bsResp{req})
}

// ModifyBudgetSector handles request put requestion to modify a sector.
func ModifyBudgetSector(ctx iris.Context) {
	bsID, err := ctx.Params().GetInt64("bsID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un secteur budgétaire, paramètre : " + err.Error()})
		return
	}
	var req models.BudgetSector
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un secteur budgétaire, décodage : " + err.Error()})
		return
	}
	if err = req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Modification d'un secteur budgétaire " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	req.ID = bsID
	if err = req.Update(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un secteur budgétaire, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(bsResp{req})
}

// DeleteBudgetSector handles the request to delete an budget sector.
func DeleteBudgetSector(ctx iris.Context) {
	bsID, err := ctx.Params().GetInt64("bsID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'un secteur budgétaire, paramètre : " + err.Error()})
		return
	}
	bs, db := models.BudgetSector{ID: bsID}, ctx.Values().Get("db").(*sql.DB)
	if err = bs.Delete(db); err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(jsonError{"Suppression d'un secteur budgétaire, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Secteur supprimé"})
}
