package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// GetPaymentTypes handles request get all payments types (chronicles names).
func GetPaymentTypes(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	var resp models.PaymentTypes
	if err := resp.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des chroniques de paiement, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// ptResp embeddes response for a single payment type
type ptResp struct {
	PaymentType models.PaymentType `json:"PaymentType"`
}

// CreatePaymentType handles post request for creating a payment type.
func CreatePaymentType(ctx iris.Context) {
	var req models.PaymentType
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'une chronique de paiement, décodage : " + err.Error()})
		return
	}
	if err := req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'une chronique de paiement : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := req.Create(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'une chronique de paiement : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(ptResp{req})
}

// ModifyPaymentType handles put request for modifying a payment type.
func ModifyPaymentType(ctx iris.Context) {
	ptID, err := ctx.Params().GetInt64("ptID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'une chronique de paiement, paramètre : " + err.Error()})
		return
	}
	var req models.PaymentType
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'une chronique de paiement, décodage : " + err.Error()})
		return
	}
	if err = req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Modification d'une chronique de paiement : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	req.ID = ptID
	if err = req.Update(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'une chronique de paiement, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(ptResp{req})
}

// DeletePaymentType handles delete request for a payment type.
func DeletePaymentType(ctx iris.Context) {
	ptID, err := ctx.Params().GetInt64("ptID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'une chronique de paiement, paramètre : " + err.Error()})
		return
	}
	pt, db := models.PaymentType{ID: ptID}, ctx.Values().Get("db").(*gorm.DB)
	if err = pt.Delete(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'une chronique de paiement, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Chronique supprimée"})
}
