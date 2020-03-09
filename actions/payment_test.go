package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testPayment(t *testing.T) {
	t.Run("Payment", func(t *testing.T) {
		getAllPaymentsTest(testCtx.E, t)
		getFcPaymentTest(testCtx.E, t)
		getPaymentsPerMonthTest(testCtx.E, t)
		getPrevisionRealizedTest(testCtx.E, t)
		getCumulatedMonthPaymentTest(testCtx.E, t)
		batchPaymentsTest(testCtx.E, t)
	})
}

// getAllPaymentsTest check route is protected and payments correctly sent.
func getAllPaymentsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:  testCtx.User.Token,
			Param:  "?year=2019",
			Status: http.StatusOK,
			BodyContains: []string{"PaymentsPerMonth", "MonthCumulatedPayment", `"Beneficiary":[`,
				`"PaymentType":`, `"PaymentCreditJournal":[`, `"PaymentCredit":[`, `"PaymentNeed":[`},
			CountItemName: `"id"`,
			ArraySize:     533},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/payments").WithQueryString("Param").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetAllPayments") {
		t.Error(r)
	}
}

// getFcPaymentTest check route is protected and payments correctly sent.
func getFcPaymentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			ID:           "0",
			Status:       http.StatusOK,
			BodyContains: []string{`"Payment":[]`}},
		{
			Token:         testCtx.User.Token,
			ID:            "219",
			Status:        http.StatusOK,
			BodyContains:  []string{"Payment"},
			CountItemName: `"id"`,
			ArraySize:     9},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/physical_ops/152/financial_commitments/"+tc.ID+"/payments").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetFcPayment") {
		t.Error(r)
	}
}

// getPaymentsPerMonthTest check if route is protected and payments per month correctly sent.
func getPaymentsPerMonthTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:         testCtx.User.Token,
			Param:         "2018",
			Status:        http.StatusOK,
			BodyContains:  []string{"PaymentsPerMonth", `"year":2017`, `"month":1`, `"value":`},
			CountItemName: "year",
			ArraySize:     15},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/payments/month").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("year", tc.Param).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetPaymentsPerMonth") {
		t.Error(r)
	}
}

// getPrevisionRealizedTest check if route is protected and sent datas are correct.
func getPrevisionRealizedTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,

		{Token: testCtx.User.Token,
			Param:        "",
			ID:           "1",
			Status:       http.StatusBadRequest,
			BodyContains: []string{"Prévu réalisé erreur sur year"}},
		{
			Token:        testCtx.User.Token,
			Param:        "2017",
			ID:           "1",
			Status:       http.StatusOK,
			BodyContains: []string{`"PaymentPrevisionAndRealized":[]`}},
		{
			Token:  testCtx.User.Token,
			Param:  "2017",
			ID:     "4",
			Status: http.StatusOK,
			//cSpell:disable
			BodyContains: []string{`{"PaymentPrevisionAndRealized":[{"name":"RATP REGIE AUTONOME DES TRANSPORTS PARISIENS","prev_payment":14879877199,"payment":21297216350},`},
			//cSpell:enable
			CountItemName: `"name"`,
			ArraySize:     386},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/payments/prevision_realized").
			WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("year", tc.Param).WithQuery("paymentTypeId", tc.ID).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetPrevisionRealized") {
		t.Error(r)
	}
}

// getCumulatedMonthPaymentTest check if route is protected and datas has good size.
func getCumulatedMonthPaymentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:         testCtx.User.Token,
			ID:            "",
			Status:        http.StatusOK,
			BodyContains:  []string{"MonthCumulatedPayment", `"cumulated":7626791.01`},
			CountItemName: `year`,
			ArraySize:     132},
		{
			Token:         testCtx.User.Token,
			ID:            "8",
			Status:        http.StatusOK,
			BodyContains:  []string{"MonthCumulatedPayment", `"year":2007`, `"month":1`, `"cumulated":1440789.04`},
			CountItemName: `year`,
			ArraySize:     110},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/payments/month_cumulated").
			WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("beneficiaryId", tc.ID).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetFcPayment") {
		t.Error(r)
	}
}

// batchPaymentsTest check route is protected and a small batch doesn't raise error
func batchPaymentsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{Token: testCtx.Admin.Token,
			Status: http.StatusOK,
			//cSpell:disable
			Sent: []byte(`{"Payment":[{"coriolis_year":"2000","coriolis_egt_code":"DAVT","coriolis_egt_num":"103323","coriolis_egt_line":"501","date":43168,"number":"4784","value":445899.87,"cancelled_value":445899.87,"beneficiary_code":14154,"receipt_date":null},
			{"coriolis_year":"2000","coriolis_egt_code":"DAVT","coriolis_egt_num":"103323","coriolis_egt_line":"504","date":43132,"number":"6078","value":445899.87,"cancelled_value":0,"beneficiary_code":14154},
			{"coriolis_year":"2003","coriolis_egt_code":"P0385","coriolis_egt_num":"132770","coriolis_egt_line":"501","date":43132,"number":"1667","value":94254.15,"cancelled_value":0,"beneficiary_code":14154},
			{"coriolis_year":"2003","coriolis_egt_code":"P0385","coriolis_egt_num":"132770","coriolis_egt_line":"501","date":43132,"number":"1668","value":183796.82,"cancelled_value":0,"beneficiary_code":14154},
			{"coriolis_year":"2003","coriolis_egt_code":"P0385","coriolis_egt_num":"132770","coriolis_egt_line":"501","date":43132,"number":"1669","value":89345.01,"cancelled_value":0,"beneficiary_code":14154},
			{"coriolis_year":"2003","coriolis_egt_code":"P0385","coriolis_egt_num":"132770","coriolis_egt_line":"501","date":43082,"number":"1670","value":99719.88,"cancelled_value":0,"beneficiary_code":14154},
			{"coriolis_year":"2005","coriolis_egt_code":"P0534","coriolis_egt_num":"162726","coriolis_egt_line":"3","date":43082,"number":"47718","value":430151.97,"cancelled_value":0,"beneficiary_code":14154},
			{"coriolis_year":"2005","coriolis_egt_code":"P0534","coriolis_egt_num":"162726","coriolis_egt_line":"3","date":43082,"number":"47719","value":351340.16,"cancelled_value":0,"beneficiary_code":14154},
			{"coriolis_year":"2005","coriolis_egt_code":"P0534","coriolis_egt_num":"162726","coriolis_egt_line":"3","date":42867,"number":"47720","value":537107.87,"cancelled_value":0,"beneficiary_code":14154},
			{"coriolis_year":"2005","coriolis_egt_code":"P0852","coriolis_egt_num":"170678","coriolis_egt_line":"1","date":43215,"number":"15390","value":5623.8,"cancelled_value":0,"beneficiary_code":22844,"receipt_date":43200}]}`),
			BodyContains: []string{"Paiements importés"}},
		//cSpell:enable
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/payments").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "BatchPayments") {
		t.Error(r)
	}
}
