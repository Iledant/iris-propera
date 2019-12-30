package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testSettings(t *testing.T) {
	testCommons(t)
	t.Run("Settings", func(t *testing.T) {
		t.Parallel()
		getSettingsTest(testCtx.E, t)
		getBudgetTablesTest(testCtx.E, t)
	})
}

// getSettingsTest check route is protected and datas sent has got items and number of lines.
func getSettingsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusOK,
			BodyContains: []string{"Beneficiary", "BudgetChapter", "BudgetSector",
				"BudgetProgram", "BudgetAction", "Commissions", "PhysicalOp",
				"PaymentType", "Plan", "BudgetCredits", "UnlinkedPendingCommitments",
				"LinkedPendingCommitments", "Step", "Category"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/settings").WithHeader("Authorization", "Bearer "+tc.Token).
			Expect()
	}
	for _, r := range chkTestCases(testCases, f, "Settings") {
		t.Error(r)
	}
}

// getBudgetTablesTest check route is protected and datas sent has got items and number of lines.
func getBudgetTablesTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusOK,
			BodyContains: []string{"BudgetChapter", "BudgetSector",
				"BudgetProgram", "BudgetAction"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/budget_tables").WithHeader("Authorization", "Bearer "+tc.Token).
			Expect()
	}
	for _, r := range chkTestCases(testCases, f, "BudgetTables") {
		t.Error(r)
	}
}
