package actions

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Iledant/iris_propera/models"
	"github.com/kataras/iris"
)

// GetScenarios handles get scenarios request.
func GetScenarios(ctx iris.Context) {
	db := ctx.Values().Get("db").(*sql.DB)
	var resp models.Scenarios
	if err := resp.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des scénarios, requête :" + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

type scenarioResp struct {
	Scenario models.Scenario `json:"Scenario"`
}

// CreateScenario handles put request to create a new scenario.
func CreateScenario(ctx iris.Context) {
	var req models.Scenario
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de scénario, décodage : " + err.Error()})
		return
	}
	if req.Invalid() {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'un scénario : mauvais format"})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := req.Create(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'un scénario, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(scenarioResp{req})
}

// ModifyScenario handles post request to modify an existing scenario.
func ModifyScenario(ctx iris.Context) {
	sID, err := ctx.Params().GetInt64("sID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de scénario, paramètre : " + err.Error()})
		return
	}
	var req models.Scenario
	db := ctx.Values().Get("db").(*sql.DB)
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de scénario, décodage : " + err.Error()})
		return
	}
	if req.Invalid() {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de scénario : mauvais format "})
		return
	}
	req.ID = sID
	if err = req.Update(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de scénario, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(scenarioResp{req})
}

// DeleteScenario handles delete request for a scenario.
func DeleteScenario(ctx iris.Context) {
	sID, err := ctx.Params().GetInt64("sID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression de scénario, paramètre : " + err.Error()})
		return
	}
	s := models.Scenario{ID: sID}
	db := ctx.Values().Get("db").(*sql.DB)
	if err = s.Delete(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression de scénario, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Scenario supprimé"})
}

// GetScenarioDatas handle the get request to get all offsets attached to a given scenario.
func GetScenarioDatas(ctx iris.Context) {
	sID, err := ctx.Params().GetInt64("sID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Datas d'un scénario, paramètre sID : " + err.Error()})
		return
	}
	firstYear, err := ctx.URLParamInt64("firstYear")
	if err != nil {
		firstYear = int64(time.Now().Year())
	}
	db := ctx.Values().Get("db").(*sql.DB)
	var resp models.ScenarioDatas
	if err = resp.Populate(sID, firstYear, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Datas d'un scénario, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// offsetReq handles an item of offset array sent.
type offsetReq struct {
	OperationID int64 `json:"physical_op_id"`
	Offset      int64 `json:"offset"`
}

// offsetsReq embeddes an array of offsetReq
type offsetsReq struct {
	OffsetReq []offsetReq `json:"offsetList"`
}

// SetScenarioOffsets handle the post request to set the offsets of a scenario.
func SetScenarioOffsets(ctx iris.Context) {
	sID, err := ctx.Params().GetInt64("sID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Offsets de scénario, paramètre : " + err.Error()})
		return
	}
	var req models.ScenarioOffsets
	db := ctx.Values().Get("db").(*sql.DB)
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Offsets de scénario, décodage : " + err.Error()})
		return
	}
	if err = req.Save(sID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Offsets de scénario, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Offsets créés"})
}

// GetScenarioActionPayments handles the get request to calculate the
// multiannual payment previsions of a scenario.
func GetScenarioActionPayments(ctx iris.Context) {
	sID, err := ctx.Params().GetInt64("sID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévisions de payment de scénario, paramètre : " + err.Error()})
		return
	}
	firstYear, err := ctx.URLParamInt64("FirstYear")
	if err != nil {
		firstYear = int64(time.Now().Year() + 1)
	}
	ptID, err := ctx.URLParamInt64("DefaultPaymentTypeId")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévisions de payment de scénario, paramètre : " + err.Error()})
		return
	}
	var resp models.ScenarioActionPayments
	db := ctx.Values().Get("db").(*sql.DB)
	if err = resp.GetAll(firstYear, sID, ptID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévisions de payment de scénario, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetScenarioStatActionPayments handles the get request to calculate the
// multiannual payment previsions of a scenario.
func GetScenarioStatActionPayments(ctx iris.Context) {
	sID, err := ctx.Params().GetInt64("sID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévisions statistique de payment de scénario, paramètre : " +
			err.Error()})
		return
	}
	firstYear, err := ctx.URLParamInt64("FirstYear")
	if err != nil {
		firstYear = int64(time.Now().Year() + 1)
	}
	ptID, err := ctx.URLParamInt64("DefaultPaymentTypeId")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévisions statistique de payment de scénario, paramètre : " +
			err.Error()})
		return
	}
	var resp models.ScenarioStatActionPayments
	db := ctx.Values().Get("db").(*sql.DB)
	if err = resp.GetAll(firstYear, sID, ptID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévisions statistique de payment de scénario, requête : " +
			err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetMultiAnnualScenario handles the get request to calculate the
// multiannual commitments prevision par budget entities of a scenario.
func GetMultiAnnualScenario(ctx iris.Context) {
	sID, err := ctx.Params().GetInt64("sID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision budgétaire pluriannuelle de scénario, paramètre : " +
			err.Error()})
		return
	}
	firstYear, err := ctx.URLParamInt64("firstYear")
	if err != nil {
		firstYear = int64(time.Now().Year() + 1)
	}
	var resp models.MultiAnnualBudgetScenario
	db := ctx.Values().Get("db").(*sql.DB)
	if err = resp.GetAll(firstYear, sID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision budgétaire pluriannuelle de scénario, requête : " +
			err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
