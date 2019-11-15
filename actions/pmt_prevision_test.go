package actions

import (
	"net/http"
	"strings"
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
		{
			Token:        "fake",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"},
		}, // 0 : bad token
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusOK,
			BodyContains: []string{`"PmtPrevision":[`, `"DifPmtPrevision":[`},
			ArraySize:    22,
		},
	}
	for i, tc := range testCases {
		response := e.GET("/api/payment_previsions").WithHeader("Authorization", "Bearer "+tc.Token).
			Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetPaymentPrevisions[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetPaymentPrevisions[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"year"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetPaymentPrevisions[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}
