package actions

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestBudgetAction embeddes all tests for budget actions insuring the configuration and DB are properly initialized.
func testBudgetAction(t *testing.T) {
	t.Run("BudgetAction", func(t *testing.T) {
		getAllBudgetActions(testCtx.E, t)
		getProgramBudgetActions(testCtx.E, t)
		baID := createBudgetActionTest(testCtx.E, t)
		modifyBudgetActionTest(testCtx.E, t)
		deleteBudgetActionTest(testCtx.E, t, baID)
		batchBudgetActionsTest(testCtx.E, t)
	})
}

// getAllBudgetActions tests route is protected and all actions are sent back.
func getAllBudgetActions(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}, ArraySize: 0},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{"BudgetAction"}, ArraySize: 117},
	}

	for i, tc := range testCases {
		response := e.GET("/api/budget_actions").WithHeader("Authorization", "Bearer "+tc.Token).
			Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nAllBudgetActions[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nAllBudgetActions[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nAllBudgetActions[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// getProgramBudgetActions tests route is protected and sent actions linked are sent back.
func getProgramBudgetActions(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}, ArraySize: 0},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{"BudgetAction"}, ArraySize: 1},
	}

	for i, tc := range testCases {
		response := e.GET("/api/budget_chapters/1/programs/123/actions").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nProgramBudgetActions[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nProgramBudgetAction[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nProgramBudgetAction[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// createBudgetActionTest tests route is protected and sent action is created.
func createBudgetActionTest(e *httpexpect.Expect, t *testing.T) int {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest, Sent: []byte(`{}`),
			BodyContains: []string{"Création d'action budgétaire : Code, nom ou ID secteur incorrect"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent:         []byte(`{"name":"Essai","sector_id":3,"code":"999"}`),
			BodyContains: []string{"BudgetAction", `"name":"Essai"`, `"sector_id":3`, `"code":"999"`}},
	}
	var baID int

	for i, tc := range testCases {
		response := e.POST("/api/budget_chapters/1/programs/123/actions").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nCreateBudgetAction[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		if tc.Status == http.StatusOK {
			baID = int(response.JSON().Object().Value("BudgetAction").Object().Value("id").Number().Raw())
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nCreateBudgetAction[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
	return baID
}

// modifyBudgetActionTest tests route is protected and modify work properly.
func modifyBudgetActionTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Essai tramways","code":"999","sector_id":3}`),
			BodyContains: []string{"Modification d'action budgétaire, update : Action introuvable"}},
		{Token: testCtx.Admin.Token, ID: "303", Status: http.StatusOK,
			Sent:         []byte(`{"name":"Essai tramways","code":"999","sector_id":3}`),
			BodyContains: []string{"BudgetAction", `"name":"Essai tramways"`, `"code":"999"`}},
	}

	for i, tc := range testCases {
		response := e.PUT("/api/budget_chapters/1/programs/123/actions/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nModifyBudgetAction[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nModifyBudgetAction[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// deleteBudgetActionTest tests route is protected and delete work properly.
func deleteBudgetActionTest(e *httpexpect.Expect, t *testing.T, baID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, ID: "0",
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Suppression d'action budgétaire, delete : Action budgétaire introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(baID), Status: http.StatusOK,
			BodyContains: []string{"Action supprimée"}},
	}

	for i, tc := range testCases {
		response := e.DELETE("/api/budget_chapters/1/programs/123/actions/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nDeleteBudgetAction[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nDeleteBudgetAction[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// batchBudgetActionsTest tests route is protected and update and creations works.
func batchBudgetActionsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token,
			Sent:   []byte(`{"BudgetAction":[{"Code":"000","Name":"batch BA name","Sector":"batch BA sector"}]}`),
			Status: http.StatusInternalServerError, BodyContains: []string{"code trop court"}},
		{Token: testCtx.Admin.Token,
			Sent: []byte(`{"BudgetAction":[{"Code":"481005999","Name":"batch BA name","Sector":"TC"},
			{"Code":"481005888","Name":"batch BA name2","Sector":"TMSP"}]}`),
			Status: http.StatusOK, BodyContains: []string{"Actions mises à jour"}},
	}

	for i, tc := range testCases {
		response := e.POST("/api/budget_actions").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nBatchBudgetAction[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nBatchBudgetAction[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}

	response := e.GET("/api/budget_actions").WithHeader("Authorization", "Bearer "+testCtx.Admin.Token).Expect()
	content := string(response.Content)
	for _, s := range []string{"batch BA name", "batch BA name2"} {
		if !strings.Contains(content, s) {
			t.Errorf("\nBatchBudgetAction[get] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", s, content)
		}
	}
	count := strings.Count(content, `"id"`)
	if count != 118 {
		t.Errorf("\nBatchBudgetAction[get] :\n  nombre attendu -> %d\n  nombre reçu <-%d", 118, count)
	}
}
