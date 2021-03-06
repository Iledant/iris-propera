package actions

import (
	"database/sql"
	"net/http"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

// GetImportLogs handles to get request for log informations
func GetImportLogs(ctx iris.Context) {
	db := ctx.Values().Get("db").(*sql.DB)
	var resp models.ImportLogs
	if err := resp.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Import logs, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
