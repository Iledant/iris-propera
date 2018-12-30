package actions

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestBudgetSector embeddes all tests for budget programs insuring the configuration and DB are properly initialized.
func testBudgetSector(t *testing.T) {
	t.Run("BudgetSector", func(t *testing.T) {
		getBudgetSectorsTest(testCtx.E, t)
		bsID := createBudgetSectorTest(testCtx.E, t)
		modifyBudgetSectorTest(testCtx.E, t, bsID)
		deleteBudgetSectorTest(testCtx.E, t, bsID)
	})
}

// getBudgetSectorsTest tests route is protected and all sectors are sent back.
func getBudgetSectorsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}, ArraySize: 0},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"BudgetSector"}, ArraySize: 4},
	}

	for i, tc := range testCases {
		response := e.GET("/api/budget_sectors").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetBudgetSectors[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetBudgetSectors[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetBudgetSectors[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// createBudgetSectorTest tests route is protected and sent sector is created.
func createBudgetSectorTest(e *httpexpect.Expect, t *testing.T) (bsID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest, Sent: []byte(`{}`),
			BodyContains: []string{"Création d'un secteur budgétaire : Code ou nom incorrect"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent:         []byte(`{"code":"XX","name":"Test création secteur"}`),
			BodyContains: []string{"BudgetSector", `"code":"XX"`, `"name":"Test création secteur"`}},
	}
	for i, tc := range testCases {
		response := e.POST("/api/budget_sectors").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nCreateBudgetSector[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nCreateBudgetSector[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.Status == http.StatusOK {
			bsID = int(response.JSON().Object().Value("BudgetSector").Object().Value("id").Number().Raw())
		}
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
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			Sent:         []byte(`{"code":"YY","name":"Test modification secteur"}`),
			BodyContains: []string{"Modification d'un secteur budgétaire, requête : Secteur budgétaire introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(bsID), Status: http.StatusOK,
			Sent:         []byte(`{"code":"YY","name":"Test modification secteur"}`),
			BodyContains: []string{"BudgetSector", `"code":"YY"`, `"name":"Test modification secteur"`}},
	}

	for i, tc := range testCases {
		response := e.PUT("/api/budget_sectors/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nModifyBudgetSector[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		response.Status(tc.Status)
	}
}

// deleteBudgetSectorTest tests route is protected and delete work properly.
func deleteBudgetSectorTest(e *httpexpect.Expect, t *testing.T, bsID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusNotFound,
			BodyContains: []string{"Suppression d'un secteur budgétaire, requête : Secteur budgétaire introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(bsID), Status: http.StatusOK,
			BodyContains: []string{"Secteur supprimé"}},
	}

	for i, tc := range testCases {
		response := e.DELETE("/api/budget_sectors/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nDeleteBudgetSector[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		response.Status(tc.Status)
	}
}
