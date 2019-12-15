package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestDocument embeddes all tests for document insuring the configuration and DB are properly initialized.
func testDocument(t *testing.T) {
	t.Run("Document", func(t *testing.T) {
		getDocumentTest(testCtx.E, t)
		doID := createDocumentTest(testCtx.E, t)
		if doID == 0 {
			t.Fatal("Impossible de créer le document")
		}
		modifyDocumentTest(testCtx.E, t, doID)
		deleteDocumentTest(testCtx.E, t, doID)
	})
}

// getDocumentTest tests route is protected and all documents are sent back.
func getDocumentTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			ID:           "0",
			Status:       http.StatusOK,
			BodyContains: []string{`"Document":[]`},
			IDName:       `"id"`,
			ArraySize:    0},
		{
			Token:        testCtx.User.Token,
			ID:           "403",
			Status:       http.StatusOK,
			BodyContains: []string{"Document"},
			IDName:       `"id"`,
			ArraySize:    1},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/physical_ops/"+tc.ID+"/documents").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetDocuments") {
		t.Error(r)
	}
}

// createDocumentTest tests route is protected and sent document is created.
func createDocumentTest(e *httpexpect.Expect, t *testing.T) (doID int) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			ID:           "403",
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{}`),
			BodyContains: []string{"Création d'un document : PhysicalOpID, Name ou Link incorrect"}},
		{
			Token:        testCtx.User.Token,
			ID:           "0",
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{"name":"Test création document", "link":"Test création lien document"}`),
			BodyContains: []string{"Création d'un document : PhysicalOpID, Name ou Link incorrect"}},
		{
			Token:        testCtx.User.Token,
			ID:           "403",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Test création document", "link":"Test création lien document"`),
			BodyContains: []string{"Création d'un document, décodage :"}},
		{
			Token:        testCtx.User.Token,
			ID:           "403",
			Status:       http.StatusCreated,
			IDName:       `"id"`,
			Sent:         []byte(`{"name":"Test création document", "link":"Test création lien document"}`),
			BodyContains: []string{"Document", `"name":"Test création document"`, `"link":"Test création lien document"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/physical_ops/"+tc.ID+"/documents").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "CreateDocument", &doID) {
		t.Error(r)
	}
	return doID
}

// modifyDocumentTest tests route is protected and modify work properly.
func modifyDocumentTest(e *httpexpect.Expect, t *testing.T, doID int) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Test modification document", "link":"Test modification lien document"}`),
			BodyContains: []string{"Modification d'un document, requête : Document introuvable"}},
		{
			Token:        testCtx.User.Token,
			ID:           strconv.Itoa(doID),
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Test modification document", "link":"Test modification lien document"`),
			BodyContains: []string{"Modification d'un document, décodage :"}},
		{
			Token:        testCtx.User.Token,
			ID:           strconv.Itoa(doID),
			Status:       http.StatusOK,
			Sent:         []byte(`{"name":"Test modification document", "link":"Test modification lien document"}`),
			BodyContains: []string{"Document", `"name":"Test modification document"`, `"link":"Test modification lien document"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.PUT("/api/physical_ops/403/documents/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "ModifyDocument") {
		t.Error(r)
	}
}

// deleteDocumentTest tests route is protected and delete work properly.
func deleteDocumentTest(e *httpexpect.Expect, t *testing.T, doID int) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Suppression d'un document, requête : Document introuvable"}},
		{
			Token:        testCtx.User.Token,
			ID:           strconv.Itoa(doID),
			Status:       http.StatusOK,
			BodyContains: []string{"Document supprimé"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.DELETE("/api/physical_ops/403/documents/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DeleteDocument") {
		t.Error(r)
	}
}
