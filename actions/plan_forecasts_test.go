package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testPlanForecasts(t *testing.T) {
	t.Run("PlanForecasts", func(t *testing.T) {
		getPlanForecasts(testCtx.E, t)
	})
}

// getPlanForecasts check if route is protected, params correctly controlled
// and sent datas matches what is needed
func getPlanForecasts(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			Param:        "firstYear=a&lastYear=2026",
			Status:       http.StatusBadRequest,
			BodyContains: []string{`Prévisions de plan, firstYear :`},
		},
		{
			Token:        testCtx.User.Token,
			Param:        "firstYear=2021&lastYear=a",
			Status:       http.StatusBadRequest,
			BodyContains: []string{`Prévisions de plan, lastYear :`},
		},
		{
			Token:        testCtx.User.Token,
			Param:        "firstYear=2021&lastYear=2019",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{`Prévisions de plan, requête : `},
		},
		{
			Token:        testCtx.User.Token,
			Param:        "firstYear=2021&lastYear=2026",
			Status:       http.StatusOK,
			BodyContains: []string{`"PlanForecast":[`},
			IDName:       `"Number":`,
			ArraySize:    54,
		},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/plan_forecasts").WithQueryString(tc.Param).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetPlanForecasts") {
		t.Error(r)
	}
}
