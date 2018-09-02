package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// beeResp embeddes an array of beneficiaries for json response
type beeResp struct {
	Beneficiary []models.Beneficiary `json:"Beneficiary"`
}

// GetBeneficiaries handles the get all beneficiaries request
func GetBeneficiaries(ctx iris.Context) {
	beneficiaries := []models.Beneficiary{}

	db := ctx.Values().Get("db").(*gorm.DB)
	if err := db.Find(&beneficiaries).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des bénéficiaires, requête : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(beeResp{beneficiaries})
}

// beReq is used to decode the update payload
type beReq struct {
	Name *string `json:"name"`
}

// beResp embeddes an array of beneficiaries for json response
type beResp struct {
	Beneficiary models.Beneficiary `json:"Beneficiary"`
}

// UpdateBeneficiary handles the put request to change the name of one beneficiary.
func UpdateBeneficiary(ctx iris.Context) {
	bID, err := ctx.Params().GetInt("beneficiaryID")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Mise à jour bénéficiaire, décodage : " + err.Error()})
		return
	}

	req := beReq{}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Mise à jour bénéficiaire, payload : " + err.Error()})
		return
	}

	if req.Name == nil || *req.Name == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Modification de bénéficiaire : champ name manquant"})
		return
	}

	db, beneficiary := ctx.Values().Get("db").(*gorm.DB), models.Beneficiary{}

	if err = db.Raw("update beneficiary set name = ? where id = ? returning *", *req.Name, bID).
		Scan(&beneficiary).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Modification de bénéficiaire : introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de bénéficiaire, update : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(beResp{beneficiary})
}
