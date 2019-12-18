package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testSummaries(t *testing.T) {
	t.Run("Summaries", func(t *testing.T) {
		multiannualProgrammationTest(testCtx.E, t)
		annualProgrammationTest(testCtx.E, t)
		initAnnualProgrammationTest(testCtx.E, t)
		programmingPrevisionTest(testCtx.E, t)
		actionProgrammationTest(testCtx.E, t)
		actionProgrammationAndYearsTest(testCtx.E, t)
		actionCommitmentTest(testCtx.E, t)
		detailedActionCommitmentTest(testCtx.E, t)
		detailedActionPaymentTest(testCtx.E, t)
		detailedStatActionPaymentTest(testCtx.E, t)
		actionPaymentTest(testCtx.E, t)
		statActionPaymentTest(testCtx.E, t)
		statCurrentYearPaymentTest(testCtx.E, t)
	})
}

// multiannualProgrammationTest check route is protected and datas sent has got items and number of lines.
func multiannualProgrammationTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:         testCtx.User.Token,
			Status:        http.StatusOK,
			BodyContains:  []string{"MultiannualProg"},
			CountItemName: `"number"`,
			ArraySize:     318},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/summaries/multiannual_programmation").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "MultiannuelProgrammation") {
		t.Error(r)
	}
}

// annualProgrammationTest check route is protected and datas sent has got items and number of lines.
func annualProgrammationTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:  testCtx.User.Token,
			Status: http.StatusOK,
			Param:  "2018",
			BodyContains: []string{"AnnualProgrammation", "ImportLog", "operation_number",
				"name", "step_name", "category_name", "date", "programmings", "total_programmings",
				"state_ratio", "commitment", "pendings"},
			CountItemName: `"name"`,
			ArraySize:     117},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/summaries/annual_programmation").WithQuery("year", tc.Param).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "AnnualProgrammation") {
		t.Error(r)
	}
}

// initAnnualProgrammationTest check route is protected and datas sent has got items and number of lines.
func initAnnualProgrammationTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase, // O : bad token
		{
			Token:  testCtx.User.Token,
			Status: http.StatusOK,
			Param:  "2018",
			BodyContains: []string{"AnnualProgrammation", "ImportLog",
				"operation_number", "name", "step_name", "category_name", "date",
				"programmings", "total_programmings", "state_ratio", "commitment",
				"pendings", "BudgetCredits", "ProgrammingsYears"},
			CountItemName: `"name"`,
			ArraySize:     117}, // 1 : tets with 2018 year
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/summaries/annual_programmation/init").WithQuery("year", tc.Param).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "AnnualProgrammation") {
		t.Error(r)
	}
}

// programmingPrevisionTest check route is protected and datas sent has got items and number of lines.
func programmingPrevisionTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:  testCtx.User.Token,
			Status: http.StatusOK,
			Param:  "2018",
			BodyContains: []string{"ProgrammingsPrevision", "number", "name",
				"programmings", "prevision"},
			CountItemName: `"number"`,
			ArraySize:     127},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/summaries/programmation_prevision").WithQuery("year", tc.Param).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "ProgrammingAndPrevision") {
		t.Error(r)
	}
}

// actionProgrammationTest check route is protected and datas sent has got items and number of lines.
func actionProgrammationTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:         testCtx.User.Token,
			Status:        http.StatusOK,
			Param:         "2018",
			BodyContains:  []string{"BudgetProgrammation", "action_code", "action_name", "value"},
			CountItemName: `"action_code"`,
			ArraySize:     26},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/summaries/budget_action_programmation").WithQuery("year", tc.Param).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "ActionProgrammation") {
		t.Error(r)
	}
}

// actionProgrammationAndYearsTest check route is protected and datas sent has
// got items and number of lines.
func actionProgrammationAndYearsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase, // 0 : bad token
		{
			Token:  testCtx.User.Token,
			Status: http.StatusOK,
			Param:  "2018",
			BodyContains: []string{"BudgetProgrammation", "action_code", "action_name",
				"value", "ProgrammingsYears"},
			CountItemName: `"action_code"`,
			ArraySize:     26}, // 1 : test with 2018 year
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/summaries/budget_action_programmation_years").
			WithQuery("year", tc.Param).WithHeader("Authorization", "Bearer "+tc.Token).
			Expect()
	}
	for _, r := range chkTestCases(testCases, f, "ActionProgrammationAndYears") {
		t.Error(r)
	}
}

// actionCommitmentTest check route is protected and datas sent has got items and number of lines.
func actionCommitmentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:  testCtx.User.Token,
			Status: http.StatusOK,
			Param:  "2019",
			BodyContains: []string{"CommitmentPerBudgetAction", "chapter", "sector", "subfunction", "program",
				"action", "action_name", "y0", "y1", "y2", "y3"},
			CountItemName: `"chapter"`,
			ArraySize:     46},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/summaries/commitment_per_budget_action").WithQuery("FirstYear", tc.Param).
			WithHeader("Authorization", "Bearer "+tc.Token).
			Expect()
	}
	for _, r := range chkTestCases(testCases, f, "ActionCommitment") {
		t.Error(r)
	}
}

// detailedActionCommitmentTest check route is protected and datas sent has got items and number of lines.
func detailedActionCommitmentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:  testCtx.User.Token,
			Status: http.StatusOK,
			Param:  "2019",
			BodyContains: []string{"DetailedCommitmentPerBudgetAction", "chapter",
				"sector", "subfunction", "program", "action", "action_name", "number", "name",
				"y0", "y1", "y2", "y3"},
			CountItemName: `"action_name"`,
			ArraySize:     150},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/summaries/detailed_commitment_per_budget_action").
			WithQuery("FirstYear", tc.Param).WithHeader("Authorization", "Bearer "+tc.Token).
			Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DetailedActionCommitment") {
		t.Error(r)
	}
}

// detailedActionPaymentTest check route is protected and datas sent has got items and number of lines.
func detailedActionPaymentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:  testCtx.User.Token,
			Status: http.StatusOK,
			ID:     "5",
			Param:  "2019",
			BodyContains: []string{"DetailedPaymentPerBudgetAction", "chapter",
				"sector", "subfunction", "program", "action", "action_name", "number",
				"name", "y1", "y2", "y3"},
			CountItemName: `"chapter"`,
			ArraySize:     433},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/summaries/detailed_payment_per_budget_action").
			WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("DefaultPaymentTypeId", tc.ID).WithQuery("FirstYear", tc.Param).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DetailedActionPayment") {
		t.Error(r)
	}
}

// detailedStatActionPaymentTest check route is protected and datas sent has got items and number of lines.
func detailedStatActionPaymentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:  testCtx.User.Token,
			Status: http.StatusOK,
			ID:     "5",
			Param:  "2019",
			BodyContains: []string{"DetailedPaymentPerBudgetAction", "chapter",
				"sector", "subfunction", "program", "action", "action_name", "number",
				"name", "y1", "y2", "y3"},
			CountItemName: `"chapter"`,
			ArraySize:     433},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/summaries/statistical_detailed_payment_per_budget_action").
			WithHeader("Authorization", "Bearer "+tc.Token).WithQuery("DefaultPaymentTypeId", tc.ID).
			WithQuery("FirstYear", tc.Param).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DetailedStatActionPayment") {
		t.Error(r)
	}
}

// actionPaymentTest check route is protected and datas sent has got items and number of lines.
func actionPaymentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:  testCtx.User.Token,
			Status: http.StatusOK,
			ID:     "5",
			BodyContains: []string{"PaymentPerBudgetAction", "chapter",
				"sector", "subfunction", "program", "action", "action_name", "y1",
				"y2", "y3"},
			CountItemName: `"chapter"`,
			ArraySize:     58},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/summaries/payment_per_budget_action").
			WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("DefaultPaymentTypeId", tc.ID).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "ActionPayment") {
		t.Error(r)
	}
}

// statActionPaymentTest check route is protected and datas sent has got items and number of lines.
func statActionPaymentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:  testCtx.User.Token,
			Status: http.StatusOK,
			ID:     "5",
			Param:  "2019",
			BodyContains: []string{"PaymentPerBudgetAction",
				`"chapter":908,"sector":"TC","subfunction":"811","program":"381006",` +
					`"action":"381006015","action_name":"Métro","y1"`, `"y2"`, `"y3"`},
			CountItemName: `"chapter"`,
			ArraySize:     58},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/summaries/statistical_payment_per_budget_action").
			WithHeader("Authorization", "Bearer "+tc.Token).WithQuery("DefaultPaymentTypeId", tc.ID).
			WithQuery("FirstYear", tc.Param).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "StatActionPayment") {
		t.Error(r)
	}
}

// statCurrentYearPaymentTest check route is protected and datas sent has got items.
func statCurrentYearPaymentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:  testCtx.User.Token,
			Status: http.StatusOK,
			ID:     "5",
			Param:  "2019",
			BodyContains: []string{`"StatisticalCurrentYearPaymentPerAction":` +
				`[{"chapter":907,"sector":"EAE","subfunction":"77","program":"477003",` +
				`"action":"477003011","action_name":"Intégration environnementale des ` +
				`infrastructures de transport","prev":10668159.432043333,"payment":null`},
			CountItemName: `"chapter"`,
			ArraySize:     53},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/summaries/statistical_current_year_payment_per_budget_action").
			WithHeader("Authorization", "Bearer "+tc.Token).WithQuery("DefaultPaymentTypeId", tc.ID).
			WithQuery("Year", tc.Param).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "StatCurrentYearPayment") {
		t.Error(r)
	}
}
