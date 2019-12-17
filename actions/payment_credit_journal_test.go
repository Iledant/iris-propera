package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testPaymentCreditJournals(t *testing.T) {
	t.Run("PaymentCreditJournals", func(t *testing.T) {
		batchPaymentCreditJournalsTest(testCtx.E, t)
		getPaymentCreditJournalsTest(testCtx.E, t)
	})
}

// batchPaymentCreditJournalsTest check route is admin protected and response is ok
func batchPaymentCreditJournalsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{"PaymentCredit":[`),
			BodyContains: []string{"Batch mouvements de crédits, décodage : "},
		},
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusOK,
			Sent: []byte(`{"PaymentCreditJournal":[{"Chapter":908,"Function":811,` +
				`"CreationDate":20190310,"ModificationDate":20190315,"Name":"Mouvement","Value":100000}]}`),
			BodyContains: []string{"Mouvements de crédits importés"},
		},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/payment_credits/journal").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "BatchPaymentCreditJournals") {
		t.Error(r)
	}
}

// getPaymentCreditJournalsTest check route is protected and datas sent back are correct
func getPaymentCreditJournalsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusBadRequest,
			Param:        "a",
			BodyContains: []string{`Mouvements de crédits, décodage : `}},
		{
			Token:  testCtx.User.Token,
			Status: http.StatusOK,
			Param:  "2019",
			BodyContains: []string{`{"PaymentCreditJournal":[{"Chapter":908,"ID":1,` +
				`"Function":811,"CreationDate":"2019-03-10T00:00:00Z","ModificationDate"` +
				`:"2019-03-15T00:00:00Z","Name":"Mouvement","Value":100000}]}`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/payment_credits/journal").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("Year", tc.Param).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetPaymentCreditjournal") {
		t.Error(r)
	}
}
