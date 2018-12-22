package actions

import (
	"net/http"
	"time"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// brReq is used to analyse creating or update requests for budget credits.
type brReq struct {
	CommissionDate     *time.Time `json:"commission_date"`
	Chapter            *int64     `json:"chapter"`
	PrimaryCommitment  *int64     `json:"primary_commitment"`
	FrozenCommitment   *int64     `json:"frozen_commitment"`
	ReservedCommitment *int64     `json:"reserved_commitment"`
}

// brResp embeddes response for a single budget credits.
type brResp struct {
	BudgetCredit models.BudgetCredit `json:"BudgetCredits"`
}

// GetBudgetCredits handles request get all budget credits.
func GetBudgetCredits(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	var resp models.BudgetCredits
	if err := resp.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des crédits budgétaire, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetLastBudgetCredits handles request for getting the most recent budget credits of current year.
func GetLastBudgetCredits(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	year := int64(time.Now().Year())
	var resp models.BudgetCredits
	if err := resp.GetLatest(year, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Crédits budgétaires les plus récents, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// CreateBudgetCredit handles post request for creating a budget credit.
func CreateBudgetCredit(ctx iris.Context) {
	var req models.CompleteBudgetCredit
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de crédits, décodage : " + err.Error()})
		return
	}
	if err := req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création de crédits : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	var resp models.BudgetCredit
	if err := resp.Create(&req, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de crédits, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(brResp{resp})
}

// ModifyBudgetCredit handles put request for modifying budget credits.
func ModifyBudgetCredit(ctx iris.Context) {
	brID, err := ctx.Params().GetInt64("brID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de crédits, paramètre : " + err.Error()})
		return
	}
	var req models.CompleteBudgetCredit
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de crédits, décodage : " + err.Error()})
		return
	}
	if err = req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Modification de crédits " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	resp := models.BudgetCredit{ID: brID}
	if err = resp.Update(&req, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de crédits, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(brResp{resp})
}

// DeleteBudgetCredit handles delete request for budget credits.
func DeleteBudgetCredit(ctx iris.Context) {
	brID, err := ctx.Params().GetInt64("brID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression de crédits, décodage : " + err.Error()})
		return
	}
	req := models.BudgetCredit{ID: brID}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := req.Delete(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression de crédits, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Crédits supprimés"})
}

// batchBrr is used to embed batch credits imports
type batchBrr struct {
	BudgetCredits []brReq `json:"BudgetCredits"`
}

// BatchBudgetCredits handles the post array request for budget credits
func BatchBudgetCredits(ctx iris.Context) {
	var req models.CompleteBudgetCredits
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Erreur de lecture du batch crédits : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := req.Save(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch crédits, requête : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Credits importés"})
}
