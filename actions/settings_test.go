package actions

import (
	"net/http"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testSettings(t *testing.T) {
	testCommons(t)
	t.Run("Settings", func(t *testing.T) {
		getSettingsTest(testCtx.E, t)
		getBudgetTablesTest(testCtx.E, t)
	})
}

// getSettingsTest check route is protected and datas sent has got items and number of lines.
func getSettingsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{"Beneficiary", "BudgetChapter", "BudgetSector",
				"BudgetProgram", "BudgetAction", "Commissions", "PhysicalOp",
				"PaymentType", "Plan", "BudgetCredits", "UnlinkedPendingCommitments",
				"LinkedPendingCommitments", "Step", "Category"}},
	}
	for i, tc := range testCases {
		response := e.GET("/api/settings").WithHeader("Authorization", "Bearer "+tc.Token).
			Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nSettings[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nSettings[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}

	}
}

// getBudgetTablesTest check route is protected and datas sent has got items and number of lines.
func getBudgetTablesTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{"BudgetChapter", "BudgetSector",
				"BudgetProgram", "BudgetAction"}},
	}
	for i, tc := range testCases {
		response := e.GET("/api/budget_tables").WithHeader("Authorization", "Bearer "+tc.Token).
			Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nBudgetTables[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nBudgetTables[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}

	}
}
