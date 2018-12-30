package actions

import (
	"net/http"
	"strings"
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
		{Token: "", Status: http.StatusInternalServerError, BodyContains: []string{"Token absent"}},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"ImportLog", `"id":1`, `"category":"Payments"`,
				`"id":2`, `"category":"FinancialCommitments"`}},
	}

	for i, tc := range testCases {
		response := e.GET("/api/import_log").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetImportLogs[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetImportLogs[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}
