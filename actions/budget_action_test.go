package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestBudgetAction embeddes all tests for budget actions insuring the configuration and DB are properly initialized.
func testBudgetAction(t *testing.T) {
	t.Run("BudgetAction", func(t *testing.T) {
		getAllBudgetActions(testCtx.E, t)
		getProgramBudgetActions(testCtx.E, t)
		baID := createBudgetActionTest(testCtx.E, t)
		if baID == 0 {
			t.Fatalf("Impossible de créer l'action budgétaire")
		}
		modifyBudgetActionTest(testCtx.E, t)
		deleteBudgetActionTest(testCtx.E, t, baID)
		batchBudgetActionsTest(testCtx.E, t)
	})
}

// getAllBudgetActions tests route is protected and all actions are sent back.
func getAllBudgetActions(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusOK,
			BodyContains: []string{"BudgetAction"},
			ArraySize:    117},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusOK,
			Param:        "FullCodeAction=true",
			BodyContains: []string{"BudgetAction", "full_code"},
			ArraySize:    117},
	}

	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/budget_actions").WithQueryString(tc.Param).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chtTestCases(testCases, f, "AllBudgetActions") {
		t.Error(r)
	}
}

// getProgramBudgetActions tests route is protected and sent actions linked are sent back.
func getProgramBudgetActions(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"},
			ArraySize:    0},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusOK,
			BodyContains: []string{"BudgetAction"},
			ArraySize:    1},
	}

	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/budget_chapters/1/programs/123/actions").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chtTestCases(testCases, f, "ProgramBudgetAction") {
		t.Error(r)
	}
}

// createBudgetActionTest tests route is protected and sent action is created.
func createBudgetActionTest(e *httpexpect.Expect, t *testing.T) int {
	testCases := []testCase{
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{
			Token: testCtx.Admin.Token, Status: http.StatusBadRequest,
			Sent:         []byte(`{}`),
			BodyContains: []string{"Création d'action budgétaire : Code, nom ou ID secteur incorrect"}},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusCreated,
			Sent:         []byte(`{"name":"Essai","sector_id":3,"code":"999"}`),
			IDName:       `"id"`,
			BodyContains: []string{"BudgetAction", `"name":"Essai"`, `"sector_id":3`, `"code":"999"`}},
	}
	var baID int
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/budget_chapters/1/programs/123/actions").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chtTestCases(testCases, f, "CreateBudgetAction", &baID) {
		t.Error(r)
	}
	return baID
}

// modifyBudgetActionTest tests route is protected and modify work properly.
func modifyBudgetActionTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{
			Token:        testCtx.User.Token,
			ID:           "0",
			Status:       http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Essai tramways","code":"999","sector_id":3}`),
			BodyContains: []string{"Modification d'action budgétaire, update : Action introuvable"}},

		{Token: testCtx.Admin.Token,
			ID:           "303",
			Status:       http.StatusOK,
			Sent:         []byte(`{"name":"Essai tramways","code":"999","sector_id":3}`),
			BodyContains: []string{"BudgetAction", `"name":"Essai tramways"`, `"code":"999"`}},
	}

	f := func(tc testCase) *httpexpect.Response {
		return e.PUT("/api/budget_chapters/1/programs/123/actions/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chtTestCases(testCases, f, "ModifyBudgetAction") {
		t.Error(r)
	}
}

// deleteBudgetActionTest tests route is protected and delete work properly.
func deleteBudgetActionTest(e *httpexpect.Expect, t *testing.T, baID int) {
	testCases := []testCase{
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusUnauthorized,
			ID:           "0",
			BodyContains: []string{"Droits administrateur requis"}},

		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Suppression d'action budgétaire, delete : Action budgétaire introuvable"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(baID),
			Status:       http.StatusOK,
			BodyContains: []string{"Action supprimée"}},
	}

	f := func(tc testCase) *httpexpect.Response {
		return e.DELETE("/api/budget_chapters/1/programs/123/actions/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chtTestCases(testCases, f, "DeleteBudgetAction") {
		t.Error(r)
	}
}

// batchBudgetActionsTest tests route is protected and update and creations works.
func batchBudgetActionsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{
			Token:        testCtx.Admin.Token,
			Sent:         []byte(`{"BudgetAction":[{"Code":"000","Name":"batch BA name","Sector":"batch BA sector"}]}`),
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"code trop court"}},
		{
			Token: testCtx.Admin.Token,
			Sent: []byte(`{"BudgetAction":[{"Code":"481005999","Name":"batch BA name","Sector":"TC"},
			{"Code":"481005888","Name":"batch BA name2","Sector":"TMSP"}]}`),
			Status:       http.StatusOK,
			BodyContains: []string{"Actions mises à jour"}},
	}

	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/budget_actions").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chtTestCases(testCases, f, "BatchBudgetAction") {
		t.Error(r)
	}

	testCases = []testCase{
		{
			Token:         testCtx.Admin.Token,
			Status:        http.StatusOK,
			BodyContains:  []string{"batch BA name", "batch BA name2"},
			CountItemName: `"id"`,
			ArraySize:     118},
	}
	f = func(tc testCase) *httpexpect.Response {
		return e.GET("/api/budget_actions").
			WithHeader("Authorization", "Bearer "+testCtx.Admin.Token).Expect()
	}
	for _, r := range chtTestCases(testCases, f, "BatchBudgetAction") {
		t.Error(r)
	}
}
