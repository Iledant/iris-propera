package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// caResp embeddes response for an single category.
type caResp struct {
	Category models.Category `json:"Category"`
}

// caReq is used for creation and modification of a category.
type caReq struct {
	Name *string `json:"name"`
}

// Invalid checks if request fields are present and well formed.
func (req caReq) Invalid() bool {
	return req.Name == nil || len(*req.Name) == 0 || len(*req.Name) > 50
}

// Populate fills category's fields with valid request content.
func (req caReq) Populate(ID int64, c *models.Category) {
	c.Name = *req.Name
	c.ID = ID
}

// GetCategories handles request get all categories.
func GetCategories(ctx iris.Context) {
	var resp models.Categories
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := resp.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des catégories, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// CreateCategory handles request post request to create a new category.
func CreateCategory(ctx iris.Context) {
	var req models.Category
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'une catégorie, décodage : " + err.Error()})
		return
	}
	if err := req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'une catégorie : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := req.Create(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'une catégorie, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(caResp{req})
}

// ModifyCategory handles request put requestion to modify a category.
func ModifyCategory(ctx iris.Context) {
	caID, err := ctx.Params().GetInt64("caID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'une catégorie, paramètre : " + err.Error()})
		return
	}
	var req models.Category
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'une catégorie, décodage : " + err.Error()})
		return
	}
	if err = req.Validate(); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'une catégorie : " + err.Error()})
		return
	}
	req.ID = caID
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = req.Update(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'une catégorie, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(caResp{req})
}

// DeleteCategory handles the request to delete an category.
func DeleteCategory(ctx iris.Context) {
	caID, err := ctx.Params().GetInt64("caID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'une catégorie, paramètre : " + err.Error()})
		return
	}
	ca, db := models.Category{ID: caID}, ctx.Values().Get("db").(*gorm.DB)
	if err = ca.Delete(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'une catégorie, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Catégorie supprimée"})
}
