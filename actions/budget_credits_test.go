package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestBudgetCredit embeddes all tests for budget credits insuring the configuration and DB are properly initialized.
func TestBudgetCredit(t *testing.T) {
	TestCommons(t)
	t.Run("BudgetCredit", func(t *testing.T) {
		getBudgetCredits(testCtx.E, t)
		getLastBudgetCredits(testCtx.E, t)
		brID := createBudgetCreditTest(testCtx.E, t)
		modifyBudgetCreditTest(testCtx.E, t, brID)
		deleteBudgetCreditTest(testCtx.E, t, brID)
		batchBudgetCreditTest(testCtx.E, t)
	})
}

// getBudgetCredits tests route is protected and all credits are sent back.
func getBudgetCredits(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token absent"}, ArraySize: 0},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"BudgetCredits"}, ArraySize: 75},
	}

	for _, tc := range testCases {
		response := e.GET("/api/budget_credits").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("BudgetCredits").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

// getLastBudgetCredits tests route is protected and all credits are sent back.
func getLastBudgetCredits(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token absent"}, ArraySize: 0},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"BudgetCredits", "2018-07-04"}, ArraySize: 3},
	}

	for _, tc := range testCases {
		response := e.GET("/api/budget_credits/year").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("BudgetCredits").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

// createBudgetCreditTest tests route is protected and sent credit is created.
func createBudgetCreditTest(e *httpexpect.Expect, t *testing.T) (brID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest, Sent: []byte(`{}`),
			BodyContains: []string{"Création de crédits : champ manquant ou incorrect"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest,
			Sent:         []byte(`{"commission_date": "2018-04-01T00:00:00.000Z", "chapter": 907}`),
			BodyContains: []string{"Création de crédits : champ manquant ou incorrect"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent: []byte(`{"commission_date": "2018-04-01T00:00:00.000Z", "chapter": 907,
			"primary_commitment":123456,"reserved_commitment":123,"frozen_commitment":456}`),
			BodyContains: []string{"BudgetCredits", `"primary_commitment":123456`,
				`"reserved_commitment":123`, `"frozen_commitment":456`, `"chapter_id":3`}},
	}

	for _, tc := range testCases {
		response := e.POST("/api/budget_credits").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		if tc.Status == http.StatusOK {
			brID = int(response.JSON().Object().Value("BudgetCredits").Object().Value("id").Number().Raw())
		}
		response.Status(tc.Status)
	}
	return brID
}

// modifyBudgetCreditTest tests route is protected and modify work properly.
func modifyBudgetCreditTest(e *httpexpect.Expect, t *testing.T, brID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusBadRequest,
			Sent: []byte(`{"chapter":908}`), BodyContains: []string{"Modification des crédits: introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(brID), Status: http.StatusOK,
			Sent: []byte(`{"chapter":908}`), BodyContains: []string{"BudgetCredits", `"chapter_id":2`}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(brID), Status: http.StatusOK,
			Sent:         []byte(`{"commission_date":"2018-08-13T09:21:56.132Z"}`),
			BodyContains: []string{"BudgetCredits", `"commission_date":"2018-08-13T09:21:56.132Z"`}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(brID), Status: http.StatusOK,
			Sent:         []byte(`{"primary_commitment":999}`),
			BodyContains: []string{"BudgetCredits", `"primary_commitment":999`}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(brID), Status: http.StatusOK,
			Sent:         []byte(`{"frozen_commitment":888}`),
			BodyContains: []string{"BudgetCredits", `"frozen_commitment":888`}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(brID), Status: http.StatusOK,
			Sent:         []byte(`{"reserved_commitment":777}`),
			BodyContains: []string{"BudgetCredits", `"reserved_commitment":777`}},
	}

	for _, tc := range testCases {
		response := e.PUT("/api/budget_credits/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}

// deleteBudgetCreditTest tests route is protected and delete work properly.
func deleteBudgetCreditTest(e *httpexpect.Expect, t *testing.T, brID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusBadRequest,
			BodyContains: []string{"Suppression de crédits: introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(brID), Status: http.StatusOK,
			BodyContains: []string{"Crédits supprimés"}},
	}

	for _, tc := range testCases {
		response := e.DELETE("/api/budget_credits/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}

//batchBudgetCreditTest tests route is protected and sent datas are correctly incorporated.
func batchBudgetCreditTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusInternalServerError,
			Sent: []byte(`fake`), BodyContains: []string{"Erreur de lecture du batch crédits"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest,
			Sent:         []byte(`{"BudgetCredits":[{"commission_date":"2018-04-01T00:00:00Z"}]}`),
			BodyContains: []string{"Batch crédits, champs manquants"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent: []byte(`{"BudgetCredits":[
				{"commission_date":"2018-07-04T20:00:00Z","chapter":907,
				 "primary_commitment":999,"reserved_commitment":888,"frozen_commitment":777},
				{"commission_date":"2018-04-01T00:00:00.000Z","chapter":907,
				 "primary_commitment":999,"reserved_commitment":888,"frozen_commitment":777}
				]}`), BodyContains: []string{"Credits importés"}},
	}

	for _, tc := range testCases {
		response := e.POST("/api/budget_credits/array").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
	// Check only one line has been created
	response := e.GET("/api/budget_credits").
		WithHeader("Authorization", "Bearer "+testCtx.Admin.Token).Expect()
	response.JSON().Object().Value("BudgetCredits").Array().Length().Equal(76)
}
