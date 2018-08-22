package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

type plansResp struct {
	Plans []models.Plan `json:"Plan"`
}

// GetPlans handles request get all plans.
func GetPlans(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	pp := plansResp{}

	if err := db.Find(&pp.Plans).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des plans :" + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(pp)
}

type planResp struct {
	Plan models.Plan `json:"Plan"`
}

type sentPlanReq struct {
	Name      *string `json:"name"`
	Descript  *string `json:"descript"`
	FirstYear *int64  `json:"first_year"`
	LastYear  *int64  `json:"last_year"`
}

// CreatePlan handles post request for creating a plan.
func CreatePlan(ctx iris.Context) {
	req := sentPlanReq{}

	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de plan, impossible de décoder : " + err.Error()})
		return
	}

	if req.Name == nil || len(*req.Name) > 255 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'un plan : mauvais format de name"})
		return
	}

	p := models.Plan{Name: *req.Name}
	if req.Descript != nil {
		p.Descript.Valid = true
		p.Descript.String = *req.Descript
	}
	if req.FirstYear != nil {
		p.FirstYear.Valid = true
		p.FirstYear.Int64 = *req.FirstYear
	}
	if req.LastYear != nil {
		p.LastYear.Valid = true
		p.LastYear.Int64 = *req.LastYear
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := db.Create(&p).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Créatin d'un plan : erreur d'insert : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(planResp{p})
}

// ModifyPlan handles put request for modifying a plan.
func ModifyPlan(ctx iris.Context) {
	pID, err := ctx.Params().GetInt64("pID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de plan, erreur identificateur : " + err.Error()})
		return
	}

	p := models.Plan{}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := db.Find(&p, pID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Modification de plan: introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de plan, erreur select : " + err.Error()})
		return
	}

	req := sentPlanReq{}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de plan, erreur décodage : " + err.Error()})
		return
	}

	if req.Name != nil && len(*req.Name) < 255 {
		p.Name = *req.Name
	}
	if req.Descript != nil {
		p.Descript.Valid = true
		p.Descript.String = *req.Descript
	}
	if req.FirstYear != nil {
		p.FirstYear.Valid = true
		p.FirstYear.Int64 = *req.FirstYear
	}
	if req.LastYear != nil {
		p.LastYear.Valid = true
		p.LastYear.Int64 = *req.LastYear
	}

	if err = db.Save(&p).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de plan, erreur update : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(planResp{p})
}

// DeletePlan handles delete request for a plan.
func DeletePlan(ctx iris.Context) {
	pID, err := ctx.Params().GetInt("pID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression de plan, erreur index : " + err.Error()})
		return
	}

	p := models.Plan{}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := db.Find(&p, pID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Suppression d'un plan: introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	tx := db.Begin()

	if err = tx.Exec("DELETE from plan_line_ratios WHERE plan_line_id IN (SELECT id FROM plan_line WHERE plan_id = ?)", pID).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'un plan, erreur de delete ratios : " + err.Error()})
		tx.Rollback()
		return
	}

	if err = tx.Exec("DELETE from plan_line WHERE plan_id = ?", pID).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'un plan, erreur de delete plan_line : " + err.Error()})
		tx.Rollback()
		return
	}

	if err = tx.Exec("DELETE from plan WHERE id = ?", pID).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'un plan, erreur de delete : " + err.Error()})
		tx.Rollback()
		return
	}

	tx.Commit()
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Plan supprimé"})
}
