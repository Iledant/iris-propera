package actions

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

// GetRatios handles the get request to fetch all payment ratios.
func GetRatios(ctx iris.Context) {
	db := ctx.Values().Get("db").(*sql.DB)
	var resp models.PaymentRatios
	if err := resp.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des ratios de paiement, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetPtRatios handles the get request to fetch ratios linked to a payment type.
func GetPtRatios(ctx iris.Context) {
	db := ctx.Values().Get("db").(*sql.DB)
	var resp models.PaymentRatios
	ptID, err := ctx.Params().GetInt64("ptID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des ratios d'une chronique, paramètre : " + err.Error()})
		return
	}
	if err = resp.GetPaymentTypeAll(ptID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des ratios d'une chronique, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// DeleteRatios handles the delete request for a ratios linked to a payment type.
func DeleteRatios(ctx iris.Context) {
	ptID, err := ctx.Params().GetInt64("ptID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression des ratios d'une chronique, paramètre : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	pt := models.PaymentType{ID: ptID}
	if err = pt.DeleteRatios(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression des ratios d'une chronique, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Ratios supprimés"})
}

// SetPtRatios handle the post request for setting all ratios of an payment type.
func SetPtRatios(ctx iris.Context) {
	db := ctx.Values().Get("db").(*sql.DB)
	ptID, err := ctx.Params().GetInt64("ptID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Ratios d'une chronique, paramètre : " + err.Error()})
		return
	}
	var req models.PaymentRatiosBatch
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Ratios d'une chronique, décodage : " + err.Error()})
		return
	}
	if err = req.Save(ptID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Ratios d'une chronique, requête : " + err.Error()})
		return
	}
	var resp models.PaymentRatios
	if err = resp.GetPaymentTypeAll(ptID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Ratios d'une chronique, requête get : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetYearRatios handles the get request to fetch the payment ratios linked to the financial commitments of a given year.
func GetYearRatios(ctx iris.Context) {
	year := ctx.URLParam("Year")
	if year == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Ratios annuels : année manquante"})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	y, err := strconv.Atoi(year)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Ratios annuels, format année : " + err.Error()})
		return
	}
	var resp models.YearRatios
	if err = resp.GetAll(int64(y), db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Ratios annuels, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
