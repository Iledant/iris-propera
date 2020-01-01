package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestConsistency embeddes all tests for document insuring the configuration and DB are properly initialized.
func testConsistency(t *testing.T) {
	t.Run("Consistency", func(t *testing.T) {
		getConsistencyDatasTest(testCtx.E, t)
	})
}

// getConsistencyDatasTest tests route is admin protected and datas are sent back.
func getConsistencyDatasTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusOK,
			BodyContains: []string{`"CommitmentWithoutAction":[`, `"UnlinkedPayment":[`},
			IDName:       `"id"`,
			ArraySize:    3},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/consistency/datas").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetConsistencyDatas") {
		t.Error(r)
	}
}
