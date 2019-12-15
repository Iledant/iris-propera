package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestBeneficiary implements tests for beneficiary handlers.
func testBeneficiary(t *testing.T) {
	t.Run("Beneficiary", func(t *testing.T) {
		getBeneficiariesTest(testCtx.E, t)
		updateBeneficiaryTest(testCtx.E, t)
		getBeneficiaryCmtsTest(testCtx.E, t)
	})
}

// getBeneficiariesTest test route is protected and the response fits.
func getBeneficiariesTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:         testCtx.User.Token,
			Status:        http.StatusOK,
			BodyContains:  []string{"Beneficiary"},
			ArraySize:     530,
			CountItemName: `"id"`},
	}

	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/beneficiaries").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetBeneficiaries") {
		t.Error(r)
	}
}

// updateBeneficiaryTest test route is protected and name changed works
func updateBeneficiaryTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			ID:           "0",
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			BodyContains: []string{"Modification de bénéficiaire : Champ name manquant"},
			Sent:         []byte("{}")},
		{
			ID:           "0",
			Token:        testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Modification de bénéficiaire, requête : Bénéficiaire introuvable"},
			Sent:         []byte(`{"Name":"Essai bénéficiaire"}`)},
		{
			ID:           "1",
			Token:        testCtx.Admin.Token,
			Status:       http.StatusOK,
			BodyContains: []string{"Beneficiary", `"name":"Essai bénéficiaire"`},
			Sent:         []byte(`{"Name":"Essai bénéficiaire"}`)},
	}

	f := func(tc testCase) *httpexpect.Response {
		return e.PUT("/api/beneficiaries/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "UpdateBeneficiary") {
		t.Error(r)
	}
}

// getBeneficiaryCmtsTest test route is users protected and the response fits.
func getBeneficiaryCmtsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			ID:           "a",
			Status:       http.StatusBadRequest,
			BodyContains: []string{"Engagement d'un bénéficiaire, paramètre :"}},
		{
			Token:         testCtx.User.Token,
			ID:            "10",
			Status:        http.StatusOK,
			BodyContains:  []string{`"BeneficiaryCommitment":[`},
			ArraySize:     171,
			CountItemName: `"id"`},
	}

	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/beneficiary/"+tc.ID+"/commitment").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetBeneficiaryCmts") {
		t.Error(r)
	}
}
