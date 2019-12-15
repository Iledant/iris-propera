package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestCommission embeddes all tests for category insuring the configuration and DB are properly initialized.
func testCommission(t *testing.T) {
	t.Run("Commissions", func(t *testing.T) {
		getCommissionsTest(testCtx.E, t)
		coID := createCommissionTest(testCtx.E, t)
		if coID == 0 {
			t.Fatal("Impossible de créer la commission")
		}
		modifyCommissionTest(testCtx.E, t, coID)
		deleteCommissionTest(testCtx.E, t, coID)
	})
}

// getCommissionsTest tests route is protected and all commissions are sent back.
func getCommissionsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusOK,
			BodyContains: []string{"Commissions"},
			ArraySize:    8},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/commissions").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "Getcommission") {
		t.Error(r)
	}
}

// createCommissionTest tests route is protected and sent commission is created.
func createCommissionTest(e *httpexpect.Expect, t *testing.T) (coID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{}`),
			BodyContains: []string{"Création d'une commission : Name ou date incorrect"}},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Test création commission", "date":"2018-04-01T20:00:00Z"`),
			BodyContains: []string{"Création d'une commission, décodage :"}},
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusCreated,
			IDName:       `"id"`,
			Sent:         []byte(`{"name":"Test création commission", "date":"2018-04-01T20:00:00Z"}`),
			BodyContains: []string{"Commissions", `"name":"Test création commission"`, `"date":"2018-04-01T20:00:00Z"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/commissions").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "CreateCommission", &coID) {
		t.Error(r)
	}
	return coID
}

// modifyCommissionTest tests route is protected and modify work properly.
func modifyCommissionTest(e *httpexpect.Expect, t *testing.T, coID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Test modification commission","date":"2017-04-01T20:00:00Z"}`),
			BodyContains: []string{"Modification d'une commission, requête : Commission introuvable"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(coID),
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Test modification commission","date":"2017-04-01T20:00:00Z"`),
			BodyContains: []string{"Modification d'une commission, décodage :"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(coID),
			Status:       http.StatusOK,
			Sent:         []byte(`{"name":"Test modification commission","date":"2017-04-01T20:00:00Z"}`),
			BodyContains: []string{"Commissions", `"name":"Test modification commission"`, `"date":"2017-04-01T20:00:00Z"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.PUT("/api/commissions/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "ModifyCommission") {
		t.Error(r)
	}
}

// deleteCommissionTest tests route is protected and delete work properly.
func deleteCommissionTest(e *httpexpect.Expect, t *testing.T, coID int) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Suppression d'une commission, requête : Commission introuvable"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(coID),
			Status:       http.StatusOK,
			BodyContains: []string{"Commission supprimée"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.DELETE("/api/commissions/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DeleteCommission") {
		t.Error(r)
	}
}
