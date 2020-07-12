package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestFlowStockDelays implements tests for beneficiary handlers.
func testFlowStockDelays(t *testing.T) {
	t.Run("FlowStockDelays", func(t *testing.T) {
		getTestFlowStockDelays(testCtx.E, t)
	})
}

// getTestFlowStockDelays test route is protected and the response fits.
func getTestFlowStockDelays(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusOK,
			Param:        "90",
			BodyContains: []string{`"FlowStockDelays"`},
		},
	}

	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/flow_stock_delays").WithQuery("Days", tc.Param).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetTestFlowStockDelays") {
		t.Error(r)
	}
}
