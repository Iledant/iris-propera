package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testStep(t *testing.T) {
	testCommons(t)
	t.Run("Step", func(t *testing.T) {
		getStepsTest(testCtx.E, t)
		stID := createStepTest(testCtx.E, t)
		if stID == 0 {
			t.Fatal("Impossible de créer l'étape")
		}
		modifyStepTest(testCtx.E, t, stID)
		deleteStepTest(testCtx.E, t, stID)
	})
}

// getStepTest check route is protected and datas sent has got items and number of lines.
func getStepsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:  testCtx.User.Token,
			Status: http.StatusOK,
			//cSpell:disable
			BodyContains: []string{"Step", "name", "Protocole",
				"Travaux en cours (financés)", "Travaux préparatoires", "SDMR"}},
		//cSpell:enable
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/steps").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetSteps") {
		t.Error(r)
	}
}

// createStepTest check route is protected and datas sent has got correct datas.
func createStepTest(e *httpexpect.Expect, t *testing.T) (stID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{"name":""}`),
			BodyContains: []string{"Création d'étape : Name incorrect"}},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusCreated,
			Sent:         []byte(`{"name":"Essai d'étape"}`),
			IDName:       `"id"`,
			BodyContains: []string{"Step", `"name":"Essai d'étape"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/steps").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "CreateStep", &stID) {
		t.Error(r)
	}
	return stID
}

// modifyStepTest check route is protected and datas sent has got correct datas.
func modifyStepTest(e *httpexpect.Expect, t *testing.T, stID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Modification d'étape"}`),
			BodyContains: []string{"Modification d'étape, requête : Etape introuvable"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(stID),
			Status:       http.StatusOK,
			Sent:         []byte(`{"name":"Modification d'étape"}`),
			BodyContains: []string{"Step", `"name":"Modification d'étape"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.PUT("/api/steps/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "ModifyStep") {
		t.Error(r)
	}
}

// deleteStepTest check route is protected and datas sent has got correct datas.
func deleteStepTest(e *httpexpect.Expect, t *testing.T, stID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Suppression d'étape, requête : Etape introuvable"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(stID),
			Status:       http.StatusOK,
			BodyContains: []string{"Etape supprimée"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.DELETE("/api/steps/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DeleteStep") {
		t.Error(r)
	}
}
