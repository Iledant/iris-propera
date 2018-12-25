package actions

import (
	"errors"
	"net/http"
	"time"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// reqFcIds is used to decode sent ids for attaching financial commitment.
type reqFcIDs struct {
	IDs []int64 `json:"fcIdList"`
}

// parseParams fetch params for linked or unlinked requests.
func parseParams(ctx iris.Context) (f models.FCSearchPattern, err error) {
	f.Page, err = ctx.URLParamInt64("page")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		return f, errors.New("erreur page:" + err.Error())
	}
	f.SearchText = "%" + ctx.URLParam("search") + "%"
	f.LinkType = ctx.URLParam("LinkType")
	minYear, err := ctx.URLParamInt("MinYear")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		return f, errors.New("erreur sur MinYear :" + err.Error())
	}
	if f.LinkType != "PhysicalOp" && f.LinkType != "PlanLine" {
		ctx.StatusCode(http.StatusBadRequest)
		return f, errors.New("mauvais paramètre LinkType")
	}
	f.MinDate = time.Date(minYear, 1, 1, 0, 0, 0, 0, time.UTC)
	return f, nil
}

// GetUnlinkedFcs handles the request to get all financial commitments not linked to a physical operation or a plan line.
// It uses a Laravel paginated request and has parameters for searching
func GetUnlinkedFcs(ctx iris.Context) {
	pattern, err := parseParams(ctx)
	if err != nil {
		ctx.JSON(jsonError{"Engagements non liés : " + err.Error()})
		return
	}
	var resp models.PaginatedUnlinkedItems
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.GetUnlinked(pattern, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Engagements non liés, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

//GetMonthFC handles the request to get the amount of financial commitments each montant of a given year.
func GetMonthFC(ctx iris.Context) {
	year, err := ctx.URLParamInt("year")
	if err != nil || year == 0 {
		year = time.Now().Year()
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	var resp models.MonthCommitments
	if err = resp.GetAll(year, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Engagements par mois, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// LinkFcToOp handles the request to link an array of financial commitments to an physical operation.
func LinkFcToOp(ctx iris.Context) {
	opID, err := ctx.Params().GetInt64("opID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Rattachement engagements / opération, paramètres : " + err.Error()})
		return
	}
	var fcIDs reqFcIDs
	if err := ctx.ReadJSON(&fcIDs); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Rattachement engagements / opération, décodage : " + err.Error()})
		return
	}
	op, db := models.PhysicalOp{ID: opID}, ctx.Values().Get("db").(*gorm.DB)
	if err = op.LinkFinancialCommitments(fcIDs.IDs, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Rattachement engagements / opération, requête : " + err.Error()})
		return
	}
	pattern := models.FCSearchPattern{LinkType: "PhysicalOp", SearchText: "%", Page: 1}
	var resp models.PaginatedUnlinkedItems
	if err = resp.GetUnlinked(pattern, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Rattachement engagements / opération, requête get : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// LinkFcToPl handles the request to link an array of financial commitments to a plan line.
func LinkFcToPl(ctx iris.Context) {
	plID, err := ctx.Params().GetInt64("plID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Rattachement engagements / ligne de plan, paramètre : " + err.Error()})
		return
	}
	var fcIDs reqFcIDs
	if err := ctx.ReadJSON(&fcIDs); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Rattachement engagements / ligne de plan, décodage : " + err.Error()})
		return
	}
	pl, db := models.PlanLine{ID: plID}, ctx.Values().Get("db").(*gorm.DB)
	if err = pl.LinkFCs(fcIDs.IDs, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Rattachement engagements / ligne de plan, requête : " + err.Error()})
		return
	}
	pattern := models.FCSearchPattern{LinkType: "PlanLine", SearchText: "%", Page: 1}
	var resp models.PaginatedUnlinkedItems
	if err = resp.GetUnlinked(pattern, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Rattachement engagements / ligne de plan, requête get : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetLinkedFcs handles the request to get all financial commitments linked to a physical operation or a plan line.
// It uses a Laravel paginated request and has parameters for searching
func GetLinkedFcs(ctx iris.Context) {
	pattern, err := parseParams(ctx)
	if err != nil {
		ctx.JSON(jsonError{"Engagements non liés : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	if pattern.LinkType == "PhysicalOp" {
		var resp models.PaginatedOpLinkedItems
		if err = resp.GetLinked(pattern, db.DB()); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Engagement non liés à opération, requête : " + err.Error()})
			return
		}
		ctx.StatusCode(http.StatusOK)
		ctx.JSON(resp)
	} else {
		var resp models.PaginatedPlanLineLinkedItems
		if err = resp.GetLinked(pattern, db.DB()); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Engagement non liés à opération, requête : " + err.Error()})
			return
		}
		ctx.StatusCode(http.StatusOK)
		ctx.JSON(resp)
	}
}

// GetOpFcs handles the request to get all financial commitments linked to a physical operation.
func GetOpFcs(ctx iris.Context) {
	opID, err := ctx.Params().GetInt64("opID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Engagement d'une opération, paramètre : " + err.Error()})
		return
	}
	var resp models.FinancialCommitments
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.GetOpAll(opID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Engagement d'une opération, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

type unlinkFcsReq struct {
	LinkType string `json:"linkType"`
	reqFcIDs
}

// UnlinkFcs handles the requests to unset link between a financial commitment and a physical operation or plan line.
func UnlinkFcs(ctx iris.Context) {
	req, db := unlinkFcsReq{}, ctx.Values().Get("db").(*gorm.DB)
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Détachement d'engagements, décodage : " + err.Error()})
		return
	}
	var f models.FinancialCommitment
	if err := f.Unlink(req.LinkType, req.reqFcIDs.IDs, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Détachement d'engagements, requête : " + err.Error()})
		return
	}
	pattern := models.FCSearchPattern{LinkType: req.LinkType, SearchText: "%", Page: 1}
	if pattern.LinkType == "PhysicalOp" {
		var resp models.PaginatedOpLinkedItems
		if err := resp.GetLinked(pattern, db.DB()); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Détachement d'engagements, requête get : " + err.Error()})
			return
		}
		ctx.StatusCode(http.StatusOK)
		ctx.JSON(resp)
	} else {
		var resp models.PaginatedPlanLineLinkedItems
		if err := resp.GetLinked(pattern, db.DB()); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Détachement d'engagements, requête get : " + err.Error()})
			return
		}
		ctx.StatusCode(http.StatusOK)
		ctx.JSON(resp)
	}
}

// BatchFcs handles the post request with an array of financial commitments (IRIS import).
func BatchFcs(ctx iris.Context) {
	db, req := ctx.Values().Get("db").(*gorm.DB), models.FinancialCommitmentsBatch{}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch engagements, décodage : " + err.Error()})
		return
	}
	if err := req.Save(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch engagements, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Engagements importés et mis à jour"})
}

// BatchOpFcs handle the post request to link of an array of physical operations with financial commitments.
func BatchOpFcs(ctx iris.Context) {
	db, opFcs := ctx.Values().Get("db").(*gorm.DB), models.OpFCsBatch{}
	if err := ctx.ReadJSON(&opFcs); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch opérations / engagements, décodage : " + err.Error()})
		return
	}
	if err := opFcs.Save(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch opérations / engagements, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON("Rattachements importés et réalisés")
}
