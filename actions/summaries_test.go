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
			WithHeader("Authorization", "Bearer "+tc.Token).
			Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("MultiannualProgrammation[%d] : attendu %s et reçu \n%s", i, s, content)
			}
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("MultiannualProgrammation").Array().Length().Equal(tc.ArraySize)
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
				t.Errorf("AnnualProgrammation[%d] : attendu %s et reçu \n%s", i, s, content)
			}
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("AnnualProgrammation").Array().Length().Equal(tc.ArraySize)
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
				t.Errorf("ProgrammingPrevision[%d] : attendu %s et reçu \n%s", i, s, content)
			}
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("ProgrammingsPrevision").Array().Length().Equal(tc.ArraySize)
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
				t.Errorf("ActionProgrammation[%d] : attendu %s et reçu \n%s", i, s, content)
			}
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("BudgetProgrammation").Array().Length().Equal(tc.ArraySize)
		}
	}
}

// actionCommitmentTest check route is protected and datas sent has got items and number of lines.
func actionCommitmentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"CommitmentPerBudgetAction", "chapter", "sector", "subfunction", "program",
				"action", "action_name", "y2018", "y2019", "y2020", "y2021"}, ArraySize: 46},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/commitment_per_budget_action").WithHeader("Authorization", "Bearer "+tc.Token).
			Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("ActionCommitment[%d] : attendu %s et reçu \n%s", i, s, content)
			}
		}
		response.Status(tc.Status)
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
				"y2018", "y2019", "y2020", "y2021"}, ArraySize: 185},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/detailed_commitment_per_budget_action").WithHeader("Authorization", "Bearer "+tc.Token).
			Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("DetailedActionCommitment[%d] : attendu %s et reçu \n%s", i, s, content)
			}
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("DetailedCommitmentPerBudgetAction").Array().Length().Equal(tc.ArraySize)
		}
	}
}

// detailedActionPaymentTest check route is protected and datas sent has got items and number of lines.
func detailedActionPaymentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK, ID: "5",
			BodyContains: []string{"DetailedPaymentPerBudgetAction", "chapter", "sector", "subfunction", "program",
				"action", "action_name", "number", "name", "y2019", "y2020", "y2021"}, ArraySize: 433},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/detailed_payment_per_budget_action").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("DefaultPaymentTypeId", tc.ID).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("DetailedActionPayment[%d] : attendu %s et reçu \n%s", i, s, content)
			}
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("DetailedPaymentPerBudgetAction").Array().Length().Equal(tc.ArraySize)
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
				"sector", "subfunction", "program", "action", "action_name", "y2019",
				"y2020", "y2021"}, ArraySize: 58},
	}
	for i, tc := range testCases {
		response := e.GET("/api/summaries/payment_per_budget_action").
			WithHeader("Authorization", "Bearer "+tc.Token).WithQuery("DefaultPaymentTypeId", tc.ID).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("DetailedActionPayment[%d] : attendu %s et reçu \n%s", i, s, content)
			}
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("PaymentPerBudgetAction").Array().Length().Equal(tc.ArraySize)
		}
	}
}
