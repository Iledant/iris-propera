package actions

import (
	"net/http"
	"time"

	"github.com/Iledant/iris_propera/actions/queries"
	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// pccResp embeddes response for an array of pendings commitments.
type pccResp struct {
	Pcc []models.PendingCommitment `json:"PendingCommitments"`
}

// GetPendings handles the get request to fetch all pending commitments.
func GetPendings(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	pcc := pccResp{}

	if err := db.Find(&pcc.Pcc).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des engagements en cours : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(pcc)
}

// GetUnlinkedPendings handles the get request to fetch all pending commitments.
func GetUnlinkedPendings(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	pcc := pccResp{}

	if err := db.Where("physical_op_id ISNULL").Find(&pcc.Pcc).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des engagements en cours non liés : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(pcc)
}

// GetLinkedPendings handles the get request to fetch all pending commitments.
func GetLinkedPendings(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	pcc := pccResp{}

	if err := db.Where("physical_op_id NOTNULL").Find(&pcc.Pcc).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des engagements en cours non liés : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(pcc)
}

// pcIDsREq is used to decode sent datas for LinkPcToOp
type pcIDsReq struct {
	IDs []int64 `json:"peIdList"`
}

// LinkPcToOp handles the post request to link an array of pending commitments to a physical operation.
func LinkPcToOp(ctx iris.Context) {
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
			ctx.JSON(jsonError{"Rattachement d'engagement en cours : opération introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
	}

	pcIDsReq := pcIDsReq{}
	if err = ctx.ReadJSON(&pcIDsReq); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Rattachement d'engagement en cours, erreur sur les identifiants : " + err.Error()})
		return
	}
	tx := db.Begin()

	count := struct{ Count int }{}
	if err = tx.Raw("select count(id) from pending_commitments where id in (?)", pcIDsReq.IDs).Scan(&count).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Rattachement d'engagement en cours, erreur: " + err.Error()})
		tx.Rollback()
		return
	}

	if count.Count != len(pcIDsReq.IDs) {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Rattachement d'engagement en cours, identifiant introuvable"})
		tx.Rollback()
		return
	}

	if err := tx.Exec("UPDATE pending_commitments SET physical_op_id = ? WHERE id IN (?)", opID, pcIDsReq.IDs).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonMessage{err.Error()})
		tx.Rollback()
		return
	}
	tx.Commit()

	GetUnlinkedPendings(ctx)
}

// UnlinkPCs handles the post request to remove link between an array of pending commitments and physical operations.
func UnlinkPCs(ctx iris.Context) {
	pcIDsReq, db := pcIDsReq{}, ctx.Values().Get("db").(*gorm.DB)

	if err := ctx.ReadJSON(&pcIDsReq); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Détachement d'engagement en cours, erreur sur les identifiants : " + err.Error()})
		return
	}

	tx, count := db.Begin(), struct{ Count int }{}

	if err := tx.Raw("select count(id) from pending_commitments where id in (?)", pcIDsReq.IDs).Scan(&count).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Détachement d'engagement en cours, erreur: " + err.Error()})
		tx.Rollback()
		return
	}

	if count.Count != len(pcIDsReq.IDs) {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Détachement d'engagement en cours, identifiant introuvable"})
		tx.Rollback()
		return
	}

	if err := tx.Exec("UPDATE pending_commitments SET physical_op_id = NULL WHERE id IN (?)", pcIDsReq.IDs).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonMessage{err.Error()})
		tx.Rollback()
		return
	}
	tx.Commit()

	GetLinkedPendings(ctx)
}

// batchPending is used to decode a row of array of a batch of pending commitments.
type batchPending struct {
	Chapter        string    `json:"chapter"`
	Action         string    `json:"action"`
	IrisCode       string    `json:"iris_code"`
	Name           string    `json:"name"`
	Beneficiary    string    `json:"beneficiary"`
	CommissionDate time.Time `json:"commission_date"`
	ProposedValue  int64     `json:"proposed_value"`
}

type batchPendings struct {
	Pcs []batchPending `json:"PendingCommitment"`
}

// BatchPendings handle the post request of an array of pendings commitments extracted from IRIS.
func BatchPendings(ctx iris.Context) {
	req := batchPendings{}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch d'engagements en cours, erreur de lecture : " + err.Error()})
		return
	}

	db := ctx.Values().Get("db").(*gorm.DB)
	tx := db.Begin()

	if err := tx.Exec(queries.DeletePendingTempTable).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch d'engagements en cours, suppression table temporaire : " + err.Error()})
		tx.Rollback()
		return
	}

	if err := tx.Exec(queries.CreatePendingTempTable).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch d'engagements en cours, création table temporaire : " + err.Error()})
		tx.Rollback()
		return
	}

	for _, pc := range req.Pcs {
		if err := tx.Exec(queries.InsertBatchPending, pc.Chapter, pc.Action, pc.IrisCode, pc.Name, pc.Beneficiary, pc.CommissionDate, pc.ProposedValue).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Batch d'engagements en cours, insertion : " + err.Error()})
			tx.Rollback()
			return
		}
	}

	queries := []string{queries.UpdatePendingWithBatch, queries.InsertPendingWithBatch,
		queries.DeletePendingOutOfBatch, queries.DeletePendingTempTable}

	for _, qry := range queries {
		if err := tx.Exec(qry).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Batch d'engagements en cours, requête : " + err.Error()})
			tx.Rollback()
			return
		}
	}

	tx.Commit()
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Engagements en cours importés"})
}
