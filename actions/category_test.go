package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestCategory embeddes all tests for category insuring the configuration and DB are properly initialized.
func TestCategory(t *testing.T) {
	TestCommons(t)
	t.Run("Category", func(t *testing.T) {
		getCategoriesTest(testCtx.E, t)
		caID := createCategoryTest(testCtx.E, t)
		modifyCategoryTest(testCtx.E, t, caID)
		deleteCategoryTest(testCtx.E, t, caID)
	})
}

// getCategoriesTest tests route is protected and all categories are sent back.
func getCategoriesTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}, ArraySize: 0},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{"Category"}, ArraySize: 22},
	}

	for _, tc := range testCases {
		response := e.GET("/api/categories").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("Category").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

// createCategoryTest tests route is protected and sent action is created.
func createCategoryTest(e *httpexpect.Expect, t *testing.T) (caID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest, Sent: []byte(`{}`),
			BodyContains: []string{"Création de catégorie, champ 'name' manquant ou incorrect"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent:         []byte(`{"name":"Test création catégorie"}`),
			BodyContains: []string{"Category", `"name":"Test création catégorie"`}},
	}

	for _, tc := range testCases {
		response := e.POST("/api/categories").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		if tc.Status == http.StatusOK {
			caID = int(response.JSON().Object().Value("Category").Object().Value("id").Number().Raw())
		}
		response.Status(tc.Status)
	}
	return caID
}

// modifyCategoryTest tests route is protected and modify work properly.
func modifyCategoryTest(e *httpexpect.Expect, t *testing.T, caID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusBadRequest,
			BodyContains: []string{"Modification de catégorie : introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(caID), Status: http.StatusOK,
			Sent:         []byte(`{"name":"Test modification catégorie"}`),
			BodyContains: []string{"Category", `"name":"Test modification catégorie"`}},
	}

	for _, tc := range testCases {
		response := e.PUT("/api/categories/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}

// deleteCategoryTest tests route is protected and delete work properly.
func deleteCategoryTest(e *httpexpect.Expect, t *testing.T, caID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusNotFound,
			BodyContains: []string{"Suppression de catégorie : introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(caID), Status: http.StatusOK,
			BodyContains: []string{"Catégorie supprimée"}},
	}

	for _, tc := range testCases {
		response := e.DELETE("/api/categories/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}
