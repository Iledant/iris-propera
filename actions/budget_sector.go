package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// bssResp embeddes response for an array of budget sectors.
type bssResp struct {
	BudgetSector []models.BudgetSector `json:"BudgetSector"`
}

// bsResp embeddes response for an single budget sector.
type bsResp struct {
	BudgetSector models.BudgetSector `json:"BudgetSector"`
}

// bsReq is used for creation and modification of a budget sector.
type bsReq struct {
	Code *string `json:"code"`
	Name *string `json:"name"`
}

// GetBudgetSectors handles request get all budget sectors.
func GetBudgetSectors(ctx iris.Context) {
	bss := []models.BudgetSector{}

	db := ctx.Values().Get("db").(*gorm.DB)
	if err := db.Find(&bss).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(bssResp{bss})
}

// CreateBudgetSector handles request post request to create a new sector.
func CreateBudgetSector(ctx iris.Context) {
	bs := bsReq{}
	if err := ctx.ReadJSON(&bs); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if bs.Code == nil || len(*bs.Code) == 0 || len(*bs.Code) > 10 ||
		bs.Name == nil || len(*bs.Name) == 0 || len(*bs.Name) > 100 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création de secteur budgétaire, champ manquant ou incorrect"})
		return
	}

	db := ctx.Values().Get("db").(*gorm.DB)
	newBs := models.BudgetSector{Code: *bs.Code, Name: *bs.Name}

	if err := db.Create(&newBs).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(bsResp{newBs})
}

// ModifyBudgetSector handles request put requestion to modify a sector.
func ModifyBudgetSector(ctx iris.Context) {
	bsID, err := ctx.Params().GetInt("bsID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	bs, db := models.BudgetSector{}, ctx.Values().Get("db").(*gorm.DB)

	if err = db.First(&bs, bsID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonMessage{"Modification de secteur : introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	req := bsReq{}
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if req.Code != nil && len(*req.Code) > 0 && len(*req.Code) < 10 {
		bs.Code = *req.Code
	}

	if req.Name != nil && len(*req.Name) > 0 && len(*req.Name) < 100 {
		bs.Name = *req.Name
	}

	if err = db.Save(&bs).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(bsResp{bs})
}

// DeleteBudgetSector handles the request to delete an budget sector.
func DeleteBudgetSector(ctx iris.Context) {
	bsID, err := ctx.Params().GetInt("bsID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	bs, db := models.BudgetSector{ID: bsID}, ctx.Values().Get("db").(*gorm.DB)

	if err = db.First(&bs, bsID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(jsonError{"Suppression de secteur : introuvable"})
		return
	}

	if err = db.Delete(&bs).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Secteur supprimé"})
}
