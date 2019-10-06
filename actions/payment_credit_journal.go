package actions

import (
	"database/sql"
	"net/http"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

// BatchPaymentCreditJournals handle the post request for a batch of payment credits
func BatchPaymentCreditJournals(ctx iris.Context) {
	var req models.PaymentCreditJournalBatch
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Batch mouvements de crédits, décodage : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := req.Save(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch mouvements de crédits, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Mouvements de crédits importés"})
}

// GetAllPaymentCreditJournals handles the get request to get all payment credits of
// the given year
func GetAllPaymentCreditJournals(ctx iris.Context) {
	year, err := ctx.URLParamInt("Year")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Mouvements de crédits, décodage : " + err.Error()})
		return
	}
	var resp models.PaymentCreditJournals
	db := ctx.Values().Get("db").(*sql.DB)
	if err = resp.GetAll(year, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Mouvements de crédits, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
