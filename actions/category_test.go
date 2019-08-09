package actions

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestCategory embeddes all tests for category insuring the configuration and DB are properly initialized.
func testCategory(t *testing.T) {
	t.Run("Category", func(t *testing.T) {
		getCategoriesTest(testCtx.E, t)
		getStepsAndCategoriesTest(testCtx.E, t)
		caID := createCategoryTest(testCtx.E, t)
		modifyCategoryTest(testCtx.E, t, caID)
		deleteCategoryTest(testCtx.E, t, caID)
	})
}

// getCategoriesTest tests route is protected and all categories are sent back.
func getCategoriesTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{
			Token:        "fake",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"},
		}, // 0 : bad token
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusOK,
			BodyContains: []string{"Category"},
			ArraySize:    22,
		}, // 1 : ok
	}

	for i, tc := range testCases {
		response := e.GET("/api/categories").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetCategories[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"",
					i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetCategories[%d],statut :  attendu ->%v  reçu <-%v",
				i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetCategories[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d",
					i, tc.ArraySize, count)
			}
		}
	}
}

// getStepsAndCategoriesTest tests route is protected and all categories are sent back.
func getStepsAndCategoriesTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{
			Token:        "fake",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"},
		}, // 0 : bad token
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusOK,
			BodyContains: []string{"Category", "Step"},
			ArraySize:    26,
		}, // 1 : ok
	}

	for i, tc := range testCases {
		response := e.GET("/api/steps_categories").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetStepsAndCategories[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"",
					i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetStepsAndCategories[%d],statut :  attendu ->%v  reçu <-%v",
				i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetStepsAndCategories[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d",
					i, tc.ArraySize, count)
			}
		}
	}
}

// createCategoryTest tests route is protected and sent action is created.
func createCategoryTest(e *httpexpect.Expect, t *testing.T) (caID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest, Sent: []byte(`{}`),
			BodyContains: []string{"Création d'une catégorie : Name invalide"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent:         []byte(`{"name":"Test création catégorie"}`),
			BodyContains: []string{"Category", `"name":"Test création catégorie"`}},
	}

	for i, tc := range testCases {
		response := e.POST("/api/categories").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nCreateCategory[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nCreateCategory[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
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
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Test modification catégorie"}`),
			BodyContains: []string{"Modification d'une catégorie, requête : Catégorie introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(caID), Status: http.StatusOK,
			Sent:         []byte(`{"name":"Test modification catégorie"}`),
			BodyContains: []string{"Category", `"name":"Test modification catégorie"`}},
	}

	for i, tc := range testCases {
		response := e.PUT("/api/categories/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nModifyCategory[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nModifyCategory[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// deleteCategoryTest tests route is protected and delete work properly.
func deleteCategoryTest(e *httpexpect.Expect, t *testing.T, caID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Suppression d'une catégorie, requête : Catégorie introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(caID), Status: http.StatusOK,
			BodyContains: []string{"Catégorie supprimée"}},
	}

	for i, tc := range testCases {
		response := e.DELETE("/api/categories/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nDeleteCategory[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nDeleteCategory[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}
