package actions

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func TestPlanLine(t *testing.T) {
	TestCommons(t)
	t.Run("PlanLine", func(t *testing.T) {
		getPlanLinesTest(testCtx.E, t)
		getDetailedPlanLinesTest(testCtx.E, t)
		plID := createPlanLineTest(testCtx.E, t)
		modifyPlanLineTest(testCtx.E, t, plID)
		deletePlanLineTest(testCtx.E, t, plID)
	})
}

// getPlanLinesTest check if route is protected and query sent correct datas.
func getPlanLinesTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, ID: "1", Status: http.StatusOK,
			BodyContains: []string{"PlanLine", "Beneficiary"}, ArraySize: 59},
	}
	for i, tc := range testCases {
		response := e.GET("/api/plans/"+tc.ID+"/planlines").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetPlanLinesTest[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetPlanLinesTest[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"total_value"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetPlanLinesTest[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// getDetailedPlanLinesTest check if route is protected and query sent correct datas.
func getDetailedPlanLinesTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Liste détaillée des lignes de plan, requête plan : "}},
		{Token: testCtx.User.Token, ID: "1", Status: http.StatusOK,
			BodyContains: []string{"DetailedPlanLine"}, ArraySize: 443},
	}
	for i, tc := range testCases {
		response := e.GET("/api/plans/"+tc.ID+"/planlines/detailed").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetDetailedPlanLines[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetDetailedPlanLines[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetDetailedPlanLines[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// createPlanLineTest check if route is protected and plan line sent back is correct.
func createPlanLineTest(e *httpexpect.Expect, t *testing.T) (plID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError, Sent: []byte(`{Plu}`),
			BodyContains: []string{"Création de ligne de plan, requête plan : "}},
		{Token: testCtx.Admin.Token, ID: "1", Status: http.StatusInternalServerError, Sent: []byte(`{Plu}`),
			BodyContains: []string{"Création de ligne de plan, décodage"}},
		{Token: testCtx.Admin.Token, ID: "1", Status: http.StatusBadRequest,
			Sent:         []byte(`{"Descript":null}`),
			BodyContains: []string{"Création de ligne de plan, erreur de name"}},
		{Token: testCtx.Admin.Token, ID: "1", Status: http.StatusBadRequest,
			Sent:         []byte(`{"name":"Essai de ligne de plan"}`),
			BodyContains: []string{"Création de ligne de plan, erreur de value"}},
		{Token: testCtx.Admin.Token, ID: "1", Status: http.StatusOK,
			Sent: []byte(`{"name":"Essai de ligne de plan", "value":123,"total_value":456,
			"descript":"Essai de description","ratios":[{"ratio":0.5,"beneficiary_id":16}]}`),
			BodyContains: []string{"PlanLine", `"name":"Essai de ligne de plan"`, `"value":123`,
				`"total_value":456`, `"descript":"Essai de description"`}},
	}
	for i, tc := range testCases {
		response := e.POST("/api/plans/"+tc.ID+"/planlines").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nCreatePlanLine[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nCreatePlanLine[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.Status == http.StatusOK {
			plID = int(response.JSON().Object().Value("PlanLine").Object().Value("id").Number().Raw())
		}
	}
	return plID
}

// modifyPlanLineTest check if route is protected and plan line sent back is correct.
func modifyPlanLineTest(e *httpexpect.Expect, t *testing.T, plID int) {
	testCases := []struct {
		Token        string
		Status       int
		ID           string
		PlanLineID   string
		Sent         []byte
		BodyContains []string
	}{
		{Token: testCtx.User.Token, ID: "0", PlanLineID: "0", Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", PlanLineID: "0", Status: http.StatusInternalServerError, Sent: []byte(`{Plu}`),
			BodyContains: []string{"Modification de ligne de plan, requête plan : "}},
		{Token: testCtx.Admin.Token, ID: "1", PlanLineID: "0", Status: http.StatusInternalServerError, Sent: []byte(`{Plu}`),
			BodyContains: []string{"Modification de ligne de plan, requête getByID : "}},
		{Token: testCtx.Admin.Token, ID: "1", PlanLineID: strconv.Itoa(plID), Status: http.StatusInternalServerError, Sent: []byte(`{Plu}`),
			BodyContains: []string{"Modification de ligne de plan, décodage : "}},
		{Token: testCtx.Admin.Token, ID: "1", PlanLineID: strconv.Itoa(plID), Status: http.StatusOK, Sent: []byte(`{"name":"Modification de ligne de plan", "value":456,"total_value":789}`),
			BodyContains: []string{"PlanLine", `"name":"Modification de ligne de plan"`, `"value":456`, `"total_value":789`, `"descript":"Essai de description"`}},
	}
	for i, tc := range testCases {
		response := e.PUT("/api/plans/"+tc.ID+"/planlines/"+tc.PlanLineID).WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nModifyPlanLine[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nModifyPlanLine[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// deletePlanLineTest check if route is protected and plan line sent back is correct.
func deletePlanLineTest(e *httpexpect.Expect, t *testing.T, plID int) {
	testCases := []struct {
		Token        string
		Status       int
		ID           string
		PlanLineID   string
		BodyContains []string
	}{
		{Token: testCtx.User.Token, ID: "0", PlanLineID: "0", Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "1", PlanLineID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Suppression de ligne de plan, requête : Ligne de plan introuvable"}},
		{Token: testCtx.Admin.Token, ID: "1", PlanLineID: strconv.Itoa(plID), Status: http.StatusOK,
			BodyContains: []string{"Ligne de plan supprimée"}},
	}
	for i, tc := range testCases {
		response := e.DELETE("/api/plans/"+tc.ID+"/planlines/"+tc.PlanLineID).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nDeletePlanLine[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nDeletePlanLine[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.Status == http.StatusOK {
			content := string(e.GET("/api/plans/"+tc.ID+"/planlines").WithHeader("Authorization", "Bearer "+tc.Token).Expect().Content)
			if strings.Contains(content, `"id" : `+tc.PlanLineID) {
				t.Errorf("DeletePlanLine[%d] : identificateur %s trouvé après suppression :\n%s", i, tc.ID, content)
			}
		}
	}
}
