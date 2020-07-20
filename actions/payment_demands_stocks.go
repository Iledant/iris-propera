package actions

import (
	"database/sql"
	"net/http"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

// GetPaymentDemandStocks handle the get requests to fetch to payment demands
// stocks over the 30 last days
func GetPaymentDemandStocks(ctx iris.Context) {
	var resp models.PaymentDemandsStocks
	db := ctx.Values().Get("db").(*sql.DB)
	if err := resp.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Stocks de DVS, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
