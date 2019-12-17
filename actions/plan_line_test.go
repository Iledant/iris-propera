package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testPlanLine(t *testing.T) {
	t.Run("PlanLine", func(t *testing.T) {
		getPlanLinesTest(testCtx.E, t)
		getDetailedPlanLinesTest(testCtx.E, t)
		plID := createPlanLineTest(testCtx.E, t)
		if plID == 0 {
			t.Fatal("Impossible de créer la ligne de plan")
		}
		modifyPlanLineTest(testCtx.E, t, plID)
		deletePlanLineTest(testCtx.E, t, plID)
		batchPlanLinesTest(testCtx.E, t)
	})
}

// getPlanLinesTest check if route is protected and query sent correct datas.
func getPlanLinesTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token: testCtx.User.Token,
			ID:    "1", Status: http.StatusOK,
			BodyContains:  []string{"PlanLine", "Beneficiary"},
			CountItemName: `"total_value"`,
			ArraySize:     59},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/plans/"+tc.ID+"/planlines").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetPlanLines") {
		t.Error(r)
	}
}

// getDetailedPlanLinesTest check if route is protected and query sent correct datas.
func getDetailedPlanLinesTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Liste détaillée des lignes de plan, requête plan : "}},
		{
			Token:         testCtx.User.Token,
			ID:            "1",
			Status:        http.StatusOK,
			BodyContains:  []string{"DetailedPlanLine"},
			CountItemName: `"id"`,
			ArraySize:     380},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/plans/"+tc.ID+"/planlines/detailed").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetDetailedPlanLines") {
		t.Error(r)
	}
}

// createPlanLineTest check if route is protected and plan line sent back is correct.
func createPlanLineTest(e *httpexpect.Expect, t *testing.T) (plID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{Plu}`),
			BodyContains: []string{"Création de ligne de plan, requête plan : "}},
		{
			Token:        testCtx.Admin.Token,
			ID:           "1",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{Plu}`),
			BodyContains: []string{"Création de ligne de plan, décodage"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           "1",
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{"Descript":null}`),
			BodyContains: []string{"Création de ligne de plan, erreur de name"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           "1",
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{"name":"Essai de ligne de plan"}`),
			BodyContains: []string{"Création de ligne de plan, erreur de value"}},
		{
			Token:  testCtx.Admin.Token,
			ID:     "1",
			Status: http.StatusCreated,
			IDName: `"id"`,
			Sent: []byte(`{"name":"Essai de ligne de plan", "value":123,"total_value":456,
			"descript":"Essai de description","ratios":[{"ratio":0.5,"beneficiary_id":16}]}`),
			BodyContains: []string{"PlanLine", `"name":"Essai de ligne de plan"`, `"value":123`,
				`"total_value":456`, `"descript":"Essai de description"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/plans/"+tc.ID+"/planlines").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "CreatePlanLine", &plID) {
		t.Error(r)
	}
	return plID
}

// modifyPlanLineTest check if route is protected and plan line sent back is correct.
func modifyPlanLineTest(e *httpexpect.Expect, t *testing.T, plID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Param:        "0",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{Plu}`),
			BodyContains: []string{"Modification de ligne de plan, requête plan : "}},
		{
			Token:        testCtx.Admin.Token,
			ID:           "1",
			Param:        "0",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{Plu}`),
			BodyContains: []string{"Modification de ligne de plan, requête getByID : "}},
		{
			Token:        testCtx.Admin.Token,
			ID:           "1",
			Param:        strconv.Itoa(plID),
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{Plu}`),
			BodyContains: []string{"Modification de ligne de plan, décodage : "}},
		{
			Token:  testCtx.Admin.Token,
			ID:     "1",
			Param:  strconv.Itoa(plID),
			Status: http.StatusOK,
			Sent:   []byte(`{"name":"Modification de ligne de plan", "value":456,"total_value":789}`),
			BodyContains: []string{"PlanLine", `"name":"Modification de ligne de plan"`,
				`"value":456`, `"total_value":789`, `"descript":"Essai de description"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.PUT("/api/plans/"+tc.ID+"/planlines/"+tc.Param).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "ModifyPlanLine") {
		t.Error(r)
	}
}

// deletePlanLineTest check if route is protected and plan line sent back is correct.
func deletePlanLineTest(e *httpexpect.Expect, t *testing.T, plID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "1",
			Param:        "0",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Suppression de ligne de plan, requête : Ligne de plan introuvable"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           "1",
			Param:        strconv.Itoa(plID),
			Status:       http.StatusOK,
			BodyContains: []string{"Ligne de plan supprimée"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.DELETE("/api/plans/"+tc.ID+"/planlines/"+tc.Param).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DeletePlanLine") {
		t.Error(r)
	}
	testCases = []testCase{
		{
			Token:         testCtx.Admin.Token,
			Status:        http.StatusOK,
			ID:            "1",
			CountItemName: `"id"`,
			ArraySize:     60},
	}
	f = func(tc testCase) *httpexpect.Response {
		return e.GET("/api/plans/"+tc.ID+"/planlines").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DeletePlanLine") {
		t.Error(r)
	}
}

// batchPlanLinesTest check if route is protected and plan line sent back is correct.
func batchPlanLinesTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{Plu}`),
			BodyContains: []string{"Batch lignes de plan, décodage : "}},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"PlanLine":[{"value":100.5}]}`),
			BodyContains: []string{`Batch lignes de plan, requête : Colonne name manquante`}},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"PlanLine":[{"name":"Ligne batch1"}]}`),
			BodyContains: []string{`Batch lignes de plan, requête : Colonne value manquante`}},
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusOK,
			Sent: []byte(`{"PlanLine":[{"name":"Ligne batch1","value":100.5,"502":0.3,"16":0.2},
			{"name":"Ligne batch2","value":200,"descript":null,"total_value":400.5,"502":null},
			{"name":"Ligne batch3","value":300,"descript":null,"total_value":400,"502":0.15},
			{"name":"Ligne batch4","value":200,"descript":"Description lige batch3","total_value":null}]}`),
			BodyContains: []string{`Batch lignes de plan importé`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/plans/1/planlines/array").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "BatchPlanLines") {
		t.Error(r)
	}
}
