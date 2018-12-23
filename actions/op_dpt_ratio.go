package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// opsWithDptRatiosResp embeddes datas for response.
type opsWithDptRatiosResp struct {
	models.OpWithDptRatios
	models.ProgrammingsYears
}

// GetOpWithDptRatios handles get operation with department ratios request.
func GetOpWithDptRatios(ctx iris.Context) {
	uID, err := getUserID(ctx)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des opérations avec ratio, user :" + err.Error()})
		return
	}
	var resp opsWithDptRatiosResp
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.OpWithDptRatios.GetAll(uID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des opérations avec ratio, requête ratios :" + err.Error()})
		return
	}
	if err = resp.ProgrammingsYears.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des opérations avec ratio, requête years :" + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// BatchOpDptRatios handles the post request to set all ratios.
func BatchOpDptRatios(ctx iris.Context) {
	uID, err := getUserID(ctx)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch ratios départements, user : " + err.Error()})
		return
	}
	var req models.OpDptRatioBatch
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch ratios départements, décodage : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = req.Save(uID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch ratios départements, requête :" + err.Error()})
		return
	}
	GetOpWithDptRatios(ctx)
}

// GetFCPerDpt handles the get request to calculate financial commitments per departments between two years.
func GetFCPerDpt(ctx iris.Context) {
	y0, err := ctx.URLParamInt("firstYear")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Engagements par départements, décodage firstYear : " + err.Error()})
		return
	}
	y1, err := ctx.URLParamInt("lastYear")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Engagements par départements, décodage lastYear : " + err.Error()})
		return
	}
	if y1 < y0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Engagements par départements, dernière année plus petite que la première"})
		return
	}
	var resp models.FCPerDepartments
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.GetAll(y0, y1, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Engagements par départements, requête : " + err.Error()})
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetDetailedFCPerDpt handles the get request to calculate financial commitments per departments
// between two years and give repartition per operation
func GetDetailedFCPerDpt(ctx iris.Context) {
	y0, err := ctx.URLParamInt("firstYear")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Engagements détaillés par départements, décodage firstYear : " + err.Error()})
		return
	}
	y1, err := ctx.URLParamInt("lastYear")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Engagements détaillés par départements, décodage lastYear : " + err.Error()})
		return
	}
	if y1 < y0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Engagements détaillés par départements, dernière année plus petite que la première"})
		return
	}
	var resp models.DetailedFCPerDepartments
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.GetAll(y0, y1, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Engagements détaillés par départements, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetDetailedPrgPerDpt handles the get request to calculate programmings per departments
func GetDetailedPrgPerDpt(ctx iris.Context) {
	y, err := ctx.URLParamInt("year")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Programmation par départements, décodage year : " + err.Error()})
		return
	}
	var resp models.DetailedPrgPerDepartments
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.GetAll(y, db.DB()); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Programmation par départements, select : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
