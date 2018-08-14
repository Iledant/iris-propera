package actions

import (
	"net/http"
	"time"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// cooResp embeddes response for an array of commissions.
type cooResp struct {
	Commission []models.Commission `json:"Commissions"`
}

// coResp embeddes response for an single commission.
type coResp struct {
	Commission models.Commission `json:"Commissions"`
}

// coReq is used for creation and modification of a commission.
type coReq struct {
	Name *string    `json:"name"`
	Date *time.Time `json:"date"`
}

// GetCommissions handles request get all commissions.
func GetCommissions(ctx iris.Context) {
	coo := []models.Commission{}

	db := ctx.Values().Get("db").(*gorm.DB)
	if err := db.Find(&coo).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(cooResp{coo})
}

// CreateCommission handles request post request to create a new commission.
func CreateCommission(ctx iris.Context) {
	co := coReq{}
	if err := ctx.ReadJSON(&co); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if co.Name == nil || len(*co.Name) == 0 || len(*co.Name) > 50 || co.Date == nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création de commission, champ manquant ou incorrect"})
		return
	}

	db := ctx.Values().Get("db").(*gorm.DB)
	newCo := models.Commission{Name: *co.Name, Date: *co.Date}

	if err := db.Create(&newCo).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(coResp{newCo})
}

// ModifyCommission handles request put requestion to modify a commission.
func ModifyCommission(ctx iris.Context) {
	coID, err := ctx.Params().GetInt("coID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	co, db := models.Commission{}, ctx.Values().Get("db").(*gorm.DB)

	if err = db.First(&co, coID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonMessage{"Modification de commission : introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	req := coReq{}
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if req.Name != nil && len(*req.Name) > 0 && len(*req.Name) < 50 {
		co.Name = *req.Name
	}

	if req.Date != nil {
		co.Date = *req.Date
	}

	if err = db.Save(&co).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(coResp{co})
}

// DeleteCommission handles the request to delete an commission.
func DeleteCommission(ctx iris.Context) {
	coID, err := ctx.Params().GetInt("coID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	co, db := models.Commission{ID: coID}, ctx.Values().Get("db").(*gorm.DB)

	if err = db.First(&co, coID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(jsonError{"Suppression de commission : introuvable"})
		return
	}

	if err = db.Delete(&co).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Commission supprimée"})
}
