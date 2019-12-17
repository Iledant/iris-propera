package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testPaymentCredits(t *testing.T) {
	t.Run("PaymentCredits", func(t *testing.T) {
		batchPaymentCreditsTest(testCtx.E, t)
		getPaymentCreditsTest(testCtx.E, t)
	})
}

// batchPaymentCreditsTest check route is admin protected and response is ok
func batchPaymentCreditsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{"PaymentCredit":[`),
			BodyContains: []string{"Batch d'enveloppes de crédits, décodage : "}},
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusOK,
			Sent: []byte(`{"PaymentCredit":[{"Chapter":908,"Function":811,` +
				`"Primitive":1000000,"Reported":0,"Added":500000,"Modified":300000,` +
				`"Movement":50000}]}`),
			BodyContains: []string{"Enveloppes de crédits importées"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/payment_credits").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "BatchPaymentCredits") {
		t.Error(r)
	}
}

// getPaymentCreditsTest check route is protected and datas sent back are correct
func getPaymentCreditsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusBadRequest,
			Param:        "a",
			BodyContains: []string{`Liste des enveloppes de crédits, décodage : `}},
		{
			Token:  testCtx.User.Token,
			Status: http.StatusOK,
			Param:  "2019",
			BodyContains: []string{`{"PaymentCredit":[{"Year":2019,"ChapterID":2,` +
				`"Chapter":908,"Function":811,"Primitive":1000000,"Reported":0,` +
				`"Added":500000,"Modified":300000,"Movement":50000}]}`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/payment_credits").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("Year", tc.Param).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetPayementCredits") {
		t.Error(r)
	}
}
