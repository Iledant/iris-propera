package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// GetBeneficiaries handles the get all beneficiaries request
func GetBeneficiaries(ctx iris.Context) {
	beneficiaries := []models.Beneficiary{}

	db := ctx.Values().Get("db").(*gorm.DB)
	if err := db.Find(&beneficiaries).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(struct {
		Beneficiary []models.Beneficiary `json:"Beneficiary"`
	}{beneficiaries})
}

type beneficiaryReq struct {
	Name string `json:"name"`
}

// UpdateBeneficiary handles the put request to change the name of one beneficiary.
func UpdateBeneficiary(ctx iris.Context) {
	bID, err := ctx.Params().GetInt("beneficiaryID")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	req := beneficiaryReq{}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if req.Name == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Modification de bénéficiaire : champ name manquant"})
		return
	}

	db, beneficiary := ctx.Values().Get("db").(*gorm.DB), models.Beneficiary{ID: bID}

	if err = db.First(&beneficiary, bID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusNotFound)
			ctx.JSON(jsonError{"Modification de bénéficiaire : introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	beneficiary.Name = req.Name

	if err = db.Save(&beneficiary).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(struct {
		Beneficiary models.Beneficiary `json:"Beneficiary"`
	}{beneficiary})
}
