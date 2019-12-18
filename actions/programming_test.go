package actions

import (
	"net/http"
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
		notLoggedTestCase,
		{
			Token:         testCtx.Admin.Token,
			Param:         "2018",
			Status:        http.StatusOK,
			BodyContains:  []string{"Programmings", `"PrevCommitmentTotal":96730644861`},
			CountItemName: `"id"`,
			ArraySize:     626},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/programmings").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("year", tc.Param).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetProgrammings") {
		t.Error(r)
	}
}

// getProgrammingsYearsTest check route is protected and programmings correctly sent.
func getProgrammingsYearsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusOK,
			BodyContains: []string{`{"ProgrammingsYears":[{"year":2018}]}`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/programmings/years").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetProgrammingsYears") {
		t.Error(r)
	}
}

// batchProgrammingsTest check route is protected and return successful.
func batchProgrammingsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusUnauthorized,
			Sent:         []byte(`{Pend}`),
			BodyContains: []string{"Droits administrateur requis"}},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{Pend}`),
			BodyContains: []string{"Batch programmation, décodage : "}},
		//cSpell:disable
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusOK,
			Sent: []byte(`{"Programmings": [
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
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/programmings/array").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "BatchProgrammings") {
		t.Error(r)
	}
}
