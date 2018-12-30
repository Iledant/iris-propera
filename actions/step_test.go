package actions

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testStep(t *testing.T) {
	testCommons(t)
	t.Run("Step", func(t *testing.T) {
		getStepsTest(testCtx.E, t)
		stID := createStepTest(testCtx.E, t)
		modifyStepTest(testCtx.E, t, stID)
		deleteStepTest(testCtx.E, t, stID)
	})
}

// getStepTest check route is protected and datas sent has got items and number of lines.
func getStepsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			//cSpell:disable
			BodyContains: []string{"Step", "name", "Protocole",
				"Travaux en cours (financés)", "Travaux préparatoires", "SDMR"}},
		//cSpell:enable
	}
	for i, tc := range testCases {
		response := e.GET("/api/steps").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetStep[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetStep[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// createStepTest check route is protected and datas sent has got correct datas.
func createStepTest(e *httpexpect.Expect, t *testing.T) (stID int) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest,
			Sent:         []byte(`{"name":""}`),
			BodyContains: []string{"Création d'étape : Name incorrect"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent:         []byte(`{"name":"Essai d'étape"}`),
			BodyContains: []string{"Step", `"name":"Essai d'étape"`}},
	}
	for i, tc := range testCases {
		response := e.POST("/api/steps").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nCreateStep[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nCreateStep[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.Status == http.StatusOK {
			stID = int(response.JSON().Object().Value("Step").Object().Value("id").Number().Raw())
		}
	}
	return stID
}

// modifyStepTest check route is protected and datas sent has got correct datas.
func modifyStepTest(e *httpexpect.Expect, t *testing.T, stID int) {
	testCases := []testCase{
		{Token: "fake", ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Modification d'étape"}`),
			BodyContains: []string{"Modification d'étape, requête : Etape introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(stID), Status: http.StatusOK,
			Sent:         []byte(`{"name":"Modification d'étape"}`),
			BodyContains: []string{"Step", `"name":"Modification d'étape"`}},
	}
	for i, tc := range testCases {
		response := e.PUT("/api/steps/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nModifyStep[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nModifyStep[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// deleteStepTest check route is protected and datas sent has got correct datas.
func deleteStepTest(e *httpexpect.Expect, t *testing.T, stID int) {
	testCases := []testCase{
		{Token: "fake", ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Suppression d'étape, requête : Etape introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(stID), Status: http.StatusOK,
			BodyContains: []string{"Etape supprimée"}},
	}
	for i, tc := range testCases {
		response := e.DELETE("/api/steps/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nDeleteStep[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nDeleteStep[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}
