package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// GetBeneficiaries handles the get all beneficiaries request
func GetBeneficiaries(ctx iris.Context) {
	var resp models.Beneficiaries
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := resp.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des bénéficiaires, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// beResp embeddes a beneficiaries for json response
type beResp struct {
	Beneficiary models.Beneficiary `json:"Beneficiary"`
}

// UpdateBeneficiary handles the put request to change the name of one beneficiary.
func UpdateBeneficiary(ctx iris.Context) {
	bID, err := ctx.Params().GetInt("beneficiaryID")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Mise à jour bénéficiaire, paramètre : " + err.Error()})
		return
	}
	var req models.Beneficiary
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Mise à jour bénéficiaire, décodage : " + err.Error()})
		return
	}
	if err = req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Modification de bénéficiaire : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	req.ID = bID
	if err = req.Update(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de bénéficiaire, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(beResp{req})
}
