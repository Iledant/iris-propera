package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestBudgetChapter embeddes all tests for budget chapters insuring the configuration and DB are properly initialized.
func testBudgetChapter(t *testing.T) {
	t.Run("BudgetChapter", func(t *testing.T) {
		getAllBudgetChapters(testCtx.E, t)
		bcID := createBudgetChapterTest(testCtx.E, t)
		if bcID == 0 {
			t.Fatalf("Impossible de créer le chapitre budgétaire")
		}
		modifyBudgetChapterTest(testCtx.E, t, bcID)
		deleteBudgetChapterTest(testCtx.E, t, bcID)
	})
}

// getAllBudgetChapters tests route is protected and all chapters are sent back.
func getAllBudgetChapters(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusOK,
			BodyContains: []string{"BudgetChapter"},
			ArraySize:    3},
	}

	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/budget_chapters").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetAllBudgetChapters") {
		t.Error(r)
	}
}

// createBudgetChapterTest tests route is protected and sent chapter is created.
func createBudgetChapterTest(e *httpexpect.Expect, t *testing.T) (bcID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{`),
			BodyContains: []string{"Création de chapitre budgétaire, décodage : "}},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{}`),
			BodyContains: []string{"Création de chapitre budgétaire : Name manquant ou trop long ou code absent"}},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusCreated,
			Sent:         []byte(`{"name":"Essai chapitre","code":999}`),
			BodyContains: []string{"BudgetChapter", `"name":"Essai chapitre"`},
			IDName:       `"id"`},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/budget_chapters").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "CreateBudgetChapter", &bcID) {
		t.Error(r)
	}
	return bcID
}

// modifyBudgetChapterTest tests route is protected and modify work properly.
func modifyBudgetChapterTest(e *httpexpect.Expect, t *testing.T, bcID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Essai chapitre 2","code":888`),
			BodyContains: []string{`Modification d'un chapitre, décodage :`}},
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Essai chapitre 2","code":888}`),
			BodyContains: []string{`Modification d'un chapitre, requête : Chapitre budgétaire introuvable`}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(bcID),
			Status:       http.StatusOK,
			Sent:         []byte(`{"name":"Essai chapitre 2","code":888}`),
			BodyContains: []string{`BudgetChapter`, `"id":` + strconv.Itoa(bcID), `"name":"Essai chapitre 2"`, `"code":888`}},
	}

	f := func(tc testCase) *httpexpect.Response {
		return e.PUT("/api/budget_chapters/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "ModifyBudgetChapter") {
		t.Error(r)
	}
}

// deleteBudgetChapterTest tests route is protected and delete work properly.
func deleteBudgetChapterTest(e *httpexpect.Expect, t *testing.T, bcID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Suppression d'un chapitre, requête : Chapitre budgétaire introuvable"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(bcID),
			Status:       http.StatusOK,
			BodyContains: []string{"Chapitre supprimé"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.DELETE("/api/budget_chapters/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DeleteBudgetChapter") {
		t.Error(r)
	}
}
