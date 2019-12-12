package actions

import (
	"net/http"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testProgramming(t *testing.T) {
	t.Run("Programming", func(t *testing.T) {
		getProgrammingsTest(testCtx.E, t)
		getProgrammingsYearsTest(testCtx.E, t)
		batchProgrammingsTest(testCtx.E, t)
	})
}

// getProgrammingsTest check route is protected and programmings correctly sent.
func getProgrammingsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{
			Token:        "fake",
			Param:        "2018",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{
			Token:        testCtx.Admin.Token,
			Param:        "2018",
			Status:       http.StatusOK,
			BodyContains: []string{"Programmings", `"PrevCommitmentTotal":96730644861`},
			ArraySize:    626},
	}
	for i, tc := range testCases {
		response := e.GET("/api/programmings").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("year", tc.Param).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetProgrammings[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetProgrammings[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetProgrammings[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// getProgrammingsYearsTest check route is protected and programmings correctly sent.
func getProgrammingsYearsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{`{"ProgrammingsYears":[{"year":2018}]}`}},
	}
	for i, tc := range testCases {
		response := e.GET("/api/programmings/years").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetProgrammingsYears[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetProgrammingsYears[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// batchProgrammingsTest check route is protected and return successful.
func batchProgrammingsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, Sent: []byte(`{Pend}`),
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusInternalServerError, Sent: []byte(`{Pend}`),
			BodyContains: []string{"Batch programmation, décodage : "}},
		//cSpell:disable
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Sent: []byte(`{"Programmings": [
			{"physical_op_id":9,"year":2018,"value":100000000,
			"commission_id":7,"total_value":null,"state_ratio":null},
			{"physical_op_id":10,"year":2018,"value":200000000,
			"commission_id":8,"total_value":400000000,"state_ratio":null},
			{"physical_op_id":14 ,"year":2018,"value":300000000,
			"commission_id":3,"total_value":600000000,"state_ratio":0.35}],
			"year":2018}`),
			BodyContains: []string{"Programmings", `"physical_op_id":9`, `"physical_op_id":10`,
				`"physical_op_id":14`, `"value":200000000`, `"commission_id":8`, `"total_value":400000000`,
				`"total_value":null`, `"state_ratio":null`, `"state_ratio":0.35`}},
	}
	//cSpell:enable
	for i, tc := range testCases {
		response := e.POST("/api/programmings/array").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nBatchProgrammings[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nBatchProgrammings[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}
