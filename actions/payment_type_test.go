package actions

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testPaymentType(t *testing.T) {
	t.Run("PaymentType", func(t *testing.T) {
		getPaymentTypesTest(testCtx.E, t)
		ptID := createPaymentTypeTest(testCtx.E, t)
		modifyPaymentTypeTest(testCtx.E, t, ptID)
		deletePaymentTypeTest(testCtx.E, t, ptID)
	})
}

// getPaymentTypesTest check route is protected and ratios correctly sent.
func getPaymentTypesTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"PaymentType"}, ArraySize: 3},
	}
	for i, tc := range testCases {
		response := e.GET("/api/payment_types").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetPaymentTypes[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetPaymentTypes[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetPaymentTypes[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// createPaymentTypeTest check route is protected and ratios correctly set.
func createPaymentTypeTest(e *httpexpect.Expect, t *testing.T) (ptID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest,
			Sent:         []byte(`{"name":""}`),
			BodyContains: []string{"Création d'une chronique de paiement : Name incorrect"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent:         []byte(`{"name":"Essai de chronique"}`),
			BodyContains: []string{"PaymentType", `"name":"Essai de chronique"`}},
	}

	for i, tc := range testCases {
		response := e.POST("/api/payment_types").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nCreatePaymentType[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nCreatePaymentType[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.Status == http.StatusOK {
			ptID = int(response.JSON().Object().Value("PaymentType").Object().Value("id").Number().Raw())
		}
	}
	return ptID
}

// modifyPaymentTypeTest check route is protected and ratios correctly set.
func modifyPaymentTypeTest(e *httpexpect.Expect, t *testing.T, ptID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Modification de chronique"}`),
			BodyContains: []string{"Modification d'une chronique de paiement, requête : Chronique de paiement introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(ptID), Status: http.StatusOK,
			Sent:         []byte(`{"name":"Modification de chronique"}`),
			BodyContains: []string{"PaymentType", `"name":"Modification de chronique"`}},
	}
	for i, tc := range testCases {
		response := e.PUT("/api/payment_types/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nModifyPaymentType[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nModifyPaymentType[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// deletePaymentTypeTest check route is protected and ratios correctly deleted
func deletePaymentTypeTest(e *httpexpect.Expect, t *testing.T, ptID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Suppression d'une chronique de paiement, requête : Chronique de paiement introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(ptID), Status: http.StatusOK,
			BodyContains: []string{"Chronique supprimée"}},
	}
	for i, tc := range testCases {
		response := e.DELETE("/api/payment_types/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nDeletePaymentType[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nDeletePaymentType[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}
