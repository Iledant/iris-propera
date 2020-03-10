package actions

import (
	"database/sql"
	"net/http"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

type pmtPrevisionsResp struct {
	models.PmtPrevisions
	models.DifPmtPrevisions
}

// GetPaymentPrevisions handle the get request to calculate the payment previsions
// of the current year using the past commitments and the actual programmation
// using two statistical methods
func GetPaymentPrevisions(ctx iris.Context) {
	var resp pmtPrevisionsResp
	db := ctx.Values().Get("db").(*sql.DB)
	if err := resp.PmtPrevisions.Get(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévisions de paiement, requête 1 : " + err.Error()})
		return
	}
	if err := resp.DifPmtPrevisions.Get(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévisions de paiement, requête 2 : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetActionPaymentPrevisions handle the get request to calculate the payment
// previsions per action using the past commitments, the programmation of the
// actual year and the commitment previsions for the coming years
func GetActionPaymentPrevisions(ctx iris.Context) {
	var resp models.DifActionPmtPrevisions
	db := ctx.Values().Get("db").(*sql.DB)
	if err := resp.Get(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévisions de paiement par action, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetOpPaymentPrevisions handle the get request to calculate the payment
// previsions per operation using the past commitments, the programmation of the
// actual year and the commitment previsions for the coming years
func GetOpPaymentPrevisions(ctx iris.Context) {
	var resp models.DifOpPmtPrevisions
	db := ctx.Values().Get("db").(*sql.DB)
	if err := resp.Get(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévisions de paiement par opération, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetCurYearActionPmtPrevisions handle the get request to calculate the payment
// previsions per action using the past commitments and programmation of the
// current year and compares it to the payment of the current year
func GetCurYearActionPmtPrevisions(ctx iris.Context) {
	var resp models.CurYearActionPmtPrevisions
	db := ctx.Values().Get("db").(*sql.DB)
	if err := resp.Get(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévisions de paiement par action de l'année, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
