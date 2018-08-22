package actions

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func TestPlan(t *testing.T) {
	TestCommons(t)
	t.Run("Plan", func(t *testing.T) {
		getPlansTest(testCtx.E, t)
		pID := createPlanTest(testCtx.E, t)
		modifyPlanTest(testCtx.E, t, pID)
		deletePlanTest(testCtx.E, t, pID)
	})
}

// getPlansTest check if route is protected and query sent correct datas.
func getPlansTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Status       int
		BodyContains []string
		Count        int
	}{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK, BodyContains: []string{"Plan", "name", "descript", "first_year", "last_year"}, Count: 5},
	}
	for i, tc := range testCases {
		response := e.GET("/api/plans").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("GetPlansTest[%d]:contenu incorrect, attendu \"%s\" et reçu \n%s", i, s, content)
			}
		}
		response.Status(tc.Status)
		response.ContentType("application/json")
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("Plan").Array().Length().Equal(tc.Count)
		}
	}
}

// createPlanTest check if route is protected and plan sent back is correct.
func createPlanTest(e *httpexpect.Expect, t *testing.T) (pID int) {
	testCases := []struct {
		Token        string
		Status       int
		Sent         []byte
		BodyContains []string
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusInternalServerError, Sent: []byte(`{Plu}`),
			BodyContains: []string{"Création de plan, impossible de décoder"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest, Sent: []byte(`{"Descript":null}`),
			BodyContains: []string{"Création d'un plan : mauvais format de name"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Sent: []byte(`{"name":"Essai de plan", "descript":"Essai de description","first_year":2015,"last_year":2025}`),
			BodyContains: []string{"Plan", `"name":"Essai de plan"`, `"first_year":2015`, `"last_year":2025`, `"descript":"Essai de description"`}},
	}
	for i, tc := range testCases {
		response := e.POST("/api/plans").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("CreatePlan[%d]:contenu incorrect, attendu \"%s\" et reçu \n%s", i, s, content)
			}
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			pID = int(response.JSON().Object().Value("Plan").Object().Value("id").Number().Raw())
		}
	}
	return pID
}

// modifyPlanTest check if route is protected and plan sent back is correct.
func modifyPlanTest(e *httpexpect.Expect, t *testing.T, pID int) {
	testCases := []struct {
		Token        string
		Status       int
		PlanID       string
		Sent         []byte
		BodyContains []string
	}{
		{Token: testCtx.User.Token, PlanID: "0", Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, PlanID: "0", Status: http.StatusBadRequest, Sent: []byte(`{Plu}`),
			BodyContains: []string{"Modification de plan: introuvable"}},
		{Token: testCtx.Admin.Token, PlanID: strconv.Itoa(pID), Status: http.StatusInternalServerError, Sent: []byte(`{Plu}`),
			BodyContains: []string{"Modification de plan, erreur décodage"}},
		{Token: testCtx.Admin.Token, PlanID: strconv.Itoa(pID), Status: http.StatusOK, Sent: []byte(`{"name":"Modification de plan", "descript":"Modification de description","first_year":2016,"last_year":2024}`),
			BodyContains: []string{"Plan", `"name":"Modification de plan"`, `"first_year":2016`, `"last_year":2024`, `"descript":"Modification de description"`}},
	}
	for i, tc := range testCases {
		response := e.PUT("/api/plans/"+tc.PlanID).WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("ModifyPlan[%d]:contenu incorrect, attendu \"%s\" et reçu \n%s", i, s, content)
			}
		}
		response.Status(tc.Status)
	}
}

// deletePlanTest check if route is protected and plan sent back is correct.
func deletePlanTest(e *httpexpect.Expect, t *testing.T, pID int) {
	testCases := []struct {
		Token        string
		Status       int
		PlanID       string
		BodyContains []string
	}{
		{Token: testCtx.User.Token, PlanID: "0", Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, PlanID: "0", Status: http.StatusBadRequest,
			BodyContains: []string{"Suppression d'un plan: introuvable"}},
		{Token: testCtx.Admin.Token, PlanID: strconv.Itoa(pID), Status: http.StatusOK,
			BodyContains: []string{"Plan supprimé"}},
	}
	for i, tc := range testCases {
		response := e.DELETE("/api/plans/"+tc.PlanID).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("DeletePlan[%d]:contenu incorrect, attendu \"%s\" et reçu \n%s", i, s, content)
			}
		}
		response.Status(tc.Status)

		if tc.Status == http.StatusOK {
			content := string(e.GET("/api/plans").WithHeader("Authorization", "Bearer "+tc.Token).Expect().Content)
			if strings.Contains(content, `"id":`+tc.PlanID) {
				t.Errorf("DeletePlan[%d]:identificateur %s trouvé après suppression :\n%s", i, tc.PlanID, content)
			}
		}
	}
}
