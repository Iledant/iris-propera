package actions

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

// BatchPaymentCredits handle the post request for a batch of payment credits
func BatchPaymentCredits(ctx iris.Context) {
	var req models.PaymentCreditBatch
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Batch d'enveloppes de crédits, décodage : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	year := (int64)(time.Now().Year())
	if err := req.Save(year, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch d'enveloppes de crédits, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Enveloppes de crédits importées"})
}

// GetAllPaymentCredits handles the get request to get all payment credits of
// the given year
func GetAllPaymentCredits(ctx iris.Context) {
	year, err := ctx.URLParamInt64("Year")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Liste des enveloppes de crédits, décodage : " + err.Error()})
		return
	}
	var resp models.PaymentCredits
	db := ctx.Values().Get("db").(*sql.DB)
	if err = resp.GetAll(year, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des enveloppes de crédits, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
