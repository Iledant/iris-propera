package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testPaymentType(t *testing.T) {
	t.Run("PaymentType", func(t *testing.T) {
		getPaymentTypesTest(testCtx.E, t)
		ptID := createPaymentTypeTest(testCtx.E, t)
		if ptID == 0 {
			t.Fatal("Impossible de créer le type de paiement")
		}
		modifyPaymentTypeTest(testCtx.E, t, ptID)
		deletePaymentTypeTest(testCtx.E, t, ptID)
	})
}

// getPaymentTypesTest check route is protected and ratios correctly sent.
func getPaymentTypesTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:         testCtx.User.Token,
			Status:        http.StatusOK,
			BodyContains:  []string{"PaymentType"},
			CountItemName: `"id"`,
			ArraySize:     3},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/payment_types").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetPaymentTypes") {
		t.Error(r)
	}
}

// createPaymentTypeTest check route is protected and ratios correctly set.
func createPaymentTypeTest(e *httpexpect.Expect, t *testing.T) (ptID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{"name":""}`),
			BodyContains: []string{"Création d'une chronique de paiement : Name incorrect"}},

		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusCreated,
			Sent:         []byte(`{"name":"Essai de chronique"}`),
			IDName:       `"id"`,
			BodyContains: []string{"PaymentType", `"name":"Essai de chronique"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/payment_types").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "CreatePaymentType", &ptID) {
		t.Error(r)
	}
	return ptID
}

// modifyPaymentTypeTest check route is protected and ratios correctly set.
func modifyPaymentTypeTest(e *httpexpect.Expect, t *testing.T, ptID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Modification de chronique"}`),
			BodyContains: []string{"Modification d'une chronique de paiement, requête : Chronique de paiement introuvable"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(ptID),
			Status:       http.StatusOK,
			Sent:         []byte(`{"name":"Modification de chronique"}`),
			BodyContains: []string{"PaymentType", `"name":"Modification de chronique"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.PUT("/api/payment_types/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "ModifyPaymentType") {
		t.Error(r)
	}
}

// deletePaymentTypeTest check route is protected and ratios correctly deleted
func deletePaymentTypeTest(e *httpexpect.Expect, t *testing.T, ptID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Suppression d'une chronique de paiement, requête : Chronique de paiement introuvable"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(ptID),
			Status:       http.StatusOK,
			BodyContains: []string{"Chronique supprimée"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.DELETE("/api/payment_types/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DeletePaymentType") {
		t.Error(r)
	}
}
