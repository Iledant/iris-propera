package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testPlan(t *testing.T) {
	t.Run("Plan", func(t *testing.T) {
		getPlansTest(testCtx.E, t)
		pID := createPlanTest(testCtx.E, t)
		if pID == 0 {
			t.Fatal("Impossible de créer le plan")
		}
		modifyPlanTest(testCtx.E, t, pID)
		deletePlanTest(testCtx.E, t, pID)
	})
}

// getPlansTest check if route is protected and query sent correct datas.
func getPlansTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:         testCtx.User.Token,
			Status:        http.StatusOK,
			CountItemName: `"id"`,
			ArraySize:     5,
			BodyContains:  []string{"Plan", "name", "descript", "first_year", "last_year"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/plans").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetPlans") {
		t.Error(r)
	}
}

// createPlanTest check if route is protected and plan sent back is correct.
func createPlanTest(e *httpexpect.Expect, t *testing.T) (pID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusInternalServerError,

			Sent:         []byte(`{Plu}`),
			BodyContains: []string{"Création de plan, décodage :"}},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{"Descript":null}`),
			BodyContains: []string{"Création d'un plan : Name incorrect"}},
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusCreated,
			IDName: `"id"`,
			Sent:   []byte(`{"name":"Essai de plan", "descript":"Essai de description","first_year":2015,"last_year":2025}`),
			BodyContains: []string{"Plan", `"name":"Essai de plan"`, `"first_year":2015`,
				`"last_year":2025`, `"descript":"Essai de description"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/plans").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "CreatePlan", &pID) {
		t.Error(r)
	}
	return pID
}

// modifyPlanTest check if route is protected and plan sent back is correct.
func modifyPlanTest(e *httpexpect.Expect, t *testing.T, pID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Modification de plan", "descript":"Modification de description","first_year":2016,"last_year":2024}`),
			BodyContains: []string{"Modification de plan, requête : Plan introuvable"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(pID),
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"Plu"}`),
			BodyContains: []string{"Modification de plan, décodage :"}},
		{
			Token:  testCtx.Admin.Token,
			ID:     strconv.Itoa(pID),
			Status: http.StatusOK,
			Sent:   []byte(`{"name":"Modification de plan", "descript":"Modification de description","first_year":2016,"last_year":2024}`),
			BodyContains: []string{"Plan", `"name":"Modification de plan"`, `"first_year":2016`,
				`"last_year":2024`, `"descript":"Modification de description"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.PUT("/api/plans/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "ModifyPlan") {
		t.Error(r)
	}
}

// deletePlanTest check if route is protected and plan sent back is correct.
func deletePlanTest(e *httpexpect.Expect, t *testing.T, pID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Suppression de plan, requête : Plan introuvable"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(pID),
			Status:       http.StatusOK,
			BodyContains: []string{"Plan supprimé"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.DELETE("/api/plans/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DeletePlan") {
		t.Error(r)
	}
	testCases = []testCase{
		{
			Token:         testCtx.Admin.Token,
			Status:        http.StatusOK,
			CountItemName: `"id"`,
			ArraySize:     5},
	}
	f = func(tc testCase) *httpexpect.Response {
		return e.GET("/api/plans").WithHeader("Authorization", "Bearer "+tc.Token).
			Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DeletePlan") {
		t.Error(r)
	}
}
