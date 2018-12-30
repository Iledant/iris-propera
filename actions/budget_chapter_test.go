package actions

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestBudgetChapter embeddes all tests for budget chapters insuring the configuration and DB are properly initialized.
func testBudgetChapter(t *testing.T) {
	t.Run("BudgetChapter", func(t *testing.T) {
		getAllBudgetChapters(testCtx.E, t)
		bcID := createBudgetChapterTest(testCtx.E, t)
		modifyBudgetChapterTest(testCtx.E, t, bcID)
		deleteBudgetChapterTest(testCtx.E, t, bcID)
	})
}

// getAllBudgetChapters tests route is protected and all chapters are sent back.
func getAllBudgetChapters(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}, ArraySize: 0},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{"BudgetChapter"}, ArraySize: 3},
	}

	for i, tc := range testCases {
		response := e.GET("/api/budget_chapters").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetAllBudgetChapters[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetAllBudgetChapters[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetAllBudgetChapters[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// createBudgetChapterTest tests route is protected and sent chapter is created.
func createBudgetChapterTest(e *httpexpect.Expect, t *testing.T) (bcID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest, Sent: []byte(`{}`),
			BodyContains: []string{"Création de chapitre budgétaire : Name manquant ou trop long ou code absent"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent:         []byte(`{"name":"Essai chapitre","code":999}`),
			BodyContains: []string{"BudgetChapter", `"name":"Essai chapitre"`}},
	}

	for i, tc := range testCases {
		response := e.POST("/api/budget_chapters").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nCreateBudgetChapter[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nCreateBudgetChapter[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.Status == http.StatusOK {
			bcID = int(response.JSON().Object().Value("BudgetChapter").Object().Value("id").Number().Raw())
		}
	}
	return bcID
}

// modifyBudgetChapterTest tests route is protected and modify work properly.
func modifyBudgetChapterTest(e *httpexpect.Expect, t *testing.T, bcID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, ID: "0",
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Essai chapitre 2","code":888}`),
			BodyContains: []string{`Modification d'un chapitre, requête : Chapitre budgétaire introuvable`}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(bcID), Status: http.StatusOK,
			Sent:         []byte(`{"name":"Essai chapitre 2","code":888}`),
			BodyContains: []string{`BudgetChapter`, `"id":` + strconv.Itoa(bcID), `"name":"Essai chapitre 2"`, `"code":888`}},
	}

	for i, tc := range testCases {
		response := e.PUT("/api/budget_chapters/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nModifyBudgetChapter[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nModifyBudgetChapter[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// deleteBudgetChapterTest tests route is protected and delete work properly.
func deleteBudgetChapterTest(e *httpexpect.Expect, t *testing.T, bcID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, ID: "0",
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Suppression d'un chapitre, requête : Chapitre budgétaire introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(bcID), Status: http.StatusOK,
			BodyContains: []string{"Chapitre supprimé"}},
	}

	for i, tc := range testCases {
		response := e.DELETE("/api/budget_chapters/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nDeleteBudgetChapter[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nDeleteBudgetChapter[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}
