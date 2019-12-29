package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testPaymentPrevisions(t *testing.T) {
	t.Run("PaymentPrevisions", func(t *testing.T) {
		getPaymentPrevisionsTest(testCtx.E, t)
		getActionPaymentPrevisionsTest(testCtx.E, t)
	})
}

// getPaymentPrevisionsTest check route is protected and pre programmings correctly sent.
func getPaymentPrevisionsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase, // 0 : bad token
		{
			Token:         testCtx.User.Token,
			Status:        http.StatusOK,
			BodyContains:  []string{`"PmtPrevision":[`, `"DifPmtPrevision":[`},
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

// getActionPaymentPrevisionsTest check route is protected and pre programmings correctly sent.
func getActionPaymentPrevisionsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase, // 0 : bad token
		{
			Token:         testCtx.User.Token,
			Status:        http.StatusOK,
			BodyContains:  []string{`"DifActionPmtPrevision":[`},
			CountItemName: `"action_id"`,
			ArraySize:     83},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/payment_previsions/actions").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetActionPaymentPrevisions") {
		t.Error(r)
	}
}
