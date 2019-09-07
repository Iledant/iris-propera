package actions

import (
	"database/sql"
	"net/http"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

// stResp embeddes response for an single step.
type stResp struct {
	Step models.Step `json:"Step"`
}

// GetSteps handles request get all steps.
func GetSteps(ctx iris.Context) {
	var resp models.Steps
	db := ctx.Values().Get("db").(*sql.DB)
	if err := resp.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des étapes, requête  : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// CreateStep handles request post request to create a new step.
func CreateStep(ctx iris.Context) {
	var req models.Step
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'étape, décodage : " + err.Error()})
		return
	}
	if err := req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'étape : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := req.Create(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'étape, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(stResp{req})
}

// ModifyStep handles request put requestion to modify a step.
func ModifyStep(ctx iris.Context) {
	stID, err := ctx.Params().GetInt64("stID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'étape, paramètre : " + err.Error()})
		return
	}
	req, db := models.Step{}, ctx.Values().Get("db").(*sql.DB)
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'étape, décodage :" + err.Error()})
		return
	}
	if err = req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Modification d'étape : " + err.Error()})
		return
	}
	req.ID = stID
	if err = req.Update(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'étape, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(stResp{req})
}

// DeleteStep handles the request to delete an step.
func DeleteStep(ctx iris.Context) {
	stID, err := ctx.Params().GetInt64("stID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'étape, paramètre : " + err.Error()})
		return
	}
	st, db := models.Step{ID: stID}, ctx.Values().Get("db").(*sql.DB)
	if err = st.Delete(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'étape, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Etape supprimée"})
}
