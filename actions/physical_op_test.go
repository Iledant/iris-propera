package actions

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

//TestPhysicalOps includes all tests for physical operation handler.
func TestPhysicalOps(t *testing.T) {
	TestCommons(t)
	t.Run("PhysicalOps", func(t *testing.T) {
		getPhysicalOpsTest(testCtx.E, t)
		opID := createPhysicalOpTest(testCtx.E, t)
		updatePhysicalOpTest(testCtx.E, t)
		deletePhysicalOpTest(testCtx.E, t, opID)
		batchPhysicalOpsTest(testCtx.E, t)
	})
}

// getPhysicalOpsTest tests if route is protected and returned list properly formatted.
func getPhysicalOpsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Status       int
		BodyContains string
		ArraySize    int
	}{
		{Token: "", Status: http.StatusInternalServerError, BodyContains: "Token absent", ArraySize: 0},
		{Token: testCtx.User.Token, Status: http.StatusOK, BodyContains: "PhysicalOp", ArraySize: 3},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, BodyContains: "PhysicalOp", ArraySize: 619},
	}

	for i, tc := range testCases {
		response := e.GET("/api/physical_ops").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		if !strings.Contains(content, tc.BodyContains) {
			t.Errorf("GetPhysicalOps[%d] : contenu incorrect, attendu \"%s\" et reçu\n\"%s\"", i, tc.BodyContains, content)
		}
		if tc.Status == http.StatusOK {
			response.JSON().Object().ContainsKey("PhysicalOp")
			response.JSON().Object().Value("PhysicalOp").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

//createPhysicalOpTest tests if route is protected, validations ok and number correctly computed.
func createPhysicalOpTest(e *httpexpect.Expect, t *testing.T) int {
	testCases := []struct {
		Token        string
		Status       int
		Sent         []byte
		BodyContains []string
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest, Sent: []byte(`{}`), BodyContains: []string{"Mauvais format de numéro d'opération"}},
		{Token: testCtx.Admin.Token, Sent: []byte(`{"number":"99XX001","name":""}`), Status: http.StatusBadRequest, BodyContains: []string{"Nom de l'opération absent"}},
		{Token: testCtx.Admin.Token, Sent: []byte(`{"number":"18VN044","name":"Essai fluvial","isr":true,"descript":"description","value":123456,"valuedate":"2018-08-21T02:00:00Z","length":123456,"tri":500,"van":123456}`),
			Status: http.StatusOK, BodyContains: []string{"PhysicalOp", `"number":"18VN045"`, `"name":"Essai fluvial"`, `"isr":true`, `"descript":"description"`, `"value":123456`, `"valuedate":"2018-08-21T00:00:00Z"`, `"length":123456`, `"tri":500`, `"van":123456`}},
	}

	var opID int
	for i, tc := range testCases {
		response := e.POST("/api/physical_ops").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, bc := range tc.BodyContains {
			if !strings.Contains(content, bc) {
				t.Errorf("CreatePhysicalOp[%d] : contenu incorrect, attendu \"%s\" et reçu\n\"%s\"", i, bc, content)
			}
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			opID = int(response.JSON().Object().Value("PhysicalOp").Object().Value("id").Number().Raw())
		}
	}
	return opID
}

// deletePhysicalOpTest tests if route is protected and destroy operation previously created.
func deletePhysicalOpTest(e *httpexpect.Expect, t *testing.T, opID int) {
	sOpID := strconv.Itoa(opID)
	testCases := []struct {
		Token        string
		OpID         string
		Status       int
		BodyContains string
	}{
		{Token: testCtx.User.Token, OpID: sOpID, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis"},
		{Token: testCtx.Admin.Token, OpID: "0", Status: http.StatusNotFound, BodyContains: "Opération introuvable"},
		{Token: testCtx.Admin.Token, OpID: sOpID, Status: http.StatusOK, BodyContains: "Opération supprimée"},
	}

	for _, tc := range testCases {
		response := e.DELETE("/api/physical_ops/"+tc.OpID).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		response.Body().Contains(tc.BodyContains)
		response.Status(tc.Status)
	}
}

//updatePhysicalOpTest tests if route is protected and fields properly updated according to role.
func updatePhysicalOpTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		opID         string
		Status       int
		Sent         []byte
		BodyContains []string
	}{
		{Token: "", opID: "0", Status: http.StatusInternalServerError, Sent: []byte(`{}`), BodyContains: []string{"Token absent"}},
		{Token: testCtx.User.Token, opID: "0", Status: http.StatusNotFound, Sent: []byte(`{}`), BodyContains: []string{"Opération introuvable"}},
		{Token: testCtx.User.Token, opID: "15", Status: http.StatusBadRequest, Sent: []byte(`{}`), BodyContains: []string{"L'utilisateur n'a pas de droits sur l'opération"}},
		{Token: testCtx.Admin.Token, opID: "14", Sent: []byte(`{"number":"01DI001"}`),
			Status: http.StatusBadRequest, BodyContains: []string{"Numéro d'opération existant"}},
		{Token: testCtx.User.Token, opID: "14", Sent: []byte(`{"name":"Nouveau nom","isr":true,"descript":"Nouvelle description","value":123456,"valuedate":"2018-08-17T00:00:00Z","length":123456,"tri":500,"van":123456,"plan_line_id":34}`),
			Status: http.StatusOK, BodyContains: []string{"PhysicalOp", `"name":"Bus - voirie - aménagement"`, `"isr":true`, `"descript":"Nouvelle description"`, `"value":123456`, `"valuedate":"2018-08-17T00:00:00Z"`, `"length":123456`, `"tri":500`, `"van":123456`, `"plan_line_id":32`}},
		{Token: testCtx.Admin.Token, opID: "14", Sent: []byte(`{"name":"Nom nouveau","isr":false,"descript":"Description nouvelle","value":546,"valuedate":"2018-08-16T00:00:00Z","length":546,"tri":300,"van":100,"plan_line_id":34}`),
			Status: http.StatusOK, BodyContains: []string{"PhysicalOp", `"name":"Nom nouveau"`, `"isr":false`, `"descript":"Description nouvelle"`, `"value":546`, `"valuedate":"2018-08-16T00:00:00Z"`, `"length":546`, `"tri":300`, `"van":100`, `"plan_line_id":34`}},
	}

	for i, tc := range testCases {
		response := e.PUT("/api/physical_ops/"+tc.opID).WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, bc := range tc.BodyContains {
			if !strings.Contains(content, bc) {
				t.Errorf("CreatePhysicalOp[%d] : contenu incorrect, attendu \"%s\" et reçu\n\"%s\"", i, bc, content)
			}
		}
		response.Status(tc.Status)
	}
}

// batchPhysicalOpsTest tests if route is protected and import passed.
func batchPhysicalOpsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Sent         []byte
		Status       int
		BodyContains string
	}{
		{Token: testCtx.User.Token, Sent: []byte(`{"PhysicalOp":[]}`),
			Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis"},
		{Token: testCtx.Admin.Token, Sent: []byte(`{"PhysicalOp":[{"number":"20XX999","isr":true}]}`),
			Status: http.StatusInternalServerError, BodyContains: "Erreur d'insertion"},
		{Token: testCtx.Admin.Token, Sent: []byte(`{"PhysicalOp":[{"number":"20XX001","name":"Essai batch1","isr":true},{"number":"18DI999","name":"Essai batch2","isr":true}]}`),
			Status: http.StatusOK, BodyContains: "Terminé"},
		{Token: testCtx.Admin.Token, Sent: []byte(`{"PhysicalOp":[{"number":"20XX003","name":"Essai batch3","isr":true,"descript":"Description batch3","value":123,"valuedate":"2018-08-15T00:00:00Z","length":123,"step":"Protocole","category":"Route","tri":500,"van":123,"action":"17700101","payment_type_id":4,"plan_line_id":20}]}`),
			Status: http.StatusOK, BodyContains: "Terminé"},
	}

	for i, tc := range testCases {
		response := e.POST("/api/physical_ops/array").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		if !strings.Contains(content, tc.BodyContains) {
			t.Errorf("Batch physical_ops[%d] : contenu incorrect, attendu \"%s\" et reçu\n\"%s\"", i, tc.BodyContains, content)
		}
		response.Status(tc.Status)
	}

	response := e.GET("/api/physical_ops").WithHeader("Authorization", "Bearer "+testCtx.Admin.Token).Expect()
	ee := []string{"Essai batch1", "Essai batch2", "Essai batch3", "Description batch3", "20XX001"}
	content := string(response.Content)
	for _, e := range ee {
		if !strings.Contains(content, e) {
			t.Errorf("Batch physical_ops[GET] : attendu \"%s\" et reçu\n\"%s\"", e, content)
		}
	}
}
