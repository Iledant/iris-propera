package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testPaymentRatio(t *testing.T) {
	t.Run("PaymentRatios", func(t *testing.T) {
		getRatiosTest(testCtx.E, t)
		getPtRatiosTest(testCtx.E, t)
		setPtRatiosTest(testCtx.E, t)
		deletePtRatiosTest(testCtx.E, t)
		getYearRatiosTest(testCtx.E, t)
	})
}

// getRatiosTest check route is protected and ratios correctly sent.
func getRatiosTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:         testCtx.User.Token,
			Status:        http.StatusOK,
			BodyContains:  []string{"PaymentRatio"},
			CountItemName: `"id"`,
			ArraySize:     26},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/payment_ratios").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetRatios") {
		t.Error(r)
	}
}

// getPtRatiosTest check route is protected and ratios correctly sent.
func getPtRatiosTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusOK,
			ID:           "0",
			BodyContains: []string{`"PaymentRatio":[]`}},
		{
			Token:         testCtx.User.Token,
			Status:        http.StatusOK,
			ID:            "5",
			BodyContains:  []string{"PaymentRatio"},
			CountItemName: `"id"`,
			ArraySize:     8},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/payment_types/"+tc.ID+"/payment_ratios").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetPtRatios") {
		t.Error(r)
	}
}

// setPtRatiosTest check route is protected and ratios correctly set.
func setPtRatiosTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:  testCtx.Admin.Token,
			ID:     "0",
			Status: http.StatusInternalServerError,
			Sent: []byte(`{"PaymentRatio":[{"ratio":0.05,"index":0},
		{"ratio":0.1,"index":1},{"ratio":0.15,"index":2},{"ratio":0.25,"index":3},
		{"ratio":0.45,"index":4}]}`),
			BodyContains: []string{"Ratios d'une chronique, requête : pq"}},
		{
			Token:         testCtx.Admin.Token,
			ID:            "5",
			Status:        http.StatusOK,
			CountItemName: `"id"`,
			Sent: []byte(`{"PaymentRatio":[{"ratio":0.05,"index":0},
			{"ratio":0.1,"index":1},{"ratio":0.15,"index":2},{"ratio":0.25,"index":3},
			{"ratio":0.45,"index":4}]}`),
			BodyContains: []string{"PaymentRatio", `"ratio":0.05,"index":0`},
			ArraySize:    5},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/payment_types/"+tc.ID+"/payment_ratios").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "SetPtRatios") {
		t.Error(r)
	}
}

// deletePtRatiosTest check route is protected and ratios correctly deleted
func deletePtRatiosTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Suppression des ratios d'une chronique, requête : Ratios de paiement introuvables"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           "5",
			Status:       http.StatusOK,
			BodyContains: []string{"Ratios supprimés"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.DELETE("/api/payment_types/"+tc.ID+"/payment_ratios").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DeletePtRatios") {
		t.Error(r)
	}
	testCases = []testCase{
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusOK,
			ID:           "5",
			BodyContains: []string{`"PaymentRatio":[]`}},
	}
	f = func(tc testCase) *httpexpect.Response {
		return e.GET("/api/payment_types/"+tc.ID+"/payment_ratios").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DeletePtRatios") {
		t.Error(r)
	}
}

// getYearRatiosTest check route is protected and ratios correctly calculated
func getYearRatiosTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusBadRequest,
			BodyContains: []string{"Ratios annuels : année manquante"}},
		{
			Token:         testCtx.User.Token,
			Status:        http.StatusOK,
			Param:         "2011",
			BodyContains:  []string{"Ratios", `"ratio":0.108592`},
			CountItemName: `"index"`,
			ArraySize:     8},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/payment_ratios/year").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("Year", tc.Param).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetYearRatios") {
		t.Error(r)
	}
}
