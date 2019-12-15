package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestImportLog embeddes test for import logs insuring the configuration and DB are properly initialized.
func testImportLog(t *testing.T) {
	t.Run("ImportLog", func(t *testing.T) {
		getImportLogsTest(testCtx.E, t)
	})
}

// getImportLogsTest tests route is protected and imports logs are sent back.
func getImportLogsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:  testCtx.User.Token,
			Status: http.StatusOK,
			BodyContains: []string{"ImportLog", `"id":1`, `"category":"Payments"`,
				`"id":2`, `"category":"FinancialCommitments"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/import_log").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetImportLogs") {
		t.Error(r)
	}
}
