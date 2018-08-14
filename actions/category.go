package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// caaResp embeddes response for an array of categories.
type caaResp struct {
	Category []models.Category `json:"Category"`
}

// caResp embeddes response for an single category.
type caResp struct {
	Category models.Category `json:"Category"`
}

// caReq is used for creation and modification of a category.
type caReq struct {
	Name *string `json:"name"`
}

// GetCategories handles request get all categories.
func GetCategories(ctx iris.Context) {
	caa := []models.Category{}

	db := ctx.Values().Get("db").(*gorm.DB)
	if err := db.Find(&caa).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(caaResp{caa})
}

// CreateCategory handles request post request to create a new category.
func CreateCategory(ctx iris.Context) {
	ca := caReq{}
	if err := ctx.ReadJSON(&ca); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if ca.Name == nil || len(*ca.Name) == 0 || len(*ca.Name) > 50 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création de catégorie, champ 'name' manquant ou incorrect"})
		return
	}

	db := ctx.Values().Get("db").(*gorm.DB)
	newCa := models.Category{Name: *ca.Name}

	if err := db.Create(&newCa).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(caResp{newCa})
}

// ModifyCategory handles request put requestion to modify a category.
func ModifyCategory(ctx iris.Context) {
	caID, err := ctx.Params().GetInt("caID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ca, db := models.Category{}, ctx.Values().Get("db").(*gorm.DB)

	if err = db.First(&ca, caID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonMessage{"Modification de catégorie : introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	req := caReq{}
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if req.Name != nil && len(*req.Name) > 0 && len(*req.Name) < 50 {
		ca.Name = *req.Name
	}

	if err = db.Save(&ca).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(caResp{ca})
}

// DeleteCategory handles the request to delete an category.
func DeleteCategory(ctx iris.Context) {
	caID, err := ctx.Params().GetInt("caID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ca, db := models.Category{ID: caID}, ctx.Values().Get("db").(*gorm.DB)

	if err = db.First(&ca, caID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(jsonError{"Suppression de catégorie : introuvable"})
		return
	}

	if err = db.Delete(&ca).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Catégorie supprimée"})
}
