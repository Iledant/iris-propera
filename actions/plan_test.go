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
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK, ArraySize: 5,
			BodyContains: []string{"Plan", "name", "descript", "first_year", "last_year"}},
	}
	for i, tc := range testCases {
		response := e.GET("/api/plans").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetPlans[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetPlans[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetPlans[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// createPlanTest check if route is protected and plan sent back is correct.
func createPlanTest(e *httpexpect.Expect, t *testing.T) (pID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusInternalServerError, Sent: []byte(`{Plu}`),
			BodyContains: []string{"Création de plan, décodage :"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest, Sent: []byte(`{"Descript":null}`),
			BodyContains: []string{"Création d'un plan : Name incorrect"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent: []byte(`{"name":"Essai de plan", "descript":"Essai de description","first_year":2015,"last_year":2025}`),
			BodyContains: []string{"Plan", `"name":"Essai de plan"`, `"first_year":2015`,
				`"last_year":2025`, `"descript":"Essai de description"`}},
	}
	for i, tc := range testCases {
		response := e.POST("/api/plans").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nCreatePlan[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nCreatePlan[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.Status == http.StatusOK {
			pID = int(response.JSON().Object().Value("Plan").Object().Value("id").Number().Raw())
		}
	}
	return pID
}

// modifyPlanTest check if route is protected and plan sent back is correct.
func modifyPlanTest(e *httpexpect.Expect, t *testing.T, pID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Modification de plan", "descript":"Modification de description","first_year":2016,"last_year":2024}`),
			BodyContains: []string{"Modification de plan, requête : Plan introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(pID), Status: http.StatusInternalServerError,
			Sent:         []byte(`{"Plu"}`),
			BodyContains: []string{"Modification de plan, décodage :"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(pID), Status: http.StatusOK,
			Sent: []byte(`{"name":"Modification de plan", "descript":"Modification de description","first_year":2016,"last_year":2024}`),
			BodyContains: []string{"Plan", `"name":"Modification de plan"`, `"first_year":2016`,
				`"last_year":2024`, `"descript":"Modification de description"`}},
	}
	for i, tc := range testCases {
		response := e.PUT("/api/plans/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nModifyPlan[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nModifyPlan[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// deletePlanTest check if route is protected and plan sent back is correct.
func deletePlanTest(e *httpexpect.Expect, t *testing.T, pID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Suppression de plan, requête : Plan introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(pID), Status: http.StatusOK,
			BodyContains: []string{"Plan supprimé"}},
	}
	for i, tc := range testCases {
		response := e.DELETE("/api/plans/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nDeletePlan[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nDeletePlan[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.Status == http.StatusOK {
			content := string(e.GET("/api/plans").WithHeader("Authorization", "Bearer "+tc.Token).Expect().Content)
			if strings.Contains(content, `"id":`+tc.ID) {
				t.Errorf("DeletePlan[%d]:identificateur %s trouvé après suppression :\n%s", i, tc.ID, content)
			}
		}
	}
}
