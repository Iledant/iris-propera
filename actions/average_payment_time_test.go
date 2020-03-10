package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestAvgPmtTime embeddes all tests for category insuring the configuration and DB are properly initialized.
func testAvgPmtTime(t *testing.T) {
	t.Run("AvgPmtTimes", func(t *testing.T) {
		getAvgPmtTimesTest(testCtx.E, t)
	})
}

// getAvgPmtTimesTest tests route is protected and all commissions are sent back.
func getAvgPmtTimesTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusOK,
			BodyContains: []string{"AveragePaymentTime"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/average_payment_time").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetAvgPmtTime") {
		t.Error(r)
	}
}
