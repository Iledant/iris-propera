package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testTodayMessage(t *testing.T) {
	t.Run("TodayMessage", func(t *testing.T) {
		getTodayMessageTest(testCtx.E, t)
		setTodayMessageTest(testCtx.E, t)
		getHomeDatasTest(testCtx.E, t)
	})
}

// getTodayMessageTest check route is protected and datas sent has got items and number of lines.
func getTodayMessageTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusOK,
			BodyContains: []string{"TodayMessage", "title", "text"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/today_message").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetTodayMessage") {
		t.Error(r)
	}
}

// setTodayMessageTest check route is protected and datas sent has got items and number of lines.
func setTodayMessageTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusOK,
			Sent:   []byte(`{"title":"Essai de titre","text":"Essai de texte"}`),
			BodyContains: []string{"TodayMessage", `"title":"Essai de titre"`,
				`"text":"Essai de texte"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/today_message").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "SetTodayMessage") {
		t.Error(r)
	}
}

// getHomeDatasTest check route is protected and all kinds of datas are sent back.
func getHomeDatasTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:  testCtx.User.Token,
			Status: http.StatusOK,
			BodyContains: []string{"TodayMessage", "Event", "BudgetCredits",
				"FinancialCommitmentsPerMonth", "ProgrammingsPerMonth",
				"PaymentsPerMonth", "PaymentDemandsStock", "CsfWeekTrend", `"FlowStockDelays":`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/home").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetHomeDatas") {
		t.Error(r)
	}
}
