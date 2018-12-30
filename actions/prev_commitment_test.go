package actions

import (
	"net/http"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testPrevCommitment(t *testing.T) {
	t.Run("Prevcommitment", func(t *testing.T) {
		batchPrevCommitmentsTest(testCtx.E, t)
	})
}

// batchPrevCommitmentsTest check route is protected and return successful.
func batchPrevCommitmentsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusInternalServerError,
			Sent:         []byte(`{Pend}`),
			BodyContains: []string{"Batch prévision d'engagements : décodage"}},
		//cSpell:disable
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Sent: []byte(`{"PrevCommitment": [
			{"number":"01BU002","year":2019,"value":100000000,"total_value":400000000,"state_ratio":0.31},
			{"number":"11AC001","year":2019,"value":500000000,"total_value":null,"state_ratio":null}]}`),
			BodyContains: []string{"Batch prévision d'engagement importé"}},
		//cSpell:enable
	}
	for i, tc := range testCases {
		response := e.POST("/api/prev_commitments").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nBatchPrevCommitments[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nBatchPrevCommitments[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}
