package actions

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

// todayMsgResp embeddes the today message for JSON response
type todayMsgResp struct {
	TodayMessage models.TodayMessage `json:"TodayMessage"`
}

// GetTodayMessage handles the get request to fetch title and text
func GetTodayMessage(ctx iris.Context) {
	var resp todayMsgResp
	var err error
	db := ctx.Values().Get("db").(*sql.DB)
	if err = resp.TodayMessage.Get(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Today message requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// SetTodayMessage handles the set request to fetch title and text
func SetTodayMessage(ctx iris.Context) {
	var req models.TodayMessage
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation today message, décodage : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := req.Update(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation today message, update : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(todayMsgResp{req})
}

type homeResp struct {
	models.TodayMessage `json:"TodayMessage"`
	models.NextMonthEvents
	models.MonthCommitments
	models.YearBudgetCredits
	models.ProgrammingsPerMonthes
	models.PaymentPerMonths
	models.ImportLogs
	models.PaymentCredits
	models.AvgPmtTimes
	models.PaymentDemandCounts
	models.PaymentDemandsStocks
	models.CsfWeekTrend    `json:"CsfWeekTrend"`
	models.FlowStockDelays `json:"FlowStockDelays"`
	models.PaymentRate     `json:"PaymentRate"`
}

// GetHomeDatas handles the get request for the home page.
func GetHomeDatas(ctx iris.Context) {
	var resp homeResp
	var err error
	db := ctx.Values().Get("db").(*sql.DB)
	if err = resp.TodayMessage.Get(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"HomeDatas, today messages requête : " + err.Error()})
		return
	}
	uID, err := getUserID(ctx)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Homedatas, user : " + err.Error()})
		return
	}
	if err = resp.NextMonthEvents.Get(uID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Homedatas, next month events : " + err.Error()})
		return
	}
	year, err := ctx.URLParamInt("year")
	if err != nil || year == 0 {
		year = time.Now().Year()
	}
	if err = resp.MonthCommitments.GetAll(year, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"HomeDatas, MonthCommitments : " + err.Error()})
		return
	}
	if err = resp.YearBudgetCredits.GetAll(year, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"HomeDatas, YearBudgetCredits : " + err.Error()})
		return
	}
	err = resp.ProgrammingsPerMonthes.GetAll(year, db)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"HomeDatas, ProgrammingsPerMonth : " + err.Error()})
		return
	}
	err = resp.PaymentPerMonths.GetAll(year, db)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"HomeDatas, PaymentPerMonths : " + err.Error()})
		return
	}
	if err = resp.ImportLogs.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"HomeDatas, ImportLogs : " + err.Error()})
		return
	}
	if err = resp.PaymentCredits.GetAll(year, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"HomeDatas, PaymentCredits : " + err.Error()})
		return
	}
	if err = resp.AvgPmtTimes.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"HomeDatas, AvgPmtTimes : " + err.Error()})
		return
	}
	if err = resp.PaymentDemandCounts.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"HomeDatas, PaymentDemandCount : " + err.Error()})
		return
	}
	if err = resp.PaymentDemandsStocks.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"HomeDatas, PaymentDemandsStock : " + err.Error()})
		return
	}
	if err = resp.CsfWeekTrend.Get(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"HomeDatas, CsfWeekTrend : " + err.Error()})
		return
	}
	if err = resp.FlowStockDelays.Get(90, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"HomeDatas, CsfWeekTrend : " + err.Error()})
		return
	}
	if err = resp.PaymentRate.Get(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"HomeDatas, PaymentRate : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
