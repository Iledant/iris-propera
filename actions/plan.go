package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// GetPlans handles request get all plans.
func GetPlans(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	var resp models.Plans
	if err := resp.GetAll(db.DB()); err != nil {
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

// Invalid checks of sent plan is well formated.
func Invalid(p *models.Plan) bool {
	return p.Name == "" || len(p.Name) > 255
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
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := req.Create(db.DB()); err != nil {
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
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = req.Update(db.DB()); err != nil {
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
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = p.Delete(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression de plan, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Plan supprimé"})
}
