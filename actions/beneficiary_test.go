package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestBeneficiary implements tests for beneficiary handlers.
func TestBeneficiary(t *testing.T) {
	TestCommons(t)
	t.Run("Beneficiary", func(t *testing.T) {
		getBeneficiariesTest(testCtx.E, t)
		updateBeneficiaryTest(testCtx.E, t)
	})
}

// getBeneficiariesTest test route is protected and the response fits.
func getBeneficiariesTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}, ArraySize: 0},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{"Beneficiary"}, ArraySize: 530},
	}

	for _, tc := range testCases {
		response := e.GET("/api/beneficiaries").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("Beneficiary").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

// updateBeneficiaryTest test route is protected and name changed works
func updateBeneficiaryTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{ID: "1", Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{ID: "0", Token: testCtx.Admin.Token, Status: http.StatusBadRequest,
			BodyContains: []string{"Modification de bénéficiaire : champ name manquant"},
			Sent:         []byte("{}")},
		{ID: "0", Token: testCtx.Admin.Token, Status: http.StatusBadRequest,
			BodyContains: []string{"Modification de bénéficiaire : introuvable"},
			Sent:         []byte(`{"Name":"Essai bénéficiaire"}`)},
		{ID: "1", Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{"Beneficiary", `"name":"Essai bénéficiaire"`},
			Sent:         []byte(`{"Name":"Essai bénéficiaire"}`)},
	}

	for _, tc := range testCases {
		response := e.PUT("/api/beneficiaries/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}
