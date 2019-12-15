package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestCategory embeddes all tests for category insuring the configuration and DB are properly initialized.
func testCategory(t *testing.T) {
	t.Run("Category", func(t *testing.T) {
		getCategoriesTest(testCtx.E, t)
		getStepsAndCategoriesTest(testCtx.E, t)
		caID := createCategoryTest(testCtx.E, t)
		if caID == 0 {
			t.Fatal("Impossible de créer la catégorie")
		}
		modifyCategoryTest(testCtx.E, t, caID)
		deleteCategoryTest(testCtx.E, t, caID)
	})
}

// getCategoriesTest tests route is protected and all categories are sent back.
func getCategoriesTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase, // 0 : bad token
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusOK,
			BodyContains: []string{"Category"},
			ArraySize:    22,
		}, // 1 : ok
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/categories").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetCategories") {
		t.Error(r)
	}
}

// getStepsAndCategoriesTest tests route is protected and all categories are sent back.
func getStepsAndCategoriesTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase, // 0 : bad token
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusOK,
			BodyContains: []string{"Category", "Step"},
			ArraySize:    26,
			IDName:       `"id"`,
		}, // 1 : ok
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/steps_categories").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetStepsAndCategories") {
		t.Error(r)
	}
}

// createCategoryTest tests route is protected and sent action is created.
func createCategoryTest(e *httpexpect.Expect, t *testing.T) (caID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{}`),
			BodyContains: []string{"Création d'une catégorie : Name invalide"}},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Test création catégorie"`),
			BodyContains: []string{"Création d'une catégorie, décodage : "}},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusCreated,
			Sent:         []byte(`{"name":"Test création catégorie"}`),
			IDName:       `"id"`,
			BodyContains: []string{"Category", `"name":"Test création catégorie"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/categories").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "CreateCategory", &caID) {
		t.Error(r)
	}
	return caID
}

// modifyCategoryTest tests route is protected and modify work properly.
func modifyCategoryTest(e *httpexpect.Expect, t *testing.T, caID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Test modification catégorie"}`),
			BodyContains: []string{"Modification d'une catégorie, requête : Catégorie introuvable"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(caID),
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Test modification catégorie"`),
			BodyContains: []string{"Modification d'une catégorie, décodage :"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(caID),
			Status:       http.StatusOK,
			Sent:         []byte(`{"name":"Test modification catégorie"}`),
			BodyContains: []string{"Category", `"name":"Test modification catégorie"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.PUT("/api/categories/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "ModifyCategory") {
		t.Error(r)
	}
}

// deleteCategoryTest tests route is protected and delete work properly.
func deleteCategoryTest(e *httpexpect.Expect, t *testing.T, caID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Suppression d'une catégorie, requête : Catégorie introuvable"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(caID),
			Status:       http.StatusOK,
			BodyContains: []string{"Catégorie supprimée"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.DELETE("/api/categories/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DeleteCategory") {
		t.Error(r)
	}
}
