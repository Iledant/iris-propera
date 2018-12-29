package actions

import (
	"net/http"
	"strconv"
	"strings"
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
		batchBudgetProgramTest(testCtx.E, t)
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

	for i, tc := range testCases {
		response := e.GET("/api/budget_programs").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetAllBudgetPrograms[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetAllBudgetPrograms[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetAllBudgetPrograms[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
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

	for i, tc := range testCases {
		response := e.GET("/api/budget_chapters/3/programs").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetChapterBudgetPrograms[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetChapterBudgetPrograms[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetChapterBudgetPrograms[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// createBudgetProgramTest tests route is protected and sent program is created.
func createBudgetProgramTest(e *httpexpect.Expect, t *testing.T) (bpID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "3", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "3", Status: http.StatusBadRequest, Sent: []byte(`{}`),
			BodyContains: []string{"Création d'un programme : Champ manquant ou incorrect"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			Sent:         []byte(`{"code_contract":"C","code_function":"FF","code_number":"NNN","name":"Programme"}`),
			BodyContains: []string{`Création d'un programme, requête : pq: une instruction insert ou update sur la table « budget_program » viole la contrainte de clé`}},
		{Token: testCtx.Admin.Token, ID: "3", Status: http.StatusOK,
			Sent: []byte(`{"code_contract":"C","code_function":"FF","code_number":"NNN","name":"Programme"}`),
			BodyContains: []string{"BudgetProgram", `"code_contract":"C"`, `"code_function":"FF"`,
				`"code_number":"NNN"`, `"code_subfunction":null`}},
	}
	for i, tc := range testCases {
		response := e.POST("/api/budget_chapters/"+tc.ID+"/programs").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nCreateBudgetProgram[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nCreateBudgetProgram[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.Status == http.StatusOK {
			bpID = int(response.JSON().Object().Value("BudgetProgram").Object().Value("id").Number().Raw())
		}
	}
	return bpID
}

// modifyBudgetProgramTest tests route is protected and modify work properly.
func modifyBudgetProgramTest(e *httpexpect.Expect, t *testing.T, bpID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			Sent:         []byte(`{"code_contract":"X","code_function":"YY","code_number":"ZZZ","code_subfunction":"9","name":"Programme"}`),
			BodyContains: []string{"Modification d'un programme, requête : Programme introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(bpID), Status: http.StatusOK,
			Sent: []byte(`{"code_contract":"X","code_function":"YY","code_number":"ZZZ","code_subfunction":"9","name":"Programme"}`),
			BodyContains: []string{"BudgetProgram", `"code_contract":"X"`, `"code_function":"YY"`,
				`"code_number":"ZZZ"`, `"code_subfunction":"9"`}},
	}

	for i, tc := range testCases {
		response := e.PUT("/api/budget_chapters/3/programs/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nModifyBudgetProgram[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nModifyBudgetProgram[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// deleteBudgetProgramTest tests route is protected and delete work properly.
func deleteBudgetProgramTest(e *httpexpect.Expect, t *testing.T, bpID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Suppression d'un programme, requête : Programme introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(bpID), Status: http.StatusOK,
			BodyContains: []string{"Programme supprimé"}},
	}

	for i, tc := range testCases {
		response := e.DELETE("/api/budget_chapters/3/programs/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nDeleteBudgetProgram[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nDeleteBudgetProgram[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// batchBudgetProgramTest tests route is protected and modify work properly.
func batchBudgetProgramTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest,
			Sent:         []byte(`{fake}`),
			BodyContains: []string{"Batch de programmes budgétaires, décodage : "}},
		{Token: testCtx.Admin.Token, Status: http.StatusInternalServerError,
			Sent: []byte(`{"BudgetProgram":[{"code":"12345","subfunction":null,"name":"Batch 1","chapter":907},
			{"code":"12345678","subfunction":"999","name":"Batch 2","chapter":908}]}`),
			BodyContains: []string{"Batch de programmes budgétaires, requête : Code 12345 trop court"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent: []byte(`{"BudgetProgram":[{"code":"1234567","subfunction":null,"name":"Batch 1","chapter":907},
			{"code":"12345678","subfunction":"999","name":"Batch 2","chapter":908}]}`),
			BodyContains: []string{"Batch importé"}},
	}

	for i, tc := range testCases {
		response := e.POST("/api/budget_programs").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nBatchBudgetProgram[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nBatchBudgetProgram[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}
