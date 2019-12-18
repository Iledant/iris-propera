package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testScenario(t *testing.T) {
	t.Run("Scenario", func(t *testing.T) {
		getScenarioTest(testCtx.E, t)
		ID := createScenarioTest(testCtx.E, t)
		if ID == 0 {
			t.Fatal("Impossible de créer le scénario")
		}
		modifyScenarioTest(testCtx.E, t, ID)
		getScenarioDatasTest(testCtx.E, t, ID)
		getMultiannualBudgetScenarioTest(testCtx.E, t, ID)
		setScenarioOffsetsText(testCtx.E, t, ID)
		getScenarioActionPaymentTest(testCtx.E, t, ID)
		getScenarioStatActionPaymentTest(testCtx.E, t, ID)
		deleteScenarioTest(testCtx.E, t, ID)
	})
}

// getScenarioTest check route is protected and datas sent has got items and number of lines.
func getScenarioTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusOK,
			BodyContains: []string{"Scenario", "name", "descript", "Scénario 750 M€ pour 2018"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/scenarios").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetScenarios") {
		t.Error(r)
	}
}

// createScenarioTest check route is protected and created scenarios works properly.
func createScenarioTest(e *httpexpect.Expect, t *testing.T) (ID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusCreated,
			IDName: `"id"`,
			Sent:   []byte(`{"name":"Scénario créé","descript":"Description du scénario créé"}`),
			BodyContains: []string{"Scenario", "name", "Scénario créé", "descript",
				"Description du scénario créé"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/scenarios").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "CreateScenario", &ID) {
		t.Error(r)
	}
	return ID
}

// getScenarioActionPaymentTest check route is protected and created scenarios works properly.
func getScenarioActionPaymentTest(e *httpexpect.Expect, t *testing.T, ID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusOK,
			ID:     strconv.Itoa(ID),
			BodyContains: []string{"ScenarioPaymentPerBudgetAction",
				`"chapter":"908","sector":"TC","subfunction":"811","program":"281005",` +
					`"action":"2810050101","action_name":"Liaisons tramways",` +
					`"y1":2767866.8312180256,"y2":-327128.643554943,"y3":-766460.1789780867`},
			CountItemName: `"chapter"`,
			ArraySize:     53},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/scenarios/"+tc.ID+"/payment_per_budget_action").
			WithHeader("Authorization", "Bearer "+tc.Token).WithQuery("FirstYear", 2018).
			WithQuery("DefaultPaymentTypeId", 5).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetScenarioActionPayment") {
		t.Error(r)
	}
}

// getScenarioStatActionPaymentTest check route is protected and created scenarios works properly.
func getScenarioStatActionPaymentTest(e *httpexpect.Expect, t *testing.T, ID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusOK,
			ID:     strconv.Itoa(ID),
			BodyContains: []string{"ScenarioStatisticalPaymentPerBudgetAction",
				`"chapter":"908","sector":"TC","subfunction":"811","program":"281005",` +
					`"action":"2810050101","action_name":"Liaisons tramways",` +
					`"y1":2767866.8312180256,"y2":-327128.643554943,"y3":-766460.1789780867`},
			CountItemName: `"chapter"`,
			ArraySize:     53},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/scenarios/"+tc.ID+"/statistical_payment_per_budget_action").
			WithHeader("Authorization", "Bearer "+tc.Token).WithQuery("FirstYear", 2018).
			WithQuery("DefaultPaymentTypeId", 5).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetScenarioStatActionPayment") {
		t.Error(r)
	}
}

// modifyScenarioTest check route is protected and modify works properly.
func modifyScenarioTest(e *httpexpect.Expect, t *testing.T, ID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			ID:           "0",
			Sent:         []byte(`{"name":"Scénario modifié","descript":"Description du scénario modifié"}`),
			BodyContains: []string{"Modification de scénario, requête : Scenario introuvable"}},
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusOK,
			ID:     strconv.Itoa(ID),
			Sent:   []byte(`{"name":"Scénario modifié","descript":"Description du scénario modifié"}`),
			BodyContains: []string{"Scenario", "name", "Scénario modifié", "descript",
				"Description du scénario modifié"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.PUT("/api/scenarios/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "ModifyScenario") {
		t.Error(r)
	}
}

// deleteScenarioTest check route is protected and delete works properly.
func deleteScenarioTest(e *httpexpect.Expect, t *testing.T, ID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			ID:           "0",
			BodyContains: []string{"Suppression de scénario, requête : Scenario introuvable"}},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusOK,
			ID:           strconv.Itoa(ID),
			BodyContains: []string{"Scenario supprimé"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.DELETE("/api/scenarios/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DeleteScenario") {
		t.Error(r)
	}
}

// getScenarioDatasTest check route is protected and correct objects sent back.
func getScenarioDatasTest(e *httpexpect.Expect, t *testing.T, ID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusOK,
			ID:           "1",
			Param:        "2018",
			BodyContains: []string{"OperationCrossTable", "ScenarioCrossTable"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/scenarios/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetScenarioDatas") {
		t.Error(r)
	}
}

// getMultiannualBudgetScenarioTest check route is protected and correct objects sent back.
func getMultiannualBudgetScenarioTest(e *httpexpect.Expect, t *testing.T, ID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusOK,
			ID:           "0",
			Param:        "2018",
			BodyContains: []string{`"MultiannualBudgetScenario":[]`}},
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusOK,
			ID:     "1",
			Param:  "2018",
			BodyContains: []string{"MultiannualBudgetScenario",
				// cSpell:disable
				`"MultiannualBudgetScenario":[{"number":"01BU003","name":"Bus - Tzen5 -` +
					` Paris-Choisy (94)","chapter":908,"sector":"MO","subfunction":"818",` +
					`"program":"481015","action":"481015011","y0":0,"y1":1274000000,` +
					`"y2":4047400000,"y3":0,"y4":0}]`}},
		// cSpell:enable
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/scenarios/"+tc.ID+"/budget").WithQuery("firstYear", tc.Param).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetMultiannualBudgetScenario") {
		t.Error(r)
	}
}

// setScenarioOffsetsText check route is protected and offset add return ok.
func setScenarioOffsetsText(e *httpexpect.Expect, t *testing.T, ID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusInternalServerError,
			ID:     "0",
			Sent: []byte(`{"offsetList":[{"physical_op_id":220,"offset":0},{"physical_op_id":546,"offset":1},
			{"physical_op_id":9,"offset":0},{"physical_op_id":543,"offset":2}]}`),
			BodyContains: []string{"Offsets de scénario, requête : pq:"}},
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusOK,
			ID:     strconv.Itoa(ID),
			Sent: []byte(`{"offsetList":[{"physical_op_id":220,"offset":0},{"physical_op_id":546,"offset":1},
			{"physical_op_id":9,"offset":0},{"physical_op_id":543,"offset":2}]}`),
			BodyContains: []string{"Offsets créés"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/scenarios/"+tc.ID+"/offsets").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "SetScenarioOffsets") {
		t.Error(r)
	}
}
