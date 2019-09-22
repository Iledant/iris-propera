package actions

import (
	"net/http"
	"strings"
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
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"},
		},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{"PaymentCredit":[`),
			BodyContains: []string{"Batch d'enveloppes de crédits, décodage : "},
		},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusOK,
			Sent:         []byte(`{"PaymentCredit":[{"ChapterCode":908,"SubFunctionCode":811,"PrimitiveBudget":1000000,"Reported":0,"AddedBudget":500000,"ModifyDecision":300000,"Movement":50000}]}`),
			BodyContains: []string{"Enveloppes de crédits importées"},
		},
	}
	for i, tc := range testCases {
		response := e.POST("/api/payment_credits").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nBatchPaymentCredits[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nBatchPaymentCredits[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// getPaymentCreditsTest check route is protected and datas sent back are correct
func getPaymentCreditsTest(e *httpexpect.Expect, t *testing.T) {
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
			BodyContains: []string{`Liste des enveloppes de crédits, décodage : `},
		},
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusOK,
			Param:        "2019",
			BodyContains: []string{`{"PaymentCredit":[{"Year":2019,"ChapterID":2,"ChapterCode":908,"SubFunctionCode":811,"PrimitiveBudget":1000000,"Reported":0,"AddedBudget":500000,"ModifyDecision":300000,"Movement":50000}]}`},
		},
	}
	for i, tc := range testCases {
		response := e.GET("/api/payment_credits").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("Year", tc.Param).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetPaymentCredits[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetPaymentCredits[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}

}
