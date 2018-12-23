package actions

import (
	"net/http"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func TestPreProgramming(t *testing.T) {
	TestCommons(t)
	t.Run("PreProgramming", func(t *testing.T) {
		getPreProgrammingsTest(testCtx.E, t)
		batchPreProgrammingsTest(testCtx.E, t)
	})
}

// getPreProgrammingsTest check route is protected and pre programmings correctly sent.
func getPreProgrammingsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Param: "2018", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Param: "2018", Status: http.StatusOK,
			BodyContains: []string{"PreProgrammings"}, ArraySize: 3},
		{Token: testCtx.Admin.Token, Param: "2018", Status: http.StatusOK,
			BodyContains: []string{"PreProgrammings"}, ArraySize: 619},
	}
	for i, tc := range testCases {
		response := e.GET("/api/pre_programmings").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("year", tc.Param).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetPreProgrammings[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetPreProgrammings[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"physical_op_id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetPreProgrammings[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// batchPreProgrammingsTest check route is protected and return successful.
func batchPreProgrammingsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusInternalServerError, Sent: []byte(`{Pend}`),
			BodyContains: []string{"Batch préprogrammation, décodage :"}},
		//cSpell:disable
		{Token: testCtx.User.Token, Status: http.StatusOK, Sent: []byte(`{"PreProgrammings": [
			{"physical_op_id":9,"pre_prog_id":null,"pre_prog_year":2018,"pre_prog_value":100000000,
			"pre_prog_commission_id":7,"pre_prog_total_value":null,"pre_prog_state_ratio":null},
			{"physical_op_id":10,"pre_prog_id":null,"pre_prog_year":2018,"pre_prog_value":200000000,
			"pre_prog_commission_id":8,"pre_prog_total_value":400000000,"pre_prog_state_ratio":null},
			{"physical_op_id":14 ,"pre_prog_id":null,"pre_prog_year":2018,"pre_prog_value":300000000,
			"pre_prog_commission_id":3,"pre_prog_total_value":600000000,"pre_prog_state_ratio":0.35}],
			"year":2018}`),
			BodyContains: []string{"PreProgrammings", `"physical_op_id":9`, `"physical_op_id":10`,
				`"physical_op_id":14`, `"pre_prog_year":2018`, `"pre_prog_value":200000000`,
				`"pre_prog_commission_id":8`, `"pre_prog_total_value":400000000`, `"pre_prog_total_value":null`,
				`"pre_prog_state_ratio":null`, `"pre_prog_state_ratio":0.35`}},
		//cSpell:enable
	}
	for i, tc := range testCases {
		response := e.POST("/api/pre_programmings").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nBatchPreProgrammings[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nBatchPreProgrammings[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}
