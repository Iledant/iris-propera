package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestBudgetProgram embeddes all tests for budget programs insuring the configuration and DB are properly initialized.
func testBudgetProgram(t *testing.T) {
	t.Run("BudgetProgram", func(t *testing.T) {
		getAllBudgetProgramsTest(testCtx.E, t)
		getChapterBudgetProgramsTest(testCtx.E, t)
		bpID := createBudgetProgramTest(testCtx.E, t)
		if bpID == 0 {
			t.Fatal("Impossible de créer le programme")
		}
		modifyBudgetProgramTest(testCtx.E, t, bpID)
		deleteBudgetProgramTest(testCtx.E, t, bpID)
		batchBudgetProgramTest(testCtx.E, t)
	})
}

// getAllBudgetProgramsTest tests route is protected and all programs are sent back.
func getAllBudgetProgramsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusOK,
			BodyContains: []string{"BudgetProgram"},
			ArraySize:    84},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/budget_programs").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetAllBudgetPrograms") {
		t.Error(r)
	}
}

// getChapterBudgetProgramsTest tests route is protected and sent programs linked to a chapter are sent back.
func getChapterBudgetProgramsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusOK,
			BodyContains: []string{"BudgetProgram"},
			ArraySize:    11},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/budget_chapters/3/programs").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetChapterBudgetPrograms") {
		t.Error(r)
	}
}

// createBudgetProgramTest tests route is protected and sent program is created.
func createBudgetProgramTest(e *httpexpect.Expect, t *testing.T) (bpID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "3",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{`),
			BodyContains: []string{"Création d'un programme, décodage : "}},
		{
			Token:        testCtx.Admin.Token,
			ID:           "3",
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{}`),
			BodyContains: []string{"Création d'un programme : Champ manquant ou incorrect"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"code_contract":"C","code_function":"FF","code_number":"NNN","name":"Programme"}`),
			BodyContains: []string{`Création d'un programme, requête : pq: une instruction insert ou update sur la table « budget_program » viole la contrainte de clé`}},
		{
			Token:  testCtx.Admin.Token,
			ID:     "3",
			Status: http.StatusCreated,
			IDName: `"id"`,
			Sent:   []byte(`{"code_contract":"C","code_function":"FF","code_number":"NNN","name":"Programme"}`),
			BodyContains: []string{"BudgetProgram", `"code_contract":"C"`, `"code_function":"FF"`,
				`"code_number":"NNN"`, `"code_subfunction":null`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/budget_chapters/"+tc.ID+"/programs").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "CreateBudgetProgram", &bpID) {
		t.Error(r)
	}
	return bpID
}

// modifyBudgetProgramTest tests route is protected and modify work properly.
func modifyBudgetProgramTest(e *httpexpect.Expect, t *testing.T, bpID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"code_contract":"X","code_function":"YY","code_number":"ZZZ","code_subfunction":"9","name":"Programme"`),
			BodyContains: []string{"Modification d'un programme, décodage : "}},
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"code_contract":"X","code_function":"YY","code_number":"ZZZ","code_subfunction":"9","name":"Programme"}`),
			BodyContains: []string{"Modification d'un programme, requête : Programme introuvable"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(bpID),
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"code_contract":"XX","code_function":"YY","code_number":"ZZZ","code_subfunction":"9","name":"Programme"}`),
			BodyContains: []string{`Modification d'un programme : Champ manquant ou incorrect`}},
		{
			Token:  testCtx.Admin.Token,
			ID:     strconv.Itoa(bpID),
			Status: http.StatusOK,
			Sent:   []byte(`{"code_contract":"X","code_function":"YY","code_number":"ZZZ","code_subfunction":"9","name":"Programme"}`),
			BodyContains: []string{"BudgetProgram", `"code_contract":"X"`, `"code_function":"YY"`,
				`"code_number":"ZZZ"`, `"code_subfunction":"9"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.PUT("/api/budget_chapters/3/programs/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "ModifyBudgetProgram") {
		t.Error(r)
	}
}

// deleteBudgetProgramTest tests route is protected and delete work properly.
func deleteBudgetProgramTest(e *httpexpect.Expect, t *testing.T, bpID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Suppression d'un programme, requête : Programme introuvable"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(bpID),
			Status:       http.StatusOK,
			BodyContains: []string{"Programme supprimé"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.DELETE("/api/budget_chapters/3/programs/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DeleteBudgetProgram") {
		t.Error(r)
	}
}

// batchBudgetProgramTest tests route is protected and modify work properly.
func batchBudgetProgramTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{fake}`),
			BodyContains: []string{"Batch de programmes budgétaires, décodage : "}},
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusInternalServerError,
			Sent: []byte(`{"BudgetProgram":[{"code":"12345","subfunction":null,"name":"Batch 1","chapter":907},
			{"code":"12345678","subfunction":"999","name":"Batch 2","chapter":908}]}`),
			BodyContains: []string{"Batch de programmes budgétaires, requête : Code 12345 trop court"}},
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusOK,
			Sent: []byte(`{"BudgetProgram":[{"code":"1234567","subfunction":null,"name":"Batch 1","chapter":907},
			{"code":"12345678","subfunction":"999","name":"Batch 2","chapter":908}]}`),
			BodyContains: []string{"Batch importé"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/budget_programs").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "BatchBudgetProgram") {
		t.Error(r)
	}
}
