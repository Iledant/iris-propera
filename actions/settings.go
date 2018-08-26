package actions

import (
	"net/http"
	"time"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

type completeBC struct {
	ID                 int64     `json:"id" gorm:"column:id"`
	CommissionDate     time.Time `json:"commission_date" gorm:"column:commission_date"`
	Chapter            int64     `json:"chapter" gorm:"column:chapter"`
	PrimaryCommitment  int64     `json:"primary_commitment" gorm:"column:primary_commitment"`
	FrozenCommitment   int64     `json:"frozen_commitment" gorm:"column:frozen_commitment"`
	ReservedCommitment int64     `json:"reserved_commitment" gorm:"column:reserved_commitment"`
}

// settingsResp embeddes the different arrays for the get settings request
type settingsResp struct {
	Beneficiary                []models.Beneficiary       `json:"Beneficiary"`
	BudgetChapter              []models.BudgetChapter     `json:"BudgetChapter"`
	BudgetSector               []models.BudgetSector      `json:"BudgetSector"`
	BudgetProgram              []models.BudgetProgram     `json:"BudgetProgram"`
	BudgetAction               []models.BudgetAction      `json:"BudgetAction"`
	Commissions                []models.Commission        `json:"Commissions"`
	PhysicalOp                 []models.PhysicalOp        `json:"PhysicalOp"`
	PaymentType                []models.PaymentType       `json:"PaymentType"`
	Plan                       []models.Plan              `json:"Plan"`
	BudgetCredits              []completeBC               `json:"BudgetCredits"`
	UnlinkedPendingCommitments []models.PendingCommitment `json:"UnlinkedPendingCommitments"`
	LinkedPendingCommitments   []linkedPe                 `json:"LinkedPendingCommitments"`
	Step                       []models.Step              `json:"Step"`
	Category                   []models.Category          `json:"Category"`
}

func getCompleteBudgetCredits(db *gorm.DB, bcs *[]completeBC) error {
	rows, err := db.Raw(`SELECT bc.id, bc.commission_date, c.code AS chapter, 
	bc.primary_commitment, bc.frozen_commitment, bc.reserved_commitment
	FROM budget_credits bc, budget_chapter c
	WHERE bc.chapter_id = c.id`).Rows()
	if err != nil {
		return nil
	}
	defer rows.Close()
	bc := completeBC{}
	for rows.Next() {
		if err = db.ScanRows(rows, &bc); err != nil {
			return err
		}
		*bcs = append(*bcs, bc)
	}
	return nil
}

// linkedPe is used to decode linked pendings query
type linkedPe struct {
	ID            int64     `json:"id" gorm:"column:id"`
	PeName        string    `json:"peName" gorm:"column:peName"`
	PeIrisCode    string    `json:"peIrisCode" gorm:"column:peIrisCode"`
	PeDate        time.Time `json:"peDate" gorm:"column:peDate"`
	PeBeneficiary string    `json:"peBeneficiary" gorm:"column:peBeneficiary"`
	PeValue       int64     `json:"peValue" gorm:"column:peValue"`
	OpName        string    `json:"opName" gorm:"column:opName"`
}

func getLinkedPendings(db *gorm.DB, lpe *[]linkedPe) error {
	rows, err := db.Raw(`SELECT pe.id, pe.name AS peName, pe.iris_code AS peIrisCode, 
	pe.commission_date AS peDate, pe.beneficiary AS peBeneficiary,
	pe.proposed_value AS peValue, op.number || ' - ' || op.name AS opName 
FROM pending_commitments pe, physical_op op 
WHERE pe.physical_op_id = op.id`).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()
	pe := linkedPe{}
	for rows.Next() {
		if err := db.ScanRows(rows, &pe); err != nil {
			return err
		}
		*lpe = append(*lpe, pe)
	}
	return nil
}

// GetSettings handle the get settings request that embeddes many arrays in juste one call
// to reduce the load time of the settings frontend page.
func getSettings(ctx iris.Context) {
	resp, db := settingsResp{}, ctx.Values().Get("db").(*gorm.DB)
	if err := db.Find(&resp.Beneficiary).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings beneficiary : " + err.Error()})
		return
	}
	if err := db.Find(&resp.BudgetChapter).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings chapter : " + err.Error()})
		return
	}
	if err := db.Find(&resp.BudgetSector).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings sector : " + err.Error()})
		return
	}
	if err := db.Find(&resp.BudgetProgram).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings program : " + err.Error()})
		return
	}
	if err := db.Find(&resp.BudgetAction).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings action : " + err.Error()})
		return
	}
	if err := db.Find(&resp.Commissions).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings commission : " + err.Error()})
		return
	}
	if err := db.Find(&resp.PhysicalOp).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings physical operation : " + err.Error()})
		return
	}
	if err := db.Find(&resp.PaymentType).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings payment type : " + err.Error()})
		return
	}
	if err := db.Find(&resp.Plan).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings plan : " + err.Error()})
		return
	}
	if err := getCompleteBudgetCredits(db, &resp.BudgetCredits); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings budget credit : " + err.Error()})
		return
	}
	if err := db.Where("physical_op_id ISNULL").Find(&resp.UnlinkedPendingCommitments).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings unlinked pendings : " + err.Error()})
		return
	}
	if err := getLinkedPendings(db, &resp.LinkedPendingCommitments); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings linked pendings : " + err.Error()})
		return
	}
	if err := db.Find(&resp.Step).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings step : " + err.Error()})
		return
	}
	if err := db.Find(&resp.Category).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings category : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
