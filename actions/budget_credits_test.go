package actions

import (
	"net/http"
	"strconv"
	"strings"
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

	for i, tc := range testCases {
		response := e.GET("/api/budget_credits").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetBudgetCredits[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetBudgetCredits[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetBudgetCredits[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
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

	for i, tc := range testCases {
		response := e.GET("/api/budget_credits/year").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetLastBudgetCredits[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetLastBudgetCredits[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetLastBudgetCredits[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// createBudgetCreditTest tests route is protected and sent credit is created.
func createBudgetCreditTest(e *httpexpect.Expect, t *testing.T) (brID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest, Sent: []byte(`{}`),
			BodyContains: []string{"Création de crédits : Erreur de chapitre ou de date de commission"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent: []byte(`{"commission_date": "2018-04-01T00:00:00.000Z", "chapter": 907,
			"primary_commitment":123456,"reserved_commitment":123,"frozen_commitment":456}`),
			BodyContains: []string{"BudgetCredits", `"primary_commitment":123456`,
				`"reserved_commitment":123`, `"frozen_commitment":456`, `"chapter_id":3`}},
	}

	for i, tc := range testCases {
		response := e.POST("/api/budget_credits").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nCreateBudgetCredits[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nCreateBudgetCredits[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.Status == http.StatusOK {
			brID = int(response.JSON().Object().Value("BudgetCredits").Object().Value("id").Number().Raw())
		}
	}
	return brID
}

// modifyBudgetCreditTest tests route is protected and modify work properly.
func modifyBudgetCreditTest(e *httpexpect.Expect, t *testing.T, brID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusBadRequest,
			Sent:         []byte(`{"chapter":908}`),
			BodyContains: []string{"Modification de crédits Erreur de chapitre ou de date de commission"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(brID), Status: http.StatusOK,
			Sent:         []byte(`{"commission_date":"2018-08-13T09:21:56.132Z","chapter":908,"primary_commitment":999,"frozen_commitment":888,"reserved_commitment":777}`),
			BodyContains: []string{"BudgetCredits", `"commission_date":"2018-08-13T00:00:00Z"`, `"primary_commitment":999`, `"frozen_commitment":888`, `"reserved_commitment":777`, `"chapter_id":2`}},
	}

	for i, tc := range testCases {
		response := e.PUT("/api/budget_credits/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nModifyBudgetCredit[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nModifyBudgetCredit[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// deleteBudgetCreditTest tests route is protected and delete work properly.
func deleteBudgetCreditTest(e *httpexpect.Expect, t *testing.T, brID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Suppression de crédits, requête : Crédits introuvables"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(brID), Status: http.StatusOK,
			BodyContains: []string{"Crédits supprimés"}},
	}
	for i, tc := range testCases {
		response := e.DELETE("/api/budget_credits/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()

		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nDeleteBudgetCredit[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nDeleteBudgetCredit[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

//batchBudgetCreditTest tests route is protected and sent datas are correctly incorporated.
func batchBudgetCreditTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusInternalServerError,
			Sent: []byte(`fake`), BodyContains: []string{"Erreur de lecture du batch crédits"}},
		{Token: testCtx.Admin.Token, Status: http.StatusInternalServerError,
			Sent:         []byte(`{"BudgetCredits":[{"commission_date":43191}]}`),
			BodyContains: []string{"Batch crédits, requête : Date de commission ou chapitre incorrect"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent: []byte(`{"BudgetCredits":[
				{"commission_date":43285,"chapter":907,
				 "primary_commitment":999,"reserved_commitment":888.50,"frozen_commitment":777},
				{"commission_date":43191,"chapter":907,
				 "primary_commitment":999,"reserved_commitment":888,"frozen_commitment":777}
				]}`), BodyContains: []string{"Credits importés"}},
	}

	for i, tc := range testCases {
		response := e.POST("/api/budget_credits/array").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nBatchBudgetCredit[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nBatchBudgetCredit[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
	// Check only one line has been created
	response := e.GET("/api/budget_credits").
		WithHeader("Authorization", "Bearer "+testCtx.Admin.Token).Expect()

	content := string(response.Content)
	count := strings.Count(content, `"id"`)
	if count != 76 {
		t.Errorf("\nBatchBudgetCredits :\n  nombre attendu -> %d\n  nombre reçu <-%d", 76, count)
	}
}
