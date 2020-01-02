package actions

import (
	"database/sql"
	"net/http"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

// GetPossibleLinkedCmts handle the get request to fetch the possible commitments
// linked to a payment
func GetPossibleLinkedCmts(ctx iris.Context) {
	var resp models.PossibleLinkedCmts
	pmtID, err := ctx.Params().GetInt64("pmtID")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Engagements possiblement liés, décodage : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := resp.Get(pmtID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Engagements possiblement liés, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
