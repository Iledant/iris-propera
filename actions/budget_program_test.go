package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestBudgetProgram embeddes all tests for budget programs insuring the configuration and DB are properly initialized.
func TestBudgetProgram(t *testing.T) {
	TestCommons(t)
	t.Run("BudgetProgram", func(t *testing.T) {
		getAllBudgetProgramsTest(testCtx.E, t)
		getChapterBudgetProgramsTest(testCtx.E, t)
		bpID := createBudgetProgramTest(testCtx.E, t)
		modifyBudgetProgramTest(testCtx.E, t, bpID)
		deleteBudgetProgramTest(testCtx.E, t, bpID)
	})
}

// getAllBudgetProgramsTest tests route is protected and all programs are sent back.
func getAllBudgetProgramsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}, ArraySize: 0},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"BudgetProgram"}, ArraySize: 84},
	}

	for _, tc := range testCases {
		response := e.GET("/api/budget_programs").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("BudgetProgram").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

// getChapterBudgetProgramsTest tests route is protected and sent programs linked to a chapter are sent back.
func getChapterBudgetProgramsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}, ArraySize: 0},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"BudgetProgram"}, ArraySize: 11},
	}

	for _, tc := range testCases {
		response := e.GET("/api/budget_chapters/3/budget_programs").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("BudgetProgram").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

// createBudgetProgramTest tests route is protected and sent program is created.
func createBudgetProgramTest(e *httpexpect.Expect, t *testing.T) (bpID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "3", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "3", Status: http.StatusBadRequest, Sent: []byte(`{}`),
			BodyContains: []string{"Création de programme budgétaire, champ manquant ou incorrect"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusBadRequest,
			Sent:         []byte(`{"code_contract":"C","code_function":"FF","code_number":"NNN"}`),
			BodyContains: []string{"Création de programme budgétaire, index chapitre incorrect"}},
		{Token: testCtx.Admin.Token, ID: "3", Status: http.StatusOK,
			Sent: []byte(`{"code_contract":"C","code_function":"FF","code_number":"NNN"}`),
			BodyContains: []string{"BudgetProgram", `"code_contract":"C"`, `"code_function":"FF"`,
				`"code_number":"NNN"`, `"code_subfunction":null`}},
	}

	for _, tc := range testCases {
		response := e.POST("/api/budget_chapters/"+tc.ID+"/budget_programs").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		if tc.Status == http.StatusOK {
			bpID = int(response.JSON().Object().Value("BudgetProgram").Object().Value("id").Number().Raw())
		}
		response.Status(tc.Status)
	}
	return bpID
}

// modifyBudgetProgramTest tests route is protected and modify work properly.
func modifyBudgetProgramTest(e *httpexpect.Expect, t *testing.T, bpID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusBadRequest,
			BodyContains: []string{"Modification de programme : introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(bpID), Status: http.StatusOK,
			Sent: []byte(`{"code_contract":"X","code_function":"YY","code_number":"ZZZ","code_subfunction":"9"}`),
			BodyContains: []string{"BudgetProgram", `"code_contract":"X"`, `"code_function":"YY"`,
				`"code_number":"ZZZ"`, `"code_subfunction":"9"`}},
	}

	for _, tc := range testCases {
		response := e.PUT("/api/budget_chapters/3/budget_programs/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}

// deleteBudgetProgramTest tests route is protected and delete work properly.
func deleteBudgetProgramTest(e *httpexpect.Expect, t *testing.T, bpID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusNotFound,
			BodyContains: []string{"Suppression de programme : introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(bpID), Status: http.StatusOK,
			BodyContains: []string{"Programme supprimé"}},
	}

	for _, tc := range testCases {
		response := e.DELETE("/api/budget_chapters/3/budget_programs/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}
