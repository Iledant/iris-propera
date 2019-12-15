package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestBudgetSector embeddes all tests for budget programs insuring the configuration and DB are properly initialized.
func testBudgetSector(t *testing.T) {
	t.Run("BudgetSector", func(t *testing.T) {
		getBudgetSectorsTest(testCtx.E, t)
		bsID := createBudgetSectorTest(testCtx.E, t)
		if bsID == 0 {
			t.Fatal("Impossible de créer le secteur budgétaire")
		}
		modifyBudgetSectorTest(testCtx.E, t, bsID)
		deleteBudgetSectorTest(testCtx.E, t, bsID)
	})
}

// getBudgetSectorsTest tests route is protected and all sectors are sent back.
func getBudgetSectorsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusOK,
			BodyContains: []string{"BudgetSector"},
			ArraySize:    4},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/budget_sectors").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetBudgetSectors") {
		t.Error(r)
	}
}

// createBudgetSectorTest tests route is protected and sent sector is created.
func createBudgetSectorTest(e *httpexpect.Expect, t *testing.T) (bsID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{}`),
			BodyContains: []string{"Création d'un secteur budgétaire : Code ou nom incorrect"}},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"code":"XX","name":"Test création secteur"`),
			BodyContains: []string{"Création d'un secteur budgétaire, décodage :"}},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusCreated,
			Sent:         []byte(`{"code":"XX","name":"Test création secteur"}`),
			IDName:       `"id"`,
			BodyContains: []string{"BudgetSector", `"code":"XX"`, `"name":"Test création secteur"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/budget_sectors").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "CreateBudgetSector", &bsID) {
		t.Error(r)
	}
	return bsID
}

// modifyBudgetSectorTest tests route is protected and modify work properly.
func modifyBudgetSectorTest(e *httpexpect.Expect, t *testing.T, bsID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"code":"YY","name":"Test modification secteur"}`),
			BodyContains: []string{"Modification d'un secteur budgétaire, requête : Secteur budgétaire introuvable"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(bsID),
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"code":"YY","name":"Test modification secteur"`),
			BodyContains: []string{"Modification d'un secteur budgétaire, décodage : "}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(bsID),
			Status:       http.StatusOK,
			Sent:         []byte(`{"code":"YY","name":"Test modification secteur"}`),
			BodyContains: []string{"BudgetSector", `"code":"YY"`, `"name":"Test modification secteur"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.PUT("/api/budget_sectors/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "ModifyBudgetSector") {
		t.Error(r)
	}
}

// deleteBudgetSectorTest tests route is protected and delete work properly.
func deleteBudgetSectorTest(e *httpexpect.Expect, t *testing.T, bsID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusNotFound,
			BodyContains: []string{"Suppression d'un secteur budgétaire, requête : Secteur budgétaire introuvable"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(bsID),
			Status:       http.StatusOK,
			BodyContains: []string{"Secteur supprimé"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.DELETE("/api/budget_sectors/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DeleteBudgetSector") {
		t.Error(r)
	}
}
