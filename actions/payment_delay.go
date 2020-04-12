package actions

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

// GetPaymentDelays handles the get request to fetch payment delay after a
// given date
func GetPaymentDelays(ctx iris.Context) {
	after, err := ctx.URLParamInt64("after")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Délais de paiement, décodage : " + err.Error()})
		return
	}
	afterTime := time.Unix(after/1000, 0)
	var resp models.PaymentDelays
	db := ctx.Values().Get("db").(*sql.DB)
	if err = resp.GetSome(afterTime, db); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Délais de paiement, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
