package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestBudgetChapter embeddes all tests for budget chapters insuring the configuration and DB are properly initialized.
func TestBudgetChapter(t *testing.T) {
	TestCommons(t)
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

	for _, tc := range testCases {
		response := e.GET("/api/budget_chapters").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("BudgetChapter").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

// createBudgetChapterTest tests route is protected and sent chapter is created.
func createBudgetChapterTest(e *httpexpect.Expect, t *testing.T) (bcID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest, Sent: []byte(`{}`),
			BodyContains: []string{"mauvais format des paramètres"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent:         []byte(`{"name":"Essai chapitre","code":999}`),
			BodyContains: []string{"BudgetChapter", `"name":"Essai chapitre"`}},
	}

	for _, tc := range testCases {
		response := e.POST("/api/budget_chapters").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		if tc.Status == http.StatusOK {
			bcID = int(response.JSON().Object().Value("BudgetChapter").Object().Value("id").Number().Raw())
		}
		response.Status(tc.Status)
	}
	return bcID
}

// modifyBudgetChapterTest tests route is protected and modify work properly.
func modifyBudgetChapterTest(e *httpexpect.Expect, t *testing.T, bcID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, ID: "0",
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusBadRequest,
			Sent:         []byte(`{"name":"Essai chapitre 2","code":888}`),
			BodyContains: []string{`Modification d'un chapitre, introuvable`}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(bcID), Status: http.StatusOK,
			Sent:         []byte(`{"name":"Essai chapitre 2","code":888}`),
			BodyContains: []string{`BudgetChapter`, `"id":` + strconv.Itoa(bcID), `"name":"Essai chapitre 2"`, `"code":888`}},
	}

	for _, tc := range testCases {
		response := e.PUT("/api/budget_chapters/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}

// deleteBudgetChapterTest tests route is protected and delete work properly.
func deleteBudgetChapterTest(e *httpexpect.Expect, t *testing.T, bcID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, ID: "0",
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusBadRequest,
			BodyContains: []string{"Suppression d'un chapitre, introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(bcID), Status: http.StatusOK,
			BodyContains: []string{"Chapitre supprimé"}},
	}

	for _, tc := range testCases {
		response := e.DELETE("/api/budget_chapters/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}
