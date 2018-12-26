package actions

import (
	"net/http"
	"time"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// GetMultiannualProg handles theget request to fetch multiannual programmation.
func GetMultiannualProg(ctx iris.Context) {
	var resp models.MultiannualProg
	y1, err := ctx.URLParamInt64("y1")
	if err != nil {
		y1 = int64(time.Now().Year())
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.GetAll(y1, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Programmation pluriannuelle, requête : " + err.Error()})
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// annualProgResp embeddes an array of annualProg for the annual programmation response.
type annualProgResp struct {
	models.AnnualProgrammation
	models.ImportLogs
}

// GetAnnualProgrammation handles the get request to fetch datas comparing
// programmation, commitments and pending commitments.
func GetAnnualProgrammation(ctx iris.Context) {
	year, err := ctx.URLParamInt("year")
	if err != nil {
		year = time.Now().Year()
	}
	var resp annualProgResp
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.AnnualProgrammation.GetAll(year, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Programmation annuelle, requête : " + err.Error()})
		return
	}
	if err = resp.ImportLogs.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Programmation annuelle, import logs : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetProgrammingAndPrevisions handles the get request to compare precisely
// programmation and previsions.
func GetProgrammingAndPrevisions(ctx iris.Context) {
	year, err := ctx.URLParamInt64("y1")
	if err != nil {
		year = int64(time.Now().Year())
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	var resp models.ProgrammingAndPrevisions
	if err = resp.GetAll(year, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévisions et programmation, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetActionProgrammation handles the get request to fetch the programmation by budget actions.
func GetActionProgrammation(ctx iris.Context) {
	year, err := ctx.URLParamInt64("y1")
	if err != nil {
		year = int64(time.Now().Year())
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	var resp models.ActionProgrammations
	if err = resp.GetAll(year, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Programmation par action, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetActionCommitment handles the get request to fetch prevision of payment by budget actions.
func GetActionCommitment(ctx iris.Context) {
	y1, err := ctx.URLParamInt64("FirstYear")
	if err != nil {
		y1 = int64(time.Now().Year()) + 1
	}
	var resp models.ActionCommitments
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.GetAll(y1, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévisions AP par actions budgétaires, requête : " + err.Error()})
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetDetailedActionCommitment handles the get request to have detailed
// commitment per budget actions.
func GetDetailedActionCommitment(ctx iris.Context) {
	y1, err := ctx.URLParamInt64("FirstYear")
	if err != nil {
		y1 = int64(time.Now().Year()) + 1
	}
	var resp models.DetailedActionCommitments
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.GetAll(y1, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévisions AP détaillées par actions budgétaires, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetDetailedActionPayment handles the get request to get payment prevision
// per physical operation using payment prevision if available, statistical
// approach otherwise.
func GetDetailedActionPayment(ctx iris.Context) {
	y1, err := ctx.URLParamInt64("FirstYear")
	if err != nil {
		y1 = int64(time.Now().Year()) + 1
	}
	dID, err := ctx.URLParamInt64("DefaultPaymentTypeId")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiement détaillé par action, décodage : " + err.Error()})
		return
	}
	var resp models.DetailedActionPayments
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.GetAll(y1, dID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiement détaillé par action, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetStatDetailedActionPayment handles the get request to get payment prevision
// per physical operation using only statistical approach.
func GetStatDetailedActionPayment(ctx iris.Context) {
	y1, err := ctx.URLParamInt64("FirstYear")
	if err != nil {
		y1 = int64(time.Now().Year()) + 1
	}
	dID, err := ctx.URLParamInt64("DefaultPaymentTypeId")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiement détaillé par action, décodage : " + err.Error()})
		return
	}
	var resp models.StatDetailedActionPayments
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.GetAll(y1, dID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiement détaillé par action, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetActionPayment handles the get request to get payment prevision by budget action.
func GetActionPayment(ctx iris.Context) {
	y1, err := ctx.URLParamInt64("FirstYear")
	if err != nil {
		y1 = int64(time.Now().Year()) + 1
	}
	dID, err := ctx.URLParamInt64("DefaultPaymentTypeId")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiement par action, décodage : " + err.Error()})
		return
	}
	var resp models.ActionPayments
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.GetAll(y1, dID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiement par action, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetStatActionPayment handles the get request to get payment prevision by budget action.
func GetStatActionPayment(ctx iris.Context) {
	y1, err := ctx.URLParamInt64("FirstYear")
	if err != nil {
		y1 = int64(time.Now().Year()) + 1
	}
	dID, err := ctx.URLParamInt64("DefaultPaymentTypeId")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiement statistique par action, décodage : " + err.Error()})
		return
	}
	var resp models.ActionPayments
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.GetStatAll(y1, dID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiement statistique par action, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetStatCurrentYearPayment handles the get request to get payment prevision by budget action.
func GetStatCurrentYearPayment(ctx iris.Context) {
	y, err := ctx.URLParamInt64("Year")
	if err != nil {
		y = int64(time.Now().Year()) + 1
	}
	dID, err := ctx.URLParamInt64("DefaultPaymentTypeId")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision annuelle statistique, décodage : " + err.Error()})
		return
	}
	var resp models.CurrentYearPrevPayments
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.GetAll(y, dID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision annuelle statistique, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
