package actions

import (
	"database/sql"
	"net/http"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

// GetPlanForecasts handles request get operations datas and forecasts
func GetPlanForecasts(ctx iris.Context) {
	firstYear, err := ctx.URLParamInt64("firstYear")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Prévisions de plan, firstYear : " + err.Error()})
		return
	}
	lastYear, err := ctx.URLParamInt64("lastYear")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Prévisions de plan, lastYear : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	var resp models.PlanForecasts
	if err := resp.GetAll(db, firstYear, lastYear); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévisions de plan, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
