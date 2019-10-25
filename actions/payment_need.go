package actions

import (
	"database/sql"
	"net/http"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

type pmtNeedReq struct {
	PaymentNeed models.PaymentNeed `json:"PaymentNeed"`
}

// CreatePaymentNeed handle the post request to create a new payment need
func CreatePaymentNeed(ctx iris.Context) {
	var req pmtNeedReq
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'un besoin de paiement, décodage : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := req.PaymentNeed.Create(db); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'un besoin de paiement, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(req)
}

// ModifyPaymentNeed handle the post request to modify a payment need
func ModifyPaymentNeed(ctx iris.Context) {
	var req pmtNeedReq
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Modification d'un besoin de paiement, décodage : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := req.PaymentNeed.Update(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un besoin de paiement, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(req)
}

// DeletePaymentNeed handle the delete request to remove a payment need from database
func DeletePaymentNeed(ctx iris.Context) {
	ID, err := ctx.Params().GetInt64("ID")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Suppression d'un besoin de paiement, décodage : " + err.Error()})
		return
	}
	req := models.PaymentNeed{ID: ID}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := req.Delete(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'un besoin de paiement, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Besoin de paiement supprimé"})
}

// GetPaymentNeeds handle the get request to calculates the latest payment needs
// and forecast of the year
func GetPaymentNeeds(ctx iris.Context) {
	year, err := ctx.URLParamInt64("Year")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Besoins de paiement, décodage year : " + err.Error()})
		return
	}
	pmtTypeID, err := ctx.URLParamInt64("PaymentTypeID")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Besoins de paiement, décodage paymentTypeID : " + err.Error()})
		return
	}
	var resp models.LastPaymentNeeds
	db := ctx.Values().Get("db").(*sql.DB)
	if err = resp.GetAll(year, pmtTypeID, db); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Besoins de paiement, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
