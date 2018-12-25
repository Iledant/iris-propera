package actions

import (
	"net/http"
	"time"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// fullOpResp embeddes an operation with plan and action for json export.
type fullOpResp struct {
	FullOp models.OpWithPlanAndAction `json:"PhysicalOp"`
}

// GetPhysicalOps handles physical operations get request.It returns all operations with plan name and action name
// for admin and observer all operations are returned, for users only operations on which the user have rights
func GetPhysicalOps(ctx iris.Context) {
	uID, err := getUserID(ctx)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Opérations avec information, user : " + err.Error()})
		return
	}
	var resp models.OpWithPlanAndActions
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.GetAll(uID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Opérations avec information, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// CreatePhysicalOp handles physical operation create request.
func CreatePhysicalOp(ctx iris.Context) {
	var op models.PhysicalOp
	if err := ctx.ReadJSON(&op); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'une opération, décodage : " + err.Error()})
		return
	}
	if err := op.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'opération : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := op.Create(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'opération, requête : " + err.Error()})
		return
	}
	var resp fullOpResp
	resp.FullOp.ID = op.ID
	if err := resp.FullOp.Get(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'opération, requête get : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// DeletePhysicalOp handles physical operation delete request.
func DeletePhysicalOp(ctx iris.Context) {
	opID, err := ctx.Params().GetInt64("opID")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Suppression d'opération, décodage : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	op := models.PhysicalOp{ID: opID}
	if err = op.Delete(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'opération, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Opération supprimée"})
}

// UpdatePhysicalOp handles physical operation put request.
func UpdatePhysicalOp(ctx iris.Context) {
	opID, err := ctx.Params().GetInt64("opID")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Modification d'opération, paramètre : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	var op models.PhysicalOp
	if err := ctx.ReadJSON(&op); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'opération, décodage : " + err.Error()})
		return
	}
	if err = op.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Modification d'opération : " + err.Error()})
		return
	}
	uID, err := getUserID(ctx)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'opération, user : " + err.Error()})
		return
	}
	op.ID = opID
	if err = op.Update(uID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'opération, requête : " + err.Error()})
		return
	}
	var resp fullOpResp
	resp.FullOp.ID = op.ID
	if err = resp.FullOp.Get(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'opération, requête get : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// BatchPhysicalOps handles the request sending an array of physical operations.
func BatchPhysicalOps(ctx iris.Context) {
	var req models.PhysicalOpsBatch
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch opération, décodage : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := req.Save(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch opération, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Terminé"})
}

// getPrevisionsResp embeddes all datas for the physical operation's previsions
type getPrevisionsResp struct {
	models.PrevCommitments
	models.PrevPayments
	models.OpCommitments
	models.OpPendings
	models.OpPayments
	models.PaymentsPerBeneficiary
	models.FCsPerBeneficiary
	models.ImportLogs
	models.Events
	models.PaymentTypes
	models.Documents
}

// GetOpPrevisions handles the get request to fetch commitments and payments prevision for a physical operation.
func GetOpPrevisions(ctx iris.Context) {
	opID, err := ctx.Params().GetInt64("opID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, paramètre : " + err.Error()})
		return
	}
	year, err := ctx.URLParamInt64("year")
	if err != nil {
		year = int64(time.Now().Year())
	}
	op, db := models.PhysicalOp{ID: opID}, ctx.Values().Get("db").(*gorm.DB)
	if err = op.Exists(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, check : " + err.Error()})
		return
	}
	var resp getPrevisionsResp
	if err = op.GetYearPrevCommitments(&resp.PrevCommitments, year, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, requête prévision engagements : " + err.Error()})
		return
	}
	if err = op.GetYearPrevPayments(&resp.PrevPayments, year, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, requête prévision paiements : " + err.Error()})
		return
	}
	if err = resp.OpCommitments.GetOpAll(opID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, requête engagements : " + err.Error()})
		return
	}
	if resp.OpPendings, err = op.GetOpPendings(year, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, requête pending : " + err.Error()})
		return
	}
	if err = resp.OpPayments.GetOpAll(op.ID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, requête payment : " + err.Error()})
		return
	}
	if resp.PaymentsPerBeneficiary.GetOpAll(opID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, requête payment par bénéficiaire : " + err.Error()})
		return
	}
	if resp.FCsPerBeneficiary.GetOpAll(opID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, requête engagement par bénéficiaire : " + err.Error()})
		return
	}
	if err = resp.ImportLogs.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, requête import logs : " + err.Error()})
		return
	}
	if err = resp.Events.GetOpAll(opID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, requête get événements : " + err.Error()})
		return
	}
	if err = resp.Documents.GetOpAll(opID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, requête get documents : " + err.Error()})
		return
	}
	if err = resp.PaymentTypes.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, requête get payment types : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// setOpPrevResp embeddes arrays of financial commitments and payments previsions
type setOpPrevResp struct {
	models.PrevCommitments
	models.PrevPayments
}

// SetOpPrevisions handles the post request to set financial commitments and payments previsions
func SetOpPrevisions(ctx iris.Context) {
	opID, err := ctx.Params().GetInt64("opID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation prévision d'opération, erreur décodage identificateur : " + err.Error()})
		return
	}
	var req models.OpPrevisions
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation prévision d'opération, erreur décodage payload : " + err.Error()})
		return
	}
	op := models.PhysicalOp{ID: opID}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = op.Exists(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation prévision d'opération, opération : " + err.Error()})
		return
	}
	if err = op.SetPrevisions(&req, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation prévision d'opération, requête : " + err.Error()})
		return
	}
	var resp setOpPrevResp
	if err = op.GetPrevCommitments(&resp.PrevCommitments, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation prévision d'opération, requête get prévision engagements : " + err.Error()})
		return
	}
	if err = op.GetPrevPayments(&resp.PrevPayments, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation prévision d'opération, requête get prévision paiements : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetOpsAndFCs handle the get request to get all linked and unlinked operations and financial commitments
func GetOpsAndFCs(ctx iris.Context) {
	var resp models.OpAndCommitments
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := resp.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liens opérations engagement, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
