package actions

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Iledant/iris_propera/models"
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
	BudgetCredit models.CompleteBudgetCredit `json:"BudgetCredits"`
}

// budgetCreditsResp embeddes all datas for the frontend page.
type budgetCreditsResp struct {
	models.CompleteBudgetCredits
	models.BudgetChapters
}

// GetBudgetCredits handles request get all budget credits.
func GetBudgetCredits(ctx iris.Context) {
	db := ctx.Values().Get("db").(*sql.DB)
	var resp budgetCreditsResp
	if err := resp.CompleteBudgetCredits.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des crédits budgétaire, requête crédits : " + err.Error()})
		return
	}
	if err := resp.BudgetChapters.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des crédits budgétaire, requête chapitres : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetLastBudgetCredits handles request for getting the most recent budget credits of current year.
func GetLastBudgetCredits(ctx iris.Context) {
	db := ctx.Values().Get("db").(*sql.DB)
	year := int64(time.Now().Year())
	var resp models.BudgetCredits
	if err := resp.GetLatest(year, db); err != nil {
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
	db := ctx.Values().Get("db").(*sql.DB)
	if err := req.Create(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de crédits, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(brResp{req})
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
	req.ID = brID
	db := ctx.Values().Get("db").(*sql.DB)
	if err = req.Update(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de crédits, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(brResp{req})
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
	db := ctx.Values().Get("db").(*sql.DB)
	if err := req.Delete(db); err != nil {
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
	var req models.BudgetCreditBatch
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Erreur de lecture du batch crédits : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := req.Save(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch crédits, requête : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Credits importés"})
}
