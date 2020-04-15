package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testPaymentDelays(t *testing.T) {
	t.Run("Payment", func(t *testing.T) {
		getPaymentDelaysTest(testCtx.E, t)
	})
}

// getPaymentDelaysTest check route is protected and payment delays correctly
// sent back.
func getPaymentDelaysTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase, // 0 no token
		{
			Token:         testCtx.User.Token,
			Status:        http.StatusBadRequest,
			Param:         "a",
			BodyContains:  []string{`Délais de paiement, décodage : `},
			CountItemName: `"delay"`,
			ArraySize:     7}, // 1 bad param
		{
			Token:         testCtx.User.Token,
			Status:        http.StatusOK,
			Param:         "1514764800000",
			BodyContains:  []string{`"payment_delay":[`, `"number":1`},
			CountItemName: `"delay"`,
			ArraySize:     13},
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusOK,
			Param:        "1546300800000",
			BodyContains: []string{`"payment_delay":[]`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/payment_delays").WithQuery("after", tc.Param).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetPaymentDelays") {
		t.Error(r)
	}
}
