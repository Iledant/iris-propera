package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/Iledant/iris_propera/models"
	"github.com/iris-contrib/httpexpect"
)

// TestBudgetAction embeddes all tests for budget actions insuring the configuration and DB are properly initialized.
func TestBudgetAction(t *testing.T) {
	TestCommons(t)
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
	testCases := []struct {
		Token        string
		Status       int
		BodyContains string
		ArraySize    int
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis", ArraySize: 0},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, BodyContains: "BudgetAction", ArraySize: 117},
	}

	for _, tc := range testCases {
		response := e.GET("/api/budget_actions").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("BudgetAction").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

// getProgramBudgetActions tests route is protected and sent actions linked are sent back.
func getProgramBudgetActions(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Status       int
		BodyContains string
		ArraySize    int
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis", ArraySize: 0},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, BodyContains: "BudgetAction", ArraySize: 1},
	}

	for _, tc := range testCases {
		response := e.GET("/api/budget_chapters/1/budget_programs/123/budget_actions").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("BudgetAction").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

// createBudgetActionTest tests route is protected and sent action is created.
func createBudgetActionTest(e *httpexpect.Expect, t *testing.T) int {
	testCases := []struct {
		Token        string
		Status       int
		BudgetAction models.BudgetAction
		BodyContains string
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis"},
		{Token: testCtx.Admin.Token, Status: http.StatusInternalServerError, BudgetAction: models.BudgetAction{}, BodyContains: "viole la contrainte"},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, BudgetAction: models.BudgetAction{Name: "Essai", SectorID: 3, Code: "999"}, BodyContains: "BudgetAction"},
	}
	var baID int

	for _, tc := range testCases {
		response := e.POST("/api/budget_chapters/1/budget_programs/123/budget_actions").WithHeader("Authorization", "Bearer "+tc.Token).WithJSON(tc.BudgetAction).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("BudgetAction").Object().Value("name").String().Equal(tc.BudgetAction.Name)
			response.JSON().Object().Value("BudgetAction").Object().Value("code").String().Equal(tc.BudgetAction.Code)
			response.JSON().Object().Value("BudgetAction").Object().Value("sector_id").Number().Equal(tc.BudgetAction.SectorID)
			baID = int(response.JSON().Object().Value("BudgetAction").Object().Value("id").Number().Raw())
		}
		response.Status(tc.Status)
	}
	return baID
}

// modifyBudgetActionTest tests route is protected and modify work properly.
func modifyBudgetActionTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Status       int
		BudgetAction models.BudgetAction
		BodyContains string
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis"},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, BudgetAction: models.BudgetAction{Name: "Essai tramways", Code: ""}, BodyContains: "BudgetAction"},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, BudgetAction: models.BudgetAction{Name: "Essai tramways2", Code: "999"}, BodyContains: "BudgetAction"},
	}

	for _, tc := range testCases {
		response := e.PUT("/api/budget_chapters/1/budget_programs/123/budget_actions/303").WithHeader("Authorization", "Bearer "+tc.Token).WithJSON(tc.BudgetAction).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.Status == http.StatusOK {
			if tc.BudgetAction.Name != "" {
				response.JSON().Object().Value("BudgetAction").Object().Value("name").String().Equal(tc.BudgetAction.Name)
			} else {
				response.JSON().Object().Value("BudgetAction").Object().Value("name").String().NotEqual(tc.BudgetAction.Name)
			}
			if tc.BudgetAction.Code != "" {
				response.JSON().Object().Value("BudgetAction").Object().Value("code").String().Equal(tc.BudgetAction.Code)
			} else {
				response.JSON().Object().Value("BudgetAction").Object().Value("code").String().NotEqual(tc.BudgetAction.Code)
			}
		}
		response.Status(tc.Status)
	}
}

// deleteBudgetActionTest tests route is protected and modify work properly.
func deleteBudgetActionTest(e *httpexpect.Expect, t *testing.T, baID int) {
	testCases := []struct {
		Token        string
		Status       int
		ID           int
		BodyContains string
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis"},
		{Token: testCtx.Admin.Token, ID: 0, Status: http.StatusNotFound, BodyContains: "Action introuvable"},
		{Token: testCtx.Admin.Token, ID: baID, Status: http.StatusOK, BodyContains: "Action supprimée"},
	}

	for _, tc := range testCases {
		response := e.DELETE("/api/budget_chapters/1/budget_programs/123/budget_actions/"+strconv.Itoa(tc.ID)).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		response.Body().Contains(tc.BodyContains)
		response.Status(tc.Status)
	}
}

type batchBa struct{ Code, Name, Sector string }

type batchBaa struct {
	BudgetAction []batchBa `json:"BudgetAction"`
}

// batchBudgetActionsTest tests route is protected and update and creations works.
func batchBudgetActionsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Status       int
		BudgetAction batchBaa
		BodyContains string
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis"},
		{Token: testCtx.Admin.Token, BudgetAction: batchBaa{BudgetAction: []batchBa{
			batchBa{Code: "000", Name: "batch BA name", Sector: "batch BA sector"}}},
			Status: http.StatusBadRequest, BodyContains: "code trop court"},
		{Token: testCtx.Admin.Token, BudgetAction: batchBaa{BudgetAction: []batchBa{
			batchBa{Code: "481005999", Name: "batch BA name", Sector: "TC"},
			batchBa{Code: "481005888", Name: "batch BA name2", Sector: "TMSP"},
		}},
			Status: http.StatusOK, BodyContains: "Actions mises à jour"},
	}

	for _, tc := range testCases {
		response := e.POST("/api/budget_actions").WithHeader("Authorization", "Bearer "+tc.Token).WithJSON(tc.BudgetAction).Expect()
		response.Body().Contains(tc.BodyContains)
		response.Status(tc.Status)
	}

	response := e.GET("/api/budget_actions").WithHeader("Authorization", "Bearer "+testCtx.Admin.Token).Expect()
	response.Body().Contains("batch BA name")
	response.Body().Contains("batch BA name2")
	response.JSON().Object().Value("BudgetAction").Array().Length().Equal(118)
}
