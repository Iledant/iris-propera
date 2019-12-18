package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testPreProgramming(t *testing.T) {
	t.Run("PreProgramming", func(t *testing.T) {
		getPreProgrammingsTest(testCtx.E, t)
		batchPreProgrammingsTest(testCtx.E, t)
	})
}

// getPreProgrammingsTest check route is protected and pre programmings correctly sent.
func getPreProgrammingsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:         testCtx.User.Token,
			Param:         "2018",
			Status:        http.StatusOK,
			BodyContains:  []string{"PreProgrammings"},
			CountItemName: `"physical_op_id"`,
			ArraySize:     3},
		{
			Token:         testCtx.Admin.Token,
			Param:         "2018",
			Status:        http.StatusOK,
			BodyContains:  []string{"PreProgrammings"},
			CountItemName: `"physical_op_id"`,
			ArraySize:     622},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/pre_programmings").
			WithHeader("Authorization", "Bearer "+tc.Token).WithQuery("year", tc.Param).
			Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetPreProgrammings") {
		t.Error(r)
	}
}

// batchPreProgrammingsTest check route is protected and return successful.
func batchPreProgrammingsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{Pend}`),
			BodyContains: []string{"Batch préprogrammation, décodage :"}},
		{
			Token:  testCtx.User.Token,
			Status: http.StatusOK,
			//cSpell:disable
			Sent: []byte(`{"PreProgrammings": [
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
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/pre_programmings").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "BatchPreProgrammings") {
		t.Error(r)
	}
}
