package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testPaymentDemandsStocks(t *testing.T) {
	t.Run("PaymentDemandsStocks", func(t *testing.T) {
		getPaymentDemandsStocksTest(testCtx.E, t)
	})
}

// getPaymentDemandsStocksTest check route is protected and datas are correctly
// sent back.
func getPaymentDemandsStocksTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusOK,
			BodyContains: []string{`{"PaymentDemandsStock":[`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/payment_demand_stocks").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetPaymentDemandsStocks") {
		t.Error(r)
	}
}
