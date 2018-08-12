package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/Iledant/iris_propera/models"
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
	testCases := []struct {
		Token        string
		Status       int
		BodyContains string
		ArraySize    int
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis", ArraySize: 0},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, BodyContains: "BudgetChapter", ArraySize: 3},
	}

	for _, tc := range testCases {
		response := e.GET("/api/budget_chapters").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("BudgetChapter").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

// createBudgetChapterTest tests route is protected and sent action is created.
func createBudgetChapterTest(e *httpexpect.Expect, t *testing.T) int {
	testCases := []struct {
		Token         string
		Status        int
		BudgetChapter models.BudgetChapter
		BodyContains  string
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis"},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest, BudgetChapter: models.BudgetChapter{}, BodyContains: "mauvais format des paramètres"},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, BudgetChapter: models.BudgetChapter{Name: "Essai chapitre", Code: 999}, BodyContains: "BudgetChapter"},
	}
	var bcID int

	for _, tc := range testCases {
		response := e.POST("/api/budget_chapters").WithHeader("Authorization", "Bearer "+tc.Token).WithJSON(tc.BudgetChapter).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("BudgetChapter").Object().Value("name").String().Equal(tc.BudgetChapter.Name)
			response.JSON().Object().Value("BudgetChapter").Object().Value("code").Number().Equal(tc.BudgetChapter.Code)
			bcID = int(response.JSON().Object().Value("BudgetChapter").Object().Value("id").Number().Raw())
		}
		response.Status(tc.Status)
	}
	return bcID
}

// modifyBudgetChapterTest tests route is protected and modify work properly.
func modifyBudgetChapterTest(e *httpexpect.Expect, t *testing.T, bcID int) {
	testCases := []struct {
		Token         string
		ID            int
		Status        int
		BudgetChapter models.BudgetChapter
		BodyContains  string
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis"},
		{Token: testCtx.Admin.Token, ID: 0, Status: http.StatusBadRequest, BudgetChapter: models.BudgetChapter{Name: "Essai chapitre 2", Code: 888}, BodyContains: "Modification d'un chapitre: introuvable"},
		{Token: testCtx.Admin.Token, ID: bcID, Status: http.StatusOK, BudgetChapter: models.BudgetChapter{Name: "Essai chapitre 2", Code: 0}, BodyContains: "BudgetChapter"},
		{Token: testCtx.Admin.Token, ID: bcID, Status: http.StatusOK, BudgetChapter: models.BudgetChapter{Name: "", Code: 888}, BodyContains: "BudgetChapter"},
		{Token: testCtx.Admin.Token, ID: bcID, Status: http.StatusOK, BudgetChapter: models.BudgetChapter{Name: "Essai chapitre 3", Code: 777}, BodyContains: "BudgetChapter"},
	}

	for _, tc := range testCases {
		response := e.PUT("/api/budget_chapters/"+strconv.Itoa(tc.ID)).WithHeader("Authorization", "Bearer "+tc.Token).WithJSON(tc.BudgetChapter).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.Status == http.StatusOK {
			if tc.BudgetChapter.Name != "" {
				response.JSON().Object().Value("BudgetChapter").Object().Value("name").String().Equal(tc.BudgetChapter.Name)
			}
			if tc.BudgetChapter.Code != 0 {
				response.JSON().Object().Value("BudgetChapter").Object().Value("code").Number().Equal(tc.BudgetChapter.Code)
			}
		}
		response.Status(tc.Status)
	}
}

// deleteBudgetChapterTest tests route is protected and modify work properly.
func deleteBudgetChapterTest(e *httpexpect.Expect, t *testing.T, bcID int) {
	testCases := []struct {
		Token        string
		Status       int
		ID           int
		BodyContains string
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis"},
		{Token: testCtx.Admin.Token, ID: 0, Status: http.StatusBadRequest, BodyContains: "Suppression d'un chapitre: introuvable"},
		{Token: testCtx.Admin.Token, ID: bcID, Status: http.StatusOK, BodyContains: "Chapitre supprimé"},
	}

	for _, tc := range testCases {
		response := e.DELETE("/api/budget_chapters/"+strconv.Itoa(tc.ID)).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		response.Body().Contains(tc.BodyContains)
		response.Status(tc.Status)
	}
}
