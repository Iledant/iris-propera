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
// of the current year according to the past commitments and the actual programmation
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

// GetMultiAnnualPaymentPrevision handle the get request to calculate the
// payment previsions for the 5 coming years using the average differential
// ratio methods
func GetMultiAnnualPaymentPrevision(ctx iris.Context) {
	var resp models.MultiannualDifPmtPrevisions
	db := ctx.Values().Get("db").(*sql.DB)
	if err := resp.Get(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévisions pluriannuelle différentielles de paiement, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
