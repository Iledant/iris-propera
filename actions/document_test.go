package actions

import (
	"net/http"
	"strconv"
	"strings"
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
	testCases := []testCase{
		{Token: "fake", ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}, ArraySize: 0},
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusOK,
			BodyContains: []string{`"Document":null`}, ArraySize: 0},
		{Token: testCtx.User.Token, ID: "403", Status: http.StatusOK,
			BodyContains: []string{"Document"}, ArraySize: 1},
	}

	for i, tc := range testCases {
		response := e.GET("/api/physical_ops/"+tc.ID+"/documents").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetDocuments[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetDocuments[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetDocuments[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// createDocumentTest tests route is protected and sent document is created.
func createDocumentTest(e *httpexpect.Expect, t *testing.T) (doID int) {
	testCases := []testCase{
		{Token: "fake", ID: "403", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, ID: "403", Status: http.StatusBadRequest, Sent: []byte(`{}`),
			BodyContains: []string{"Création d'un document : PhysicalOpID, Name ou Link incorrect"}},
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusBadRequest,
			Sent:         []byte(`{"name":"Test création document", "link":"Test création lien document"}`),
			BodyContains: []string{"Création d'un document : PhysicalOpID, Name ou Link incorrect"}},
		{Token: testCtx.User.Token, ID: "403", Status: http.StatusOK,
			Sent:         []byte(`{"name":"Test création document", "link":"Test création lien document"}`),
			BodyContains: []string{"Document", `"name":"Test création document"`, `"link":"Test création lien document"`}},
	}

	for i, tc := range testCases {
		response := e.POST("/api/physical_ops/"+tc.ID+"/documents").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nCreateDocument[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nCreateDocument[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.Status == http.StatusOK {
			doID = int(response.JSON().Object().Value("Document").Object().Value("id").Number().Raw())
		}
	}
	return doID
}

// modifyDocumentTest tests route is protected and modify work properly.
func modifyDocumentTest(e *httpexpect.Expect, t *testing.T, doID int) {
	testCases := []testCase{
		{Token: "fake", ID: "403", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Test modification document", "link":"Test modification lien document"}`),
			BodyContains: []string{"Modification d'un document, requête : Document introuvable"}},
		{Token: testCtx.User.Token, ID: strconv.Itoa(doID), Status: http.StatusOK,
			Sent:         []byte(`{"name":"Test modification document", "link":"Test modification lien document"}`),
			BodyContains: []string{"Document", `"name":"Test modification document"`, `"link":"Test modification lien document"`}},
	}

	for i, tc := range testCases {
		response := e.PUT("/api/physical_ops/403/documents/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nModifyDocument[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nModifyDocument[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}

	}
}

// deleteDocumentTest tests route is protected and delete work properly.
func deleteDocumentTest(e *httpexpect.Expect, t *testing.T, doID int) {
	testCases := []testCase{
		{Token: "fake", ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Suppression d'un document, requête : Document introuvable"}},
		{Token: testCtx.User.Token, ID: strconv.Itoa(doID), Status: http.StatusOK,
			BodyContains: []string{"Document supprimé"}},
	}

	for i, tc := range testCases {
		response := e.DELETE("/api/physical_ops/403/documents/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nDeleteDocument[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nDeleteDocument[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}

	}
}
