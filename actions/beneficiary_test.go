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
	testCases := []struct {
		Token        string
		Status       int
		BodyContains string
		ArraySize    int
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis", ArraySize: 0},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, BodyContains: "Beneficiary", ArraySize: 530},
	}

	for _, tc := range testCases {
		response := e.GET("/api/beneficiaries").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("Beneficiary").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

// updateBeneficiaryTest test route is protected and name changed works
func updateBeneficiaryTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		BeneficiaryID, Token string
		Status               int
		BodyContains         string
		Name                 string
	}{
		{BeneficiaryID: "1", Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis", Name: ""},
		{BeneficiaryID: "0", Token: testCtx.Admin.Token, Status: http.StatusBadRequest, BodyContains: "Champ name manquant", Name: ""},
		{BeneficiaryID: "0", Token: testCtx.Admin.Token, Status: http.StatusNotFound, BodyContains: "Bénéficiaire introuvable", Name: "Essai bénéficiaire"},
		{BeneficiaryID: "1", Token: testCtx.Admin.Token, Status: http.StatusOK, BodyContains: "Beneficiary", Name: "Essai bénéficiaire"},
	}

	for _, tc := range testCases {
		response := e.PUT("/api/beneficiaries/"+tc.BeneficiaryID).WithHeader("Authorization", "Bearer "+tc.Token).WithJSON(struct{ Name string }{tc.Name}).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.Status == http.StatusOK {
			response.JSON().Object().ContainsKey("Beneficiary")
			response.JSON().Object().Value("Beneficiary").Object().Value("name").String().Equal(tc.Name)
		}
		response.Status(tc.Status)
	}
}
