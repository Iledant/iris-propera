package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// settingsResp embeddes the different arrays for the get settings request
type settingsResp struct {
	models.Beneficiaries
	models.BudgetChapters
	models.BudgetSectors
	models.BudgetPrograms
	models.BudgetActions
	models.Commissions
	models.PhysicalOps
	models.PaymentTypes
	models.Plans
	models.CompleteBudgetCredits
	models.UnlinkedPendingCommitments
	models.CompletePendingCommitments
	models.Steps
	models.Categories
}

// GetSettings handle the get settings request that embeddes many arrays in juste one call
// to reduce the load time of the settings frontend page.
func getSettings(ctx iris.Context) {
	resp, db := settingsResp{}, ctx.Values().Get("db").(*gorm.DB)
	if err := resp.Beneficiaries.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings beneficiary : " + err.Error()})
		return
	}
	if err := resp.BudgetChapters.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings chapter : " + err.Error()})
		return
	}
	if err := resp.BudgetSectors.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings sector : " + err.Error()})
		return
	}
	if err := resp.BudgetPrograms.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings program : " + err.Error()})
		return
	}
	if err := resp.BudgetActions.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings action : " + err.Error()})
		return
	}
	if err := resp.Commissions.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings commission : " + err.Error()})
		return
	}
	if err := resp.PhysicalOps.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings physical operation : " + err.Error()})
		return
	}
	if err := resp.PaymentTypes.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings payment type : " + err.Error()})
		return
	}
	if err := resp.Plans.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings plan : " + err.Error()})
		return
	}
	if err := resp.CompleteBudgetCredits.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings budget credit : " + err.Error()})
		return
	}
	if err := resp.UnlinkedPendingCommitments.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings unlinked pendings : " + err.Error()})
		return
	}
	if err := resp.CompletePendingCommitments.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings linked pendings : " + err.Error()})
		return
	}
	if err := resp.Steps.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings step : " + err.Error()})
		return
	}
	if err := resp.Categories.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Settings category : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
