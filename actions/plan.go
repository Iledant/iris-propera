package actions

import (
	"database/sql"
	"net/http"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

// GetPlans handles request get all plans.
func GetPlans(ctx iris.Context) {
	db := ctx.Values().Get("db").(*sql.DB)
	var resp models.Plans
	if err := resp.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des plans, requête :" + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

type planResp struct {
	Plan models.Plan `json:"Plan"`
}

// CreatePlan handles post request for creating a plan.
func CreatePlan(ctx iris.Context) {
	var req models.Plan
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de plan, décodage : " + err.Error()})
		return
	}
	if err := req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'un plan : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := req.Create(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Créatin d'un plan, requête : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(planResp{req})
}

// ModifyPlan handles put request for modifying a plan.
func ModifyPlan(ctx iris.Context) {
	pID, err := ctx.Params().GetInt64("pID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de plan, paramètre : " + err.Error()})
		return
	}
	var req models.Plan
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de plan, décodage : " + err.Error()})
		return
	}
	if err = req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Modification de plan : " + err.Error()})
		return
	}
	req.ID = pID
	db := ctx.Values().Get("db").(*sql.DB)
	if err = req.Update(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de plan, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(planResp{req})
}

// DeletePlan handles delete request for a plan.
func DeletePlan(ctx iris.Context) {
	pID, err := ctx.Params().GetInt64("pID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression de plan, erreur index : " + err.Error()})
		return
	}

	p := models.Plan{ID: pID}
	db := ctx.Values().Get("db").(*sql.DB)
	if err = p.Delete(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression de plan, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Plan supprimé"})
}
