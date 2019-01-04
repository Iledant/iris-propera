package actions

import (
	"net/http"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testSummaries(t *testing.T) {
	t.Run("Summaries", func(t *testing.T) {
		multiannualProgrammationTest(testCtx.E, t)
		annualProgrammationTest(testCtx.E, t)
		programmingPrevisionTest(testCtx.E, t)
		actionProgrammationTest(testCtx.E, t)
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
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"MultiannualProg"}, ArraySize: 318},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/multiannual_programmation").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nMultiannuelProgrammation[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nMultiannuelProgrammation[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"number"`)
			if count != tc.ArraySize {
				t.Errorf("\nMultiannualProg[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// annualProgrammationTest check route is protected and datas sent has got items and number of lines.
func annualProgrammationTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK, Param: "2018",
			BodyContains: []string{"AnnualProgrammation", "ImportLog", "operation_number",
				"name", "step_name", "category_name", "date", "programmings", "total_programmings",
				"state_ratio", "commitment", "pendings"}, ArraySize: 117},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/annual_programmation").WithQuery("year", tc.Param).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nAnnualProgrammation[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nAnnualProgrammation[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"name"`)
			if count != tc.ArraySize {
				t.Errorf("\nAnnualProgrammation[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// programmingPrevisionTest check route is protected and datas sent has got items and number of lines.
func programmingPrevisionTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK, Param: "2018",
			BodyContains: []string{"ProgrammingsPrevision", "number", "name",
				"programmings", "prevision"}, ArraySize: 127},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/programmation_prevision").WithQuery("year", tc.Param).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nProgrammingAndPrevision[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nProgrammingAndPrevision[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"number"`)
			if count != tc.ArraySize {
				t.Errorf("\nProgrammingAndPrevision[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// actionProgrammationTest check route is protected and datas sent has got items and number of lines.
func actionProgrammationTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK, Param: "2018",
			BodyContains: []string{"BudgetProgrammation", "action_code", "action_name", "value"}, ArraySize: 26},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/budget_action_programmation").WithQuery("year", tc.Param).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nActionProgrammation[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nActionProgrammation[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"action_code"`)
			if count != tc.ArraySize {
				t.Errorf("\nActionProgrammation[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// actionCommitmentTest check route is protected and datas sent has got items and number of lines.
func actionCommitmentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK, Param: "2019",
			BodyContains: []string{"CommitmentPerBudgetAction", "chapter", "sector", "subfunction", "program",
				"action", "action_name", "y0", "y1", "y2", "y3"}, ArraySize: 46},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/commitment_per_budget_action").WithQuery("FirstYear", tc.Param).
			WithHeader("Authorization", "Bearer "+tc.Token).
			Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nActionCommitment[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nActionCommitment[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"chapter"`)
			if count != tc.ArraySize {
				t.Errorf("\nActionCommitment[%d] :\n  nombre attendu -> %d   nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// detailedActionCommitmentTest check route is protected and datas sent has got items and number of lines.
func detailedActionCommitmentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK, Param: "2019",
			BodyContains: []string{"DetailedCommitmentPerBudgetAction", "chapter",
				"sector", "subfunction", "program", "action", "action_name", "number", "name",
				"y0", "y1", "y2", "y3"}, ArraySize: 150},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/detailed_commitment_per_budget_action").
			WithQuery("FirstYear", tc.Param).WithHeader("Authorization", "Bearer "+tc.Token).
			Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nDetailedActionCommitment[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nDetailedActionCommitment[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"action_name"`)
			if count != tc.ArraySize {
				t.Errorf("\nDetailedActionCommitment[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// detailedActionPaymentTest check route is protected and datas sent has got items and number of lines.
func detailedActionPaymentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK, ID: "5", Param: "2019",
			BodyContains: []string{"DetailedPaymentPerBudgetAction", "chapter", "sector", "subfunction", "program",
				"action", "action_name", "number", "name", "y1", "y2", "y3"}, ArraySize: 433},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/detailed_payment_per_budget_action").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("DefaultPaymentTypeId", tc.ID).WithQuery("FirstYear", tc.Param).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nDetailedActionPayment[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nDetailedActionPayment[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"chapter"`)
			if count != tc.ArraySize {
				t.Errorf("\nDetailedActionPayment[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// detailedStatActionPaymentTest check route is protected and datas sent has got items and number of lines.
func detailedStatActionPaymentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK, ID: "5", Param: "2019",
			BodyContains: []string{"DetailedPaymentPerBudgetAction", "chapter", "sector", "subfunction", "program",
				"action", "action_name", "number", "name", "y1", "y2", "y3"}, ArraySize: 433},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/statistical_detailed_payment_per_budget_action").
			WithHeader("Authorization", "Bearer "+tc.Token).WithQuery("DefaultPaymentTypeId", tc.ID).
			WithQuery("FirstYear", tc.Param).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nDetailedStatActionPayment[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nDetailedStatActionPayment[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"chapter"`)
			if count != tc.ArraySize {
				t.Errorf("\nDetailedStatActionPayment[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// actionPaymentTest check route is protected and datas sent has got items and number of lines.
func actionPaymentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK, ID: "5",
			BodyContains: []string{"PaymentPerBudgetAction", "chapter",
				"sector", "subfunction", "program", "action", "action_name", "y1",
				"y2", "y3"}, ArraySize: 58},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/payment_per_budget_action").
			WithHeader("Authorization", "Bearer "+tc.Token).WithQuery("DefaultPaymentTypeId", tc.ID).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nActionPayment[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nActionPayment[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"chapter"`)
			if count != tc.ArraySize {
				t.Errorf("\nActionPayment[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// statActionPaymentTest check route is protected and datas sent has got items and number of lines.
func statActionPaymentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK, ID: "5", Param: "2019",
			BodyContains: []string{"PaymentPerBudgetAction",
				`"chapter":908,"sector":"TC","subfunction":"811","program":"381006","action":"381006015","action_name":"Métro","y1":46221880.838196725,"y2":20857793.16879844,"y3":18905566.532886185`}, ArraySize: 58},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/statistical_payment_per_budget_action").
			WithHeader("Authorization", "Bearer "+tc.Token).WithQuery("DefaultPaymentTypeId", tc.ID).
			WithQuery("FirstYear", tc.Param).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nStatActionPayment[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nStatActionPayment[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"chapter"`)
			if count != tc.ArraySize {
				t.Errorf("\nStatActionPayment[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// statCurrentYearPaymentTest check route is protected and datas sent has got items.
func statCurrentYearPaymentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK, ID: "5", Param: "2019",
			BodyContains: []string{`"StatisticalCurrentYearPaymentPerAction":[{"chapter":907,"sector":"EAE","subfunction":"77","program":"477003","action":"477003011","action_name":"Intégration environnementale des infrastructures de transport","prev":10668159.432043333,"payment":null`},
			ArraySize:    53},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/statistical_current_year_payment_per_budget_action").
			WithHeader("Authorization", "Bearer "+tc.Token).WithQuery("DefaultPaymentTypeId", tc.ID).
			WithQuery("Year", tc.Param).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nStatCurrentYearPayment[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nStatCurrentYearPayment[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"chapter"`)
			if count != tc.ArraySize {
				t.Errorf("\nStatCurrentYearPayment[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}
