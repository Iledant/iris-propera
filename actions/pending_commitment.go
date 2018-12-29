package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// GetPendings handles the get request to fetch all pending commitments.
func GetPendings(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	var resp models.PendingCommitments
	if err := resp.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des engagements en cours : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetUnlinkedPendings handles the get request to fetch all pending commitments.
func GetUnlinkedPendings(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	var resp models.UnlinkedPendingCommitments
	if err := resp.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des engagements en cours non liés : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetLinkedPendings handles the get request to fetch all pending commitments.
func GetLinkedPendings(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	var resp models.PendingCommitments
	if err := resp.GetAllLinked(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des engagements en cours non liés : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// LinkPcToOp handles the post request to link an array of pending commitments to a physical operation.
func LinkPcToOp(ctx iris.Context) {
	opID, err := ctx.Params().GetInt64("opID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Rattachement d'engagement en cours, paramètre : " + err.Error()})
		return
	}
	var req models.PendingIDs
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Rattachement d'engagement en cours, décodage : " + err.Error()})
		return
	}
	op, db := models.PhysicalOp{ID: opID}, ctx.Values().Get("db").(*gorm.DB)
	if err = op.LinkPendings(&req, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Rattachement d'engagement en cours, requête : " + err.Error()})
		return
	}
	GetUnlinkedPendings(ctx)
}

// UnlinkPCs handles the post request to remove link between an array of pending commitments and physical operations.
func UnlinkPCs(ctx iris.Context) {
	var req models.PendingIDs
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Détachement d'engagement en cours, décodage : " + err.Error()})
		return
	}
	var p models.PendingCommitments
	if err := p.Unlink(&req, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Détachement d'engagement en cours, requête : " + err.Error()})
		return
	}
	GetLinkedPendings(ctx)
}

// BatchPendings handle the post request of an array of pendings commitments extracted from IRIS.
func BatchPendings(ctx iris.Context) {
	var req models.PendingsBatch
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch d'engagements en cours, décodage : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := req.Save(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch d'engagements en cours, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Engagements en cours importés"})
}
