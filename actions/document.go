package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// doResp embeddes response for an single document.
type doResp struct {
	Document models.Document `json:"Document"`
}

// GetDocuments handles request get all documents.
func GetDocuments(ctx iris.Context) {
	opID, err := ctx.Params().GetInt64("opID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Documents d'une opération, paramètre : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	var resp models.Documents
	if err = resp.GetOpAll(opID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Documents d'une opération, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// CreateDocument handles request post request to create a new document.
func CreateDocument(ctx iris.Context) {
	opID, err := ctx.Params().GetInt64("opID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'un document, paramètre : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	var req models.Document
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'un document, décodage : " + err.Error()})
		return
	}
	req.PhysicalOpID = opID
	if err = req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'un document : " + err.Error()})
		return
	}
	if err = req.Create(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'un document, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(doResp{req})
}

// ModifyDocument handles request put requestion to modify a document.
func ModifyDocument(ctx iris.Context) {
	doID, err := ctx.Params().GetInt64("doID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un document, paramètre : " + err.Error()})
		return
	}
	opID, err := ctx.Params().GetInt64("opID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un document, paramètre : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	var req models.Document
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un document, décodage : " + err.Error()})
		return
	}
	req.PhysicalOpID = opID
	if err = req.Validate(); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un document : " + err.Error()})
		return
	}
	req.ID = doID
	if err = req.Update(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un document, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(doResp{req})
}

// DeleteDocument handles the request to delete an document.
func DeleteDocument(ctx iris.Context) {
	doID, err := ctx.Params().GetInt64("doID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'un document, paramètre : " + err.Error()})
		return
	}
	do, db := models.Document{ID: doID}, ctx.Values().Get("db").(*gorm.DB)
	if err = do.Delete(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'un document, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Document supprimé"})
}
