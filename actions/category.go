package actions

import (
	"database/sql"
	"net/http"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

// caResp embeddes response for an single category.
type caResp struct {
	Category models.Category `json:"Category"`
}

// GetCategories handles request get all categories.
func GetCategories(ctx iris.Context) {
	var resp models.Categories
	db := ctx.Values().Get("db").(*sql.DB)
	if err := resp.GetAll(db); err != nil {
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
	db := ctx.Values().Get("db").(*sql.DB)
	if err := req.Create(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'une catégorie, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusCreated)
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
	db := ctx.Values().Get("db").(*sql.DB)
	if err = req.Update(db); err != nil {
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
	ca, db := models.Category{ID: caID}, ctx.Values().Get("db").(*sql.DB)
	if err = ca.Delete(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'une catégorie, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Catégorie supprimée"})
}

// stepsCategoriesResp embeddes the data for the frontend page dedicated to steps
// and categories.
type stepsCategoriesResp struct {
	models.Steps
	models.Categories
}

// GetStepsAndCategories handles the get request of the frontend page dedicated
// to steps and categories.
func GetStepsAndCategories(ctx iris.Context) {
	var resp stepsCategoriesResp
	db := ctx.Values().Get("db").(*sql.DB)
	if err := resp.Categories.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des catégories et étapes, requête catégories : " + err.Error()})
		return
	}
	if err := resp.Steps.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des catégories et étapes, requête étapes : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
