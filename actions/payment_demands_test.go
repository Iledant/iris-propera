package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testPaymentDemands(t *testing.T) {
	t.Run("Payment", func(t *testing.T) {
		batchPaymentDemandsTest(testCtx.E, t)
		updatePaymentDemandsTest(testCtx.E, t)
		getAllPaymentDemandsTest(testCtx.E, t)
	})
}

// batchPaymentDemandsTest check route is protected and a small batch doesn't raise error
func batchPaymentDemandsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase, // 0 unauthorized
		{Token: testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{"PaymentDemand":{"iris_code":"1900001","iris_name":"Etudes T9","commitment_date":43168,"beneficiary_code":1989,"demand_number":1,"demand_date":43268,"receipt_date":43278,"demand_value":10000000,"csf_date":null,"csf_comment":null,"demand_status":null,"status_comment":null}]}`),
			BodyContains: []string{"Batch de demandes de paiement, décodage"}}, // 1 bad json
		{Token: testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"PaymentDemand":[{"iris_name":"Etudes T9","commitment_date":43168,"beneficiary_code":1989,"demand_number":1,"demand_date":43268,"receipt_date":43278,"demand_value":10000000,"csf_date":null,"csf_comment":null,"demand_status":null,"status_comment":null}]}`),
			BodyContains: []string{"Batch de demandes de paiement, requête : ligne 1 iris_code vide"}}, // 2 iris_code empty
		{Token: testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"PaymentDemand":[{"iris_code":"1900001","commitment_date":43168,"beneficiary_code":1989,"demand_number":1,"demand_date":43268,"receipt_date":43278,"demand_value":10000000,"csf_date":null,"csf_comment":null,"demand_status":null,"status_comment":null}]}`),
			BodyContains: []string{"Batch de demandes de paiement, requête : ligne 1 iris_name vide"}}, //3 iris_name empty
		{Token: testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"PaymentDemand":[{"iris_code":"1900001","iris_name":"Etudes T9","beneficiary_code":1989,"demand_number":1,"demand_date":43268,"receipt_date":43278,"demand_value":10000000,"csf_date":null,"csf_comment":null,"demand_status":null,"status_comment":null}]}`),
			BodyContains: []string{"Batch de demandes de paiement, requête : ligne 1 commitment_date vide"}}, // 4 commiment_date empty
		{Token: testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"PaymentDemand":[{"iris_code":"1900001","iris_name":"Etudes T9","commitment_date":43168,"demand_number":1,"demand_date":43268,"receipt_date":43278,"demand_value":10000000,"csf_date":null,"csf_comment":null,"demand_status":null,"status_comment":null}]}`),
			BodyContains: []string{"Batch de demandes de paiement, requête : ligne 1 beneficiary_code vide"}}, // 5 beneficiary_code empty
		{Token: testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"PaymentDemand":[{"iris_code":"1900001","iris_name":"Etudes T9","commitment_date":43168,"beneficiary_code":1989,"demand_date":43268,"receipt_date":43278,"demand_value":10000000,"csf_date":null,"csf_comment":null,"demand_status":null,"status_comment":null}]}`),
			BodyContains: []string{"Batch de demandes de paiement, requête : ligne 1 demand_number vide"}}, // 6 demande_number empty
		{Token: testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"PaymentDemand":[{"iris_code":"1900001","iris_name":"Etudes T9","commitment_date":43168,"beneficiary_code":1989,"demand_number":1,"receipt_date":43278,"demand_value":10000000,"csf_date":null,"csf_comment":null,"demand_status":null,"status_comment":null}]}`),
			BodyContains: []string{"Batch de demandes de paiement, requête : ligne 1 demand_date vide"}}, // 7 demand_date empty
		{Token: testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"PaymentDemand":[{"iris_code":"1900001","iris_name":"Etudes T9","commitment_date":43168,"beneficiary_code":1989,"demand_number":1,"demand_date":43268,"demand_value":10000000,"csf_date":null,"csf_comment":null,"demand_status":null,"status_comment":null}]}`),
			BodyContains: []string{"Batch de demandes de paiement, requête : ligne 1 receipt_date vide"}}, // 8 receipt_date empty
		{Token: testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"PaymentDemand":[{"iris_code":"1900001","iris_name":"Etudes T9","commitment_date":43168,"beneficiary_code":1989,"demand_number":1,"demand_date":43268,"receipt_date":43278,"csf_date":null,"csf_comment":null,"demand_status":null,"status_comment":null}]}`),
			BodyContains: []string{"Batch de demandes de paiement, requête : ligne 1 demand_value vide"}}, // 9 demand_value empty
		{Token: testCtx.Admin.Token,
			Status:       http.StatusOK,
			Sent:         []byte(`{"PaymentDemand":[{"iris_code":"1900001","iris_name":"Etudes T9","commitment_date":43168,"beneficiary_code":1989,"demand_number":1,"demand_date":43268,"receipt_date":43278,"demand_value":10000000,"csf_date":null,"csf_comment":null,"demand_status":null,"status_comment":null}]}`),
			BodyContains: []string{"Batch de demande de paiement importé"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/payment_demands").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "BatchPaymentDemands") {
		t.Error(r)
	}
}

// updatePaymentDemandsTest check route is protected and payment demands correctly
// sent back
func updatePaymentDemandsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{Token: testCtx.Admin.Token,
			Status:       http.StatusOK,
			Sent:         []byte(`{"PaymentDemand":{"id":1,"iris_code":"1900001","iris_name":"Etudes T9","commitment_date":"2018-03-09T01:00:00Z","beneficiary_code":1989,"demand_number":1,"demand_date":"2018-06-17T01:00:00Z","receipt_date":"2018-06-27T01:00:00Z","csf_date":null,"csf_comment":null,"demand_status":null,"status_comment":null,"iris_code":"1900001","iris_name":"Etudes T9","beneficiary_code":1989,"demand_number":1,"demand_date":"2018-06-17T01:00:00Z","receipt_date":"2018-06-27T01:00:00Z","demand_value":10000000,"csf_date":null,"csf_comment":null,"demand_status":null,"status_comment":null,"excluded":true,"excluded_comment":"commentaire"}}`),
			BodyContains: []string{`"id":1`, `"commentaire"`, `"excluded":true`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.PUT("/api/payment_demands").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "UpdatePaymentDemands") {
		t.Error(r)
	}
}

// getAllPaymentDemandsTest check route is protected and payment demands correctly sent.
func getAllPaymentDemandsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:         testCtx.User.Token,
			Status:        http.StatusOK,
			BodyContains:  []string{`"PaymentDemand"`, `"iris_name":"Etudes T9"`},
			CountItemName: `"id"`,
			ArraySize:     1},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/payment_demands").WithQueryString("Param").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetAllPaymentDemands") {
		t.Error(r)
	}
}
