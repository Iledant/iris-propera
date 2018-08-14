package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// dooResp embeddes response for an array of documents.
type dooResp struct {
	Document []models.Document `json:"Document"`
}

// doResp embeddes response for an single document.
type doResp struct {
	Document models.Document `json:"Document"`
}

// doReq is used for creation and modification of a document.
type doReq struct {
	PhysicalOpID *int    `json:"physical_op_id"`
	Name         *string `json:"name"`
	Link         *string `json:"link"`
}

// GetDocuments handles request get all documents.
func GetDocuments(ctx iris.Context) {
	opID, err := ctx.Params().GetInt("opID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	op, db := models.PhysicalOp{}, ctx.Values().Get("db").(*gorm.DB)
	if err = db.First(&op, opID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonMessage{"Liste des documents : opération introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	doo := []models.Document{}
	if err := db.Where("physical_op_id = ?", opID).Find(&doo).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(dooResp{doo})
}

// CreateDocument handles request post request to create a new document.
func CreateDocument(ctx iris.Context) {
	opID, err := ctx.Params().GetInt("opID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	op, db := models.PhysicalOp{}, ctx.Values().Get("db").(*gorm.DB)
	if err = db.First(&op, opID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonMessage{"Création de document : opération introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	do := doReq{}
	if err := ctx.ReadJSON(&do); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if do.Name == nil || len(*do.Name) == 0 || len(*do.Name) > 255 ||
		do.Link == nil || len(*do.Link) == 0 || len(*do.Link) > 255 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création de document, champ manquant ou incorrect"})
		return
	}

	newCo := models.Document{PhysicalOpID: opID, Name: *do.Name, Link: *do.Link}
	if err := db.Create(&newCo).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(doResp{newCo})
}

// ModifyDocument handles request put requestion to modify a document.
func ModifyDocument(ctx iris.Context) {
	doID, err := ctx.Params().GetInt("doID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	do, db := models.Document{}, ctx.Values().Get("db").(*gorm.DB)

	if err = db.First(&do, doID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonMessage{"Modification de document : introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	req := doReq{}
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if req.Name != nil && len(*req.Name) > 0 && len(*req.Name) < 255 {
		do.Name = *req.Name
	}

	if req.Link != nil && len(*req.Link) > 0 && len(*req.Link) < 255 {
		do.Link = *req.Link
	}

	if err = db.Save(&do).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(doResp{do})
}

// DeleteDocument handles the request to delete an document.
func DeleteDocument(ctx iris.Context) {
	doID, err := ctx.Params().GetInt("doID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	do, db := models.Document{ID: doID}, ctx.Values().Get("db").(*gorm.DB)

	if err = db.First(&do, doID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(jsonError{"Suppression de document : introuvable"})
		return
	}

	if err = db.Delete(&do).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Document supprimé"})
}
