package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func TestPaymentType(t *testing.T) {
	TestCommons(t)
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
	for _, tc := range testCases {
		response := e.GET("/api/payment_types").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("PaymentType").Array().Length().Equal(tc.ArraySize)
		}
	}
}

// createPaymentTypeTest check route is protected and ratios correctly set.
func createPaymentTypeTest(e *httpexpect.Expect, t *testing.T) (ptID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent:         []byte(`{"name":"Essai de chronique"}`),
			BodyContains: []string{"PaymentType", `"name":"Essai de chronique"`}},
	}

	for _, tc := range testCases {
		response := e.POST("/api/payment_types").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
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
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusBadRequest,
			BodyContains: []string{"Modification d'une chronique : introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(ptID), Status: http.StatusOK,
			Sent:         []byte(`{"name":"Modification de chronique"}`),
			BodyContains: []string{"PaymentType", `"name":"Modification de chronique"`}},
	}
	for _, tc := range testCases {
		response := e.PUT("/api/payment_types/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}

// deletePaymentTypeTest check route is protected and ratios correctly deleted
func deletePaymentTypeTest(e *httpexpect.Expect, t *testing.T, ptID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusBadRequest,
			BodyContains: []string{"Suppression d'une chronique : introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(ptID), Status: http.StatusOK,
			BodyContains: []string{"Chronique supprimée"}},
	}
	for _, tc := range testCases {
		response := e.DELETE("/api/payment_types/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}
