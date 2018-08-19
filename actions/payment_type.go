package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// pttResp is used to embeddes response of array of payment types.
type pttResp struct {
	Pts []models.PaymentType `json:"PaymentType"`
}

// GetPaymentTypes handles request get all payments types (chronicles names).
func GetPaymentTypes(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	ptt := pttResp{}

	if err := db.Find(&ptt.Pts).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(ptt)
}

// ptResp embeddes response for a single payment type
type ptResp struct {
	PaymentType models.PaymentType `json:"PaymentType"`
}

// sentPt is used to decode sent datas to create a payment type
type sentPt struct {
	Name string `json:"name"`
}

// CreatePaymentType handles post request for creating a payment type.
func CreatePaymentType(ctx iris.Context) {
	req := sentPt{}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if req.Name == "" || len(req.Name) > 255 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'une chronique : mauvais format de name"})
		return
	}

	resp, db := ptResp{}, ctx.Values().Get("db").(*gorm.DB)
	resp.PaymentType.Name = req.Name
	if err := db.Create(&resp.PaymentType).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'une chronique : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// ModifyPaymentType handles put request for modifying a payment type.
func ModifyPaymentType(ctx iris.Context) {
	ptID, err := ctx.Params().GetInt("ptID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	pt, db := models.PaymentType{}, ctx.Values().Get("db").(*gorm.DB)
	if err := db.Find(&pt, ptID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Modification d'une chronique : introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	req := sentPt{}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if req.Name != "" && len(req.Name) < 255 {
		pt.Name = req.Name
	}

	if err = db.Save(&pt).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'une chronique : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(ptResp{pt})
}

// DeletePaymentType handles delete request for a payment type.
func DeletePaymentType(ctx iris.Context) {
	ptID, err := ctx.Params().GetInt("ptID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	pt, db := models.PaymentType{}, ctx.Values().Get("db").(*gorm.DB)
	if err := db.Find(&pt, ptID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Suppression d'une chronique : introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	tx := db.Begin()
	// Delete payment ratios linked
	if err = tx.Exec("DELETE from payment_ratios where payment_types_id = ?", ptID).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	// Remove physical operations link
	if err = tx.Exec("UPDATE physical_op SET payment_types_id = null WHERE payment_types_id = ?", ptID).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	if err = tx.Delete(&pt).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'une chronique : " + err.Error()})
		tx.Rollback()
		return
	}

	tx.Commit()
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Chronique supprimée"})
}
