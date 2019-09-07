package actions

import (
	"database/sql"
	"net/http"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

// coResp embeddes response for an single commission.
type coResp struct {
	Commission models.Commission `json:"Commissions"`
}

// GetCommissions handles request get all commissions.
func GetCommissions(ctx iris.Context) {
	var resp models.Commissions
	db := ctx.Values().Get("db").(*sql.DB)
	if err := resp.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des commissions, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// CreateCommission handles request post request to create a new commission.
func CreateCommission(ctx iris.Context) {
	var req models.Commission
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'une commission, décodage : " + err.Error()})
		return
	}
	if err := req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'une commission : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := req.Create(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'une commission, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(coResp{req})
}

// ModifyCommission handles request put requestion to modify a commission.
func ModifyCommission(ctx iris.Context) {
	coID, err := ctx.Params().GetInt64("coID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'une commission, paramètre : " + err.Error()})
		return
	}
	var req models.Commission
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'une commission, décodage : " + err.Error()})
		return
	}
	if err = req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Modification d'une commission : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	req.ID = coID
	if err = req.Update(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'une commission, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(coResp{req})
}

// DeleteCommission handles the request to delete an commission.
func DeleteCommission(ctx iris.Context) {
	coID, err := ctx.Params().GetInt64("coID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'une commission, paramètre : " + err.Error()})
		return
	}
	co, db := models.Commission{ID: coID}, ctx.Values().Get("db").(*sql.DB)
	if err = co.Delete(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'une commission, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Commission supprimée"})
}
