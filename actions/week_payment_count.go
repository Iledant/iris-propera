package actions

import (
	"database/sql"
	"net/http"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

// GetWeekPaymentCounts handle the get request to fetch all payment counts per
// week of the given year
func GetWeekPaymentCounts(ctx iris.Context) {
	year, err := ctx.URLParamInt64("year")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Paiements par semaine, décodage : " + err.Error()})
		return
	}
	var resp models.WeekPaymentCounts
	db := ctx.Values().Get("db").(*sql.DB)
	if err = resp.GetAll(year, db); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Paiements par semaine, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)

}
