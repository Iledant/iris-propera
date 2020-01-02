package actions

import (
	"database/sql"
	"net/http"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

type consistencyResp struct {
	models.CommitmentWithoutActions
	models.UnlinkedPayments
}

// GetConsistencyDatas handle the get request to fetch consistency datas from
// database
func GetConsistencyDatas(ctx iris.Context) {
	var resp consistencyResp
	db := ctx.Values().Get("db").(*sql.DB)
	if err := resp.CommitmentWithoutActions.Get(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Datas de cohérence, engagements : " + err.Error()})
		return
	}
	if err := resp.UnlinkedPayments.Get(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Datas de cohérence, paiements : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// LinkPaymentToCmt handle the post request to link a payment to a commitment
func LinkPaymentToCmt(ctx iris.Context) {
	pmtID, err := ctx.Params().GetInt64("pmtID")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Lien paiement engagement, décodage ID paiement : " + err.Error()})
		return
	}
	cmtID, err := ctx.Params().GetInt64("cmtID")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Lien paiement engagement, décodage ID engagement : " + err.Error()})
		return
	}
	p := models.Payment{ID: pmtID}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := p.LinkCmt(cmtID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Lien paiement engagement, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Paiement rattaché à l'engagement"})

}
