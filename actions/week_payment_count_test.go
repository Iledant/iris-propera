package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testWeekPaymentCounts(t *testing.T) {
	t.Run("Payment", func(t *testing.T) {
		getWeekPaymentCountsTest(testCtx.E, t)
	})
}

// getWeekPaymentCountsTest check route is protected and payment delays correctly
// sent back.
func getWeekPaymentCountsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase, // 0 no token
		{
			Token:         testCtx.User.Token,
			Status:        http.StatusBadRequest,
			Param:         "a",
			BodyContains:  []string{`Paiements par semaine, décodage : `},
			CountItemName: `"delay"`,
			ArraySize:     7}, // 1 bad param
		{
			Token:  testCtx.User.Token,
			Status: http.StatusOK,
			Param:  "2017",
			BodyContains: []string{`WeekPaymentCount":[`, `"week_number":30`,
				`"received_number":0`, `"payment_number":14`},
			CountItemName: `"week_number"`,
			ArraySize:     52},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/week_payment_counts").WithQuery("year", tc.Param).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetWeekPaymentCounts") {
		t.Error(r)
	}
}
