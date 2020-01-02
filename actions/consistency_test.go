package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestConsistency embeddes all tests for document insuring the configuration and DB are properly initialized.
func testConsistency(t *testing.T) {
	t.Run("Consistency", func(t *testing.T) {
		getConsistencyDatasTest(testCtx.E, t)
		getPossibleLinkedCmtsTest(testCtx.E, t)
		linkPaymentToCmtTest(testCtx.E, t)
	})
}

// getConsistencyDatasTest tests route is admin protected and datas are sent back.
func getConsistencyDatasTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusOK,
			BodyContains: []string{`"CommitmentWithoutAction":[`, `"UnlinkedPayment":[`},
			IDName:       `"id"`,
			ArraySize:    3},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/consistency/datas").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetConsistencyDatas") {
		t.Error(r)
	}
}

// getPossibleLinkedCmtsTest tests route is admin protected and datas are sent back.
func getPossibleLinkedCmtsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:         testCtx.Admin.Token,
			Status:        http.StatusOK,
			ID:            "11326",
			BodyContains:  []string{`"Commitment":[`, `"id":2267`},
			CountItemName: `"id"`,
			ArraySize:     10},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/payment/"+tc.ID+"/possible_linked_commitment").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetPossibleLinkedCmts") {
		t.Error(r)
	}
}

// linkPaymentToCmtTest tests route is admin protected and ok status sent back.
func linkPaymentToCmtTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:         testCtx.Admin.Token,
			Status:        http.StatusInternalServerError,
			ID:            "11326",
			Param:         "0",
			BodyContains:  []string{`Lien paiement engagement, requête : `},
			CountItemName: `"id"`,
			ArraySize:     10},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusOK,
			ID:           "11326",
			Param:        "2267",
			BodyContains: []string{`Paiement rattaché à l'engagement`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/payment/"+tc.ID+"/link_commitment/"+tc.Param).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "LinkPaymentToCmt") {
		t.Error(r)
	}
}
