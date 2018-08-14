package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestBudgetSector embeddes all tests for budget programs insuring the configuration and DB are properly initialized.
func TestBudgetSector(t *testing.T) {
	TestCommons(t)
	t.Run("BudgetSector", func(t *testing.T) {
		getBudgetSectorsTest(testCtx.E, t)
		bsID := createBudgetSectorTest(testCtx.E, t)
		modifyBudgetSectorTest(testCtx.E, t, bsID)
		deleteBudgetSectorTest(testCtx.E, t, bsID)
	})
}

// getBudgetSectorsTest tests route is protected and all sectors are sent back.
func getBudgetSectorsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Status       int
		BodyContains string
		ArraySize    int
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis", ArraySize: 0},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, BodyContains: "BudgetSector", ArraySize: 4},
	}

	for _, tc := range testCases {
		response := e.GET("/api/budget_sectors").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("BudgetSector").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

// createBudgetSectorTest tests route is protected and sent sector is created.
func createBudgetSectorTest(e *httpexpect.Expect, t *testing.T) int {
	testCases := []struct {
		Token        string
		Status       int
		Sent         []byte
		BodyContains []string
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest, Sent: []byte(`{}`), BodyContains: []string{"Création de secteur budgétaire, champ manquant ou incorrect"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Sent: []byte(`{"code":"XX","name":"Test création secteur"}`), BodyContains: []string{"BudgetSector", `"code":"XX"`, `"name":"Test création secteur"`}},
	}
	var bsID int

	for _, tc := range testCases {
		response := e.POST("/api/budget_sectors").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		if tc.Status == http.StatusOK {
			bsID = int(response.JSON().Object().Value("BudgetSector").Object().Value("id").Number().Raw())
		}
		response.Status(tc.Status)
	}
	return bsID
}

// modifyBudgetSectorTest tests route is protected and modify work properly.
func modifyBudgetSectorTest(e *httpexpect.Expect, t *testing.T, bsID int) {
	testCases := []struct {
		Token        string
		Status       int
		ID           string
		Sent         []byte
		BodyContains []string
	}{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusBadRequest, BodyContains: []string{"Modification de secteur : introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(bsID), Status: http.StatusOK, Sent: []byte(`{"code":"YY","name":"Test modification secteur"}`), BodyContains: []string{"BudgetSector", `"code":"YY"`, `"name":"Test modification secteur"`}},
	}

	for _, tc := range testCases {
		response := e.PUT("/api/budget_sectors/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}

// deleteBudgetSectorTest tests route is protected and delete work properly.
func deleteBudgetSectorTest(e *httpexpect.Expect, t *testing.T, bsID int) {
	testCases := []struct {
		Token        string
		Status       int
		ID           string
		BodyContains string
	}{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis"},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusNotFound, BodyContains: "Suppression de secteur : introuvable"},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(bsID), Status: http.StatusOK, BodyContains: "Secteur supprimé"},
	}

	for _, tc := range testCases {
		response := e.DELETE("/api/budget_sectors/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		response.Body().Contains(tc.BodyContains)
		response.Status(tc.Status)
	}
}
