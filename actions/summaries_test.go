package actions

import (
	"net/http"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func TestSummaries(t *testing.T) {
	TestCommons(t)
	t.Run("Summaries", func(t *testing.T) {
		multiannualProgrammationTest(testCtx.E, t)
		annualProgrammationTest(testCtx.E, t)
		programmingPrevisionTest(testCtx.E, t)
		actionProgrammationTest(testCtx.E, t)
		actionCommitmentTest(testCtx.E, t)
		detailedActionCommitmentTest(testCtx.E, t)
		detailedActionPaymentTest(testCtx.E, t)
		actionPaymentTest(testCtx.E, t)
	})
}

// multiannualProgrammationTest check route is protected and datas sent has got items and number of lines.
func multiannualProgrammationTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"MultiannualProgrammation"}, ArraySize: 318},
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
				t.Errorf("\nMultiannualProgrammation[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// annualProgrammationTest check route is protected and datas sent has got items and number of lines.
func annualProgrammationTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"AnnualProgrammation", "ImportLog", "operation_number",
				"name", "step_name", "category_name", "date", "programmings", "total_programmings",
				"state_ratio", "commitment", "pendings"}, ArraySize: 117},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/annual_programmation").
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
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"ProgrammingsPrevision", "number", "name",
				"programmings", "prevision"}, ArraySize: 127},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/programmation_prevision").
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
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"BudgetProgrammation", "action_code", "action_name", "value"}, ArraySize: 26},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/budget_action_programmation").
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
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"CommitmentPerBudgetAction", "chapter", "sector", "subfunction", "program",
				"action", "action_name", "y0", "y1", "y2", "y3"}, ArraySize: 46},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/commitment_per_budget_action").WithHeader("Authorization", "Bearer "+tc.Token).
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
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("CommitmentPerBudgetAction").Array().Length().Equal(tc.ArraySize)
		}
	}
}

// detailedActionCommitmentTest check route is protected and datas sent has got items and number of lines.
func detailedActionCommitmentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"DetailedCommitmentPerBudgetAction", "chapter",
				"sector", "subfunction", "program", "action", "action_name", "number", "name",
				"y0", "y1", "y2", "y3"}, ArraySize: 150},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/detailed_commitment_per_budget_action").WithHeader("Authorization", "Bearer "+tc.Token).
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
		{Token: testCtx.User.Token, Status: http.StatusOK, ID: "5",
			BodyContains: []string{"DetailedPaymentPerBudgetAction", "chapter", "sector", "subfunction", "program",
				"action", "action_name", "number", "name", "y1", "y2", "y3"}, ArraySize: 433},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/detailed_payment_per_budget_action").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("DefaultPaymentTypeId", tc.ID).Expect()
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
