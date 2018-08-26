package actions

import (
	"net/http"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func TestSettings(t *testing.T) {
	TestCommons(t)
	t.Run("Settings", func(t *testing.T) {
		getSettingsTest(testCtx.E, t)
	})
}

// getSettingsTest check route is protected and datas sent has got items and number of lines.
func getSettingsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Status       int
		BodyContains []string
	}{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{"Beneficiary", "BudgetChapter", "BudgetSector", "BudgetProgram", "BudgetAction", "Commissions", "PhysicalOp",
				"PaymentType", "Plan", "BudgetCredits", "UnlinkedPendingCommitments", "LinkedPendingCommitments", "Step", "Category"}},
	}
	for i, tc := range testCases {
		response := e.GET("/api/settings").WithHeader("Authorization", "Bearer "+tc.Token).
			Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("GetSettings[%d] : attendu %s et reçu \n%s", i, s, content)
			}
		}
		response.Status(tc.Status)
	}
}
