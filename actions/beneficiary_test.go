package actions

import (
	"net/http"
	"strings"
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

	for i, tc := range testCases {
		response := e.GET("/api/beneficiaries").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetBeneficiaries[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetBeneficiaries[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetBeneficiaries[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// updateBeneficiaryTest test route is protected and name changed works
func updateBeneficiaryTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{ID: "1", Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{ID: "0", Token: testCtx.Admin.Token, Status: http.StatusBadRequest,
			BodyContains: []string{"Modification de bénéficiaire : Champ name manquant"},
			Sent:         []byte("{}")},
		{ID: "0", Token: testCtx.Admin.Token, Status: http.StatusInternalServerError,
			BodyContains: []string{"Modification de bénéficiaire, requête : Bénéficiaire introuvable"},
			Sent:         []byte(`{"Name":"Essai bénéficiaire"}`)},
		{ID: "1", Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{"Beneficiary", `"name":"Essai bénéficiaire"`},
			Sent:         []byte(`{"Name":"Essai bénéficiaire"}`)},
	}

	for i, tc := range testCases {
		response := e.PUT("/api/beneficiaries/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nUpdateBeneficiary[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nUpdateBeneficiary[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}
