package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testPaymentPrevisions(t *testing.T) {
	t.Run("PaymentPrevisions", func(t *testing.T) {
		getPaymentPrevisionsTest(testCtx.E, t)
	})
}

// getPaymentPrevisionsTest check route is protected and pre programmings correctly sent.
func getPaymentPrevisionsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase, // 0 : bad token
		{
			Token:         testCtx.User.Token,
			Status:        http.StatusOK,
			BodyContains:  []string{`"PmtPreviion":[`, `"DifPmtPrevision":[`},
			CountItemName: `"year"`,
			ArraySize:     9},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/payment_previsions").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetPaymentPrevisions") {
		t.Error(r)
	}
}
