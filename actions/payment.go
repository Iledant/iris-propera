package actions

import (
	"net/http"
	"time"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// GetFcPayment handles the get request fetching all payments of a financial commitment.
func GetFcPayment(ctx iris.Context) {
	fcID, err := ctx.Params().GetInt64("fcID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiements d'un engagement, paramètre : " + err.Error()})
		return
	}
	var resp models.Payments
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.GetFcAll(fcID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiements d'un engagement, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetPaymentsPerMonth handles the get request fetching payments per month of a given year and the precedent.
func GetPaymentsPerMonth(ctx iris.Context) {
	y, err := ctx.URLParamInt("year")
	if err != nil {
		y = time.Now().Year()
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	var resp models.PaymentPerMonths
	if err = resp.GetAll(y, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiements par mois, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// BatchPayments handles the request sending an array of payments.
func BatchPayments(ctx iris.Context) {
	var req models.PaymentBatch
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch de paiements, décodage : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := req.Save(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch de paiements, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Paiements importés"})
}

// GetPrevisionRealized handles the request to the payment prevision and real payments for the given year and beneficiary.
func GetPrevisionRealized(ctx iris.Context) {
	year, err := ctx.URLParamInt64("year")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Prévu réalisé erreur sur year : " + err.Error()})
		return
	}
	ptID, err := ctx.URLParamInt64("paymentTypeId")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévu réalisé erreur sur paymentTypeId : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	var resp models.PrevisionsRealized
	if err = resp.GetAll(year, ptID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévu réalisé, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetCumulatedMonthPayment handles the request to calculate cumulated payment per month for all or for one beneficiary.
func GetCumulatedMonthPayment(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	bID, err := ctx.URLParamInt64("beneficiaryId")
	if err != nil {
		bID = 0
	}
	var resp models.MonthCumulatedPayments
	if err := resp.GetAll(bID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiement cumulés, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

type getAllPaymentsResp struct {
	models.PaymentPerMonths
	models.MonthCumulatedPayments
	models.Beneficiaries
	models.PaymentTypes
}

// GetAllPayments handle the get request to fetch all datas linked to payment frontend page.
func GetAllPayments(ctx iris.Context) {
	year, err := ctx.URLParamInt("year")
	if err != nil {
		year = time.Now().Year()
	}
	var resp getAllPaymentsResp
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.PaymentPerMonths.GetAll(year, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Tous les paiements, paiements par mois : " + err.Error()})
		return
	}
	if err = resp.MonthCumulatedPayments.GetAll(0, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Tous les paiements, paiements cumulés : " + err.Error()})
		return
	}
	if err = resp.Beneficiaries.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Tous les paiements, bénéficiaires : " + err.Error()})
		return
	}
	if err = resp.PaymentTypes.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Tous les paiements, chroniques de paiement : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
