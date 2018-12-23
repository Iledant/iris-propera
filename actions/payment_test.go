package actions

import (
	"net/http"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func TestPayment(t *testing.T) {
	TestCommons(t)
	t.Run("Payment", func(t *testing.T) {
		getFcPaymentTest(testCtx.E, t)
		getPaymentsPerMonthTest(testCtx.E, t)
		getPrevisionRealizedTest(testCtx.E, t)
		getCumulatedMonthPaymentTest(testCtx.E, t)
		batchPaymentsTest(testCtx.E, t)
	})
}

// getFcPaymentTest check route is protected and payments correctly sent.
func getFcPaymentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", ID: "219", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusOK,
			BodyContains: []string{`"Payment":null`}},
		{Token: testCtx.User.Token, ID: "219", Status: http.StatusOK,
			BodyContains: []string{"Payment"}, ArraySize: 9},
	}
	for i, tc := range testCases {
		response := e.GET("/api/physical_ops/152/financial_commitments/"+tc.ID+"/payments").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetFcPayment[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetFcPayment[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetFcPayment[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// getPaymentsPerMonthTest check if route is protected and payments per month correctly sent.
func getPaymentsPerMonthTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Param: "2018", Status: http.StatusOK,
			BodyContains: []string{"PaymentsPerMonth", `"year":2017`, `"month":1`, `"value":`}, ArraySize: 15},
	}
	for i, tc := range testCases {
		response := e.GET("/api/payments/month").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("year", tc.Param).Expect()
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetPaymentsPerMonth[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetPaymentsPerMonth[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `year`)
			if count != tc.ArraySize {
				t.Errorf("\nGetPaymentsPerMonth[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// getPrevisionRealizedTest check if route is protected and sent datas are correct.
func getPrevisionRealizedTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Param: "", ID: "1", Status: http.StatusBadRequest,
			BodyContains: []string{"Prévu réalisé erreur sur year"}, ArraySize: 0},
		{Token: testCtx.User.Token, Param: "2017", ID: "1", Status: http.StatusOK,
			BodyContains: []string{`"PaymentPrevisionAndRealized":null`}, ArraySize: 0},
		{Token: testCtx.User.Token, Param: "2017", ID: "4", Status: http.StatusOK,
			//cSpell:disable
			BodyContains: []string{`{"PaymentPrevisionAndRealized":[{"name":"RATP REGIE AUTONOME DES TRANSPORTS PARISIENS","prev_payment":14879877199,"payment":21297216350},`},
			//cSpell:enable
			ArraySize: 386},
	}
	for i, tc := range testCases {
		response := e.GET("/api/payments/prevision_realized").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("year", tc.Param).WithQuery("paymentTypeId", tc.ID).Expect()
		content := string(response.Content)
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetPrevisionRealized[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetPrevisionRealized(%d) :\n attendu -> \"%s\"\n reçu <-\"%s\"\n", i, s, content)
			}
		}
		if tc.ArraySize != 0 {
			count := strings.Count(content, "\"name\"")
			if count != tc.ArraySize {
				t.Errorf("\nGetPrevisionRealized(%d) :\n nombre attendu de champs -> \"%d\"\n nombre reçu de champ <-\"%d\"\n",
					i, tc.ArraySize, count)
			}
		}
	}
}

// getCumulatedMonthPaymentTest check if route is protected and datas has good size.
func getCumulatedMonthPaymentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, ID: "", Status: http.StatusOK,
			BodyContains: []string{"MonthCumulatedPayment", `"cumulated":7626791.01`}, ArraySize: 132},
		{Token: testCtx.User.Token, ID: "8", Status: http.StatusOK,
			BodyContains: []string{"MonthCumulatedPayment", `"year":2007`, `"month":1`, `"cumulated":1440789.04`}, ArraySize: 110},
	}
	for i, tc := range testCases {
		response := e.GET("/api/payments/month_cumulated").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("beneficiaryId", tc.ID).Expect()
		content := string(response.Content)
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetFcPayment[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetCumulatedMonthPayment[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `year`)
			if count != tc.ArraySize {
				t.Errorf("\nGetCumulatedMonthPayment[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// batchPaymentsTest check route is protected and a small batch doesn't raise error
func batchPaymentsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			//cSpell:disable
			Sent: []byte(`{"Payment":[{"coriolis_year":"2000","coriolis_egt_code":"DAVT","coriolis_egt_num":"103323","coriolis_egt_line":"501","date":"2018-03-09T00:00:00Z","number":"4784","value":445899.87,"cancelled_value":445899.87,"beneficiary_code":14154},
			{"coriolis_year":"2000","coriolis_egt_code":"DAVT","coriolis_egt_num":"103323","coriolis_egt_line":"504","date":"2018-02-01T00:00:00Z","number":"6078","value":445899.87,"cancelled_value":0,"beneficiary_code":14154},
			{"coriolis_year":"2003","coriolis_egt_code":"P0385","coriolis_egt_num":"132770","coriolis_egt_line":"501","date":"2018-02-01T00:00:00Z","number":"1667","value":94254.15,"cancelled_value":0,"beneficiary_code":14154},
			{"coriolis_year":"2003","coriolis_egt_code":"P0385","coriolis_egt_num":"132770","coriolis_egt_line":"501","date":"2018-02-01T00:00:00Z","number":"1668","value":183796.82,"cancelled_value":0,"beneficiary_code":14154},
			{"coriolis_year":"2003","coriolis_egt_code":"P0385","coriolis_egt_num":"132770","coriolis_egt_line":"501","date":"2018-02-01T00:00:00Z","number":"1669","value":89345.01,"cancelled_value":0,"beneficiary_code":14154},
			{"coriolis_year":"2003","coriolis_egt_code":"P0385","coriolis_egt_num":"132770","coriolis_egt_line":"501","date":"2017-12-13T00:00:00Z","number":"1670","value":99719.88,"cancelled_value":0,"beneficiary_code":14154},
			{"coriolis_year":"2005","coriolis_egt_code":"P0534","coriolis_egt_num":"162726","coriolis_egt_line":"3","date":"2017-12-13T00:00:00Z","number":"47718","value":430151.97,"cancelled_value":0,"beneficiary_code":14154},
			{"coriolis_year":"2005","coriolis_egt_code":"P0534","coriolis_egt_num":"162726","coriolis_egt_line":"3","date":"2017-12-13T00:00:00Z","number":"47719","value":351340.16,"cancelled_value":0,"beneficiary_code":14154},
			{"coriolis_year":"2005","coriolis_egt_code":"P0534","coriolis_egt_num":"162726","coriolis_egt_line":"3","date":"2017-05-12T00:00:00Z","number":"47720","value":537107.87,"cancelled_value":0,"beneficiary_code":14154},
			{"coriolis_year":"2005","coriolis_egt_code":"P0852","coriolis_egt_num":"170678","coriolis_egt_line":"1","date":"2018-01-25T00:00:00Z","number":"15390","value":5623.8,"cancelled_value":0,"beneficiary_code":22844}]}`),
			BodyContains: []string{"Paiements importés"}},
		//cSpell:enable
	}
	for i, tc := range testCases {
		response := e.POST("/api/payments").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).
			Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nBatchPayments[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nBatchPayments[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}
