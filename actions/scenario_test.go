package actions

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func TestScenario(t *testing.T) {
	TestCommons(t)
	t.Run("Scenario", func(t *testing.T) {
		getScenarioTest(testCtx.E, t)
		ID := createScenarioTest(testCtx.E, t)
		modifyScenarioTest(testCtx.E, t, ID)
		getScenarioDatasTest(testCtx.E, t, ID)
		setScenarioOffsetsText(testCtx.E, t, ID)
		deleteScenarioTest(testCtx.E, t, ID)
	})
}

// getScenarioTest check route is protected and datas sent has got items and number of lines.
func getScenarioTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{"Scenario", "name", "descript", "Scénario 750 M€ pour 2018"}},
	}
	for i, tc := range testCases {
		response := e.GET("/api/scenarios").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetScenarios[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetScenarios[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// createScenarioTest check route is protected and created scenarios works properly.
func createScenarioTest(e *httpexpect.Expect, t *testing.T) (ID int) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent:         []byte(`{"name":"Scénario créé","descript":"Description du scénario créé"}`),
			BodyContains: []string{"Scenario", "name", "Scénario créé", "descript", "Description du scénario créé"}},
	}
	for i, tc := range testCases {
		response := e.POST("/api/scenarios").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nCreateScenario[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nCreateScenario[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.Status == http.StatusOK {
			ID = int(response.JSON().Object().Value("Scenario").Object().Value("id").Number().Raw())
		}

	}
	return ID
}

// modifyScenarioTest check route is protected and modify works properly.
func modifyScenarioTest(e *httpexpect.Expect, t *testing.T, ID int) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			ID:           "0",
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			ID:           "0",
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusInternalServerError,
			ID:           "0",
			Sent:         []byte(`{"name":"Scénario modifié","descript":"Description du scénario modifié"}`),
			BodyContains: []string{"Modification de scénario, requête : Scenario introuvable"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			ID:           strconv.Itoa(ID),
			Sent:         []byte(`{"name":"Scénario modifié","descript":"Description du scénario modifié"}`),
			BodyContains: []string{"Scenario", "name", "Scénario modifié", "descript", "Description du scénario modifié"}},
	}
	for i, tc := range testCases {
		response := e.PUT("/api/scenarios/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nModifyScenario[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nModifyScenario[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// deleteScenarioTest check route is protected and delete works properly.
func deleteScenarioTest(e *httpexpect.Expect, t *testing.T, ID int) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			ID:           "0",
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			ID:           "0",
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusInternalServerError,
			ID:           "0",
			BodyContains: []string{"Suppression de scénario, requête : Scenario introuvable"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			ID:           strconv.Itoa(ID),
			BodyContains: []string{"Scenario supprimé"}},
	}
	for i, tc := range testCases {
		response := e.DELETE("/api/scenarios/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nDeleteScenario[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nDeleteScenario[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// getScenarioDatasTest check route is protected and correct objects sent back.
func getScenarioDatasTest(e *httpexpect.Expect, t *testing.T, ID int) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			ID:           "0",
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			ID:           "0",
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusInternalServerError,
			ID:           "0",
			Param:        "2018",
			BodyContains: []string{"Datas d'un scénario, requête : "}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			ID:           "1",
			Param:        "2018",
			BodyContains: []string{"OperationCrossTable", "ScenarioCrossTable"}},
	}
	for i, tc := range testCases {
		response := e.GET("/api/scenarios/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetScenarioDatas[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetScenarioDatas[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// setScenarioOffsetsText check route is protected and offset add return ok.
func setScenarioOffsetsText(e *httpexpect.Expect, t *testing.T, ID int) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			ID:           "0",
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			ID:           "0",
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusInternalServerError,
			ID: "0",
			Sent: []byte(`{"offsetList":[{"physical_op_id":220,"offset":0},{"physical_op_id":546,"offset":1},
			{"physical_op_id":9,"offset":0},{"physical_op_id":543,"offset":2}]}`),
			BodyContains: []string{"Offsets de scénario, requête : pq:"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			ID: strconv.Itoa(ID),
			Sent: []byte(`{"offsetList":[{"physical_op_id":220,"offset":0},{"physical_op_id":546,"offset":1},
			{"physical_op_id":9,"offset":0},{"physical_op_id":543,"offset":2}]}`),
			BodyContains: []string{"Offsets créés"}},
	}
	for i, tc := range testCases {
		response := e.POST("/api/scenarios/"+tc.ID+"/offsets").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nSetScenarioOffsets[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nSetScenarioOffsets[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}
