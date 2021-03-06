package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestBudgetCredit embeddes all tests for budget credits insuring the configuration and DB are properly initialized.
func testBudgetCredit(t *testing.T) {
	t.Run("BudgetCredit", func(t *testing.T) {
		getBudgetCredits(testCtx.E, t)
		getLastBudgetCredits(testCtx.E, t)
		brID := createBudgetCreditTest(testCtx.E, t)
		if brID == 0 {
			t.Fatal("Impossible de créer le BudgetCredit")
		}
		modifyBudgetCreditTest(testCtx.E, t, brID)
		deleteBudgetCreditTest(testCtx.E, t, brID)
		batchBudgetCreditTest(testCtx.E, t)
	})
}

// getBudgetCredits tests route is protected and all credits are sent back.
func getBudgetCredits(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase, // 0 : missing token
		{
			Token:         testCtx.User.Token,
			Status:        http.StatusOK,
			BodyContains:  []string{"BudgetCredits", "BudgetChapter"},
			ArraySize:     78,
			CountItemName: `"id"`,
		}, // 1 : ok
	}

	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/budget_credits").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetBudgetCredits") {
		t.Error(r)
	}
}

// getLastBudgetCredits tests route is protected and all credits are sent back.
func getLastBudgetCredits(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusOK,
			BodyContains: []string{"BudgetCredits"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/budget_credits/year").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetLastBudgetCredits") {
		t.Error(r)
	}
}

// createBudgetCreditTest tests route is protected and sent credit is created.
func createBudgetCreditTest(e *httpexpect.Expect, t *testing.T) (brID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{}`),
			BodyContains: []string{"Création de crédits : Erreur de chapitre ou de date de commission"}},
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusCreated,
			Sent: []byte(`{"commission_date": "2018-04-01T00:00:00.000Z", "chapter": 907,
			"primary_commitment":123456,"reserved_commitment":123,"frozen_commitment":456}`),
			BodyContains: []string{"BudgetCredits", `"primary_commitment":123456`,
				`"reserved_commitment":123`, `"frozen_commitment":456`, `"chapter":907`},
			IDName: `"id"`},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/budget_credits").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "CreateBudgetCredits", &brID) {
		t.Error(r)
	}
	return brID
}

// modifyBudgetCreditTest tests route is protected and modify work properly.
func modifyBudgetCreditTest(e *httpexpect.Expect, t *testing.T, brID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{"chapter":908}`),
			BodyContains: []string{"Modification de crédits Erreur de chapitre ou de date de commission"}},
		{
			Token: testCtx.Admin.Token, ID: strconv.Itoa(brID),
			Status:       http.StatusOK,
			Sent:         []byte(`{"commission_date":"2018-08-13T09:21:56.132Z","chapter":908,"primary_commitment":999,"frozen_commitment":888,"reserved_commitment":777}`),
			BodyContains: []string{"BudgetCredits", `"commission_date":"2018-08-13`, `"primary_commitment":999`, `"frozen_commitment":888`, `"reserved_commitment":777`, `"chapter":908`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.PUT("/api/budget_credits/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "ModifyBudgetCredit") {
		t.Error(r)
	}
}

// deleteBudgetCreditTest tests route is protected and delete work properly.
func deleteBudgetCreditTest(e *httpexpect.Expect, t *testing.T, brID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Suppression de crédits, requête : Crédits introuvables"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(brID),
			Status:       http.StatusOK,
			BodyContains: []string{"Crédits supprimés"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.DELETE("/api/budget_credits/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DeleteBudgetCredit") {
		t.Error(r)
	}
}

//batchBudgetCreditTest tests route is protected and sent datas are correctly incorporated.
func batchBudgetCreditTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`fake`),
			BodyContains: []string{"Erreur de lecture du batch crédits"}},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"BudgetCredits":[{"commission_date":43191}]}`),
			BodyContains: []string{"Batch crédits, requête : Date de commission ou chapitre incorrect"}},
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusOK,
			Sent: []byte(`{"BudgetCredits":[
				{"commission_date":43285,"chapter":907,
				 "primary_commitment":999,"reserved_commitment":888.50,"frozen_commitment":777},
				{"commission_date":43191,"chapter":907,
				 "primary_commitment":999,"reserved_commitment":888,"frozen_commitment":777}
				]}`),
			BodyContains: []string{"Credits importés"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/budget_credits/array").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "BatchBudgetCredits") {
		t.Error(r)
	}

	// Check only one line has been created
	f = func(tc testCase) *httpexpect.Response {
		return e.GET("/api/budget_credits").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	testCases = []testCase{
		{
			Token:         testCtx.Admin.Token,
			Status:        http.StatusOK,
			ArraySize:     79,
			CountItemName: `"id"`},
	}
	for _, r := range chkTestCases(testCases, f, "BatchBudgetCredits") {
		t.Error(r)
	}
}
