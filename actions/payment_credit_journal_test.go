package actions

import (
	"net/http"
	"strings"
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
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"},
		},
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
	for i, tc := range testCases {
		response := e.POST("/api/payment_credits/journal").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nBatchPaymentCreditJournals[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nBatchPaymentCreditJournals[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// getPaymentCreditJournalsTest check route is protected and datas sent back are correct
func getPaymentCreditJournalsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{
			Token:        "fake",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"},
		},
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusBadRequest,
			Param:        "a",
			BodyContains: []string{`Mouvements de crédits, décodage : `},
		},
		{
			Token:  testCtx.User.Token,
			Status: http.StatusOK,
			Param:  "2019",
			BodyContains: []string{`{"PaymentCreditJournal":[{"Chapter":908,"ID":1,` +
				`"Function":811,"CreationDate":"2019-03-10T00:00:00Z","ModificationFDate"` +
				`:"2019-03-15T00:00:00Z","Name":"Mouvement","Value":100000}]}`},
		},
	}
	for i, tc := range testCases {
		response := e.GET("/api/payment_credits/journal").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("Year", tc.Param).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetPaymentCreditJournals[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetPaymentCreditJournals[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}

}
