package actions

import (
	"database/sql"
	"net/http"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

type flowStockDelayResp struct {
	FlowStockDelays models.FlowStockDelays `json:"FlowStockDelays"`
}

// GetFlowStockDelays handles the get all beneficiaries request
func GetFlowStockDelays(ctx iris.Context) {
	var resp flowStockDelayResp
	days, err := ctx.URLParamInt64("Days")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON((jsonError{"Délais de flux et de stock, paramètre : " + err.Error()}))
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := resp.FlowStockDelays.Get(days, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Délais de flux et de stock, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
