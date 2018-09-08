package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestImportLog embeddes test for import logs insuring the configuration and DB are properly initialized.
func TestImportLog(t *testing.T) {
	TestCommons(t)
	t.Run("ImportLog", func(t *testing.T) {
		getImportLogsTest(testCtx.E, t)
	})
}

// getImportLogsTest tests route is protected and imports logs are sent back.
func getImportLogsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Year         int
		Status       int
		BodyContains []string
	}{
		{Token: "", Status: http.StatusInternalServerError, BodyContains: []string{"Token absent"}},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"ImportLog", `"id":1`, `"category":"Payments"`, `"id":2`, `"category":"FinancialCommitments"`}},
	}

	for _, tc := range testCases {
		response := e.GET("/api/import_log").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}
