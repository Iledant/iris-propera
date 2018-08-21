package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestDocument embeddes all tests for document insuring the configuration and DB are properly initialized.
func TestDocument(t *testing.T) {
	TestCommons(t)
	t.Run("Document", func(t *testing.T) {
		getDocumentTest(testCtx.E, t)
		doID := createDocumentTest(testCtx.E, t)
		modifyDocumentTest(testCtx.E, t, doID)
		deleteDocumentTest(testCtx.E, t, doID)
	})
}

// getDocumentTest tests route is protected and all documents are sent back.
func getDocumentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Status       int
		OpID         string
		BodyContains string
		ArraySize    int
	}{
		{Token: "fake", OpID: "0", Status: http.StatusInternalServerError, BodyContains: "Token invalide", ArraySize: 0},
		{Token: testCtx.User.Token, OpID: "0", Status: http.StatusBadRequest, BodyContains: "Liste des documents : opération introuvable", ArraySize: 0},
		{Token: testCtx.User.Token, OpID: "403", Status: http.StatusOK, BodyContains: "Document", ArraySize: 1},
	}

	for _, tc := range testCases {
		response := e.GET("/api/physical_ops/"+tc.OpID+"/documents").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("Document").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

// createDocumentTest tests route is protected and sent document is created.
func createDocumentTest(e *httpexpect.Expect, t *testing.T) int {
	testCases := []struct {
		Token        string
		opID         string
		Status       int
		Sent         []byte
		BodyContains []string
	}{
		{Token: "fake", opID: "403", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, opID: "403", Status: http.StatusBadRequest, Sent: []byte(`{}`), BodyContains: []string{"Création de document, champ manquant ou incorrect"}},
		{Token: testCtx.User.Token, opID: "0", Status: http.StatusBadRequest, Sent: []byte(`{}`), BodyContains: []string{"Création de document : opération introuvable"}},
		{Token: testCtx.User.Token, opID: "403", Status: http.StatusOK, Sent: []byte(`{"name":"Test création document", "link":"Test création lien document"}`), BodyContains: []string{"Document", `"name":"Test création document"`, `"link":"Test création lien document"`}},
	}
	var doID int

	for _, tc := range testCases {
		response := e.POST("/api/physical_ops/"+tc.opID+"/documents").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		if tc.Status == http.StatusOK {
			doID = int(response.JSON().Object().Value("Document").Object().Value("id").Number().Raw())
		}
		response.Status(tc.Status)
	}
	return doID
}

// modifyDocumentTest tests route is protected and modify work properly.
func modifyDocumentTest(e *httpexpect.Expect, t *testing.T, doID int) {
	testCases := []struct {
		Token        string
		Status       int
		ID           string
		Sent         []byte
		BodyContains []string
	}{
		{Token: "fake", ID: "403", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusBadRequest, BodyContains: []string{"Modification de document : introuvable"}},
		{Token: testCtx.User.Token, ID: strconv.Itoa(doID), Status: http.StatusOK, Sent: []byte(`{"name":"Test modification document", "link":"Test modification lien document"}`), BodyContains: []string{"Document", `"name":"Test modification document"`, `"link":"Test modification lien document"`}},
	}

	for _, tc := range testCases {
		response := e.PUT("/api/physical_ops/403/documents/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}

// deleteDocumentTest tests route is protected and delete work properly.
func deleteDocumentTest(e *httpexpect.Expect, t *testing.T, doID int) {
	testCases := []struct {
		Token        string
		Status       int
		ID           string
		BodyContains string
	}{
		{Token: "fake", ID: "0", Status: http.StatusInternalServerError, BodyContains: "Token invalide"},
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusNotFound, BodyContains: "Suppression de document : introuvable"},
		{Token: testCtx.User.Token, ID: strconv.Itoa(doID), Status: http.StatusOK, BodyContains: "Document supprimé"},
	}

	for _, tc := range testCases {
		response := e.DELETE("/api/physical_ops/403/documents/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		response.Body().Contains(tc.BodyContains)
		response.Status(tc.Status)
	}
}
