package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestCommission embeddes all tests for category insuring the configuration and DB are properly initialized.
func TestCommission(t *testing.T) {
	TestCommons(t)
	t.Run("Commissions", func(t *testing.T) {
		getCommissionsTest(testCtx.E, t)
		coID := createCommissionTest(testCtx.E, t)
		modifyCommissionTest(testCtx.E, t, coID)
		deleteCommissionTest(testCtx.E, t, coID)
	})
}

// getCommissionsTest tests route is protected and all commissions are sent back.
func getCommissionsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}, ArraySize: 0},
		{Token: testCtx.User.Token, Status: http.StatusOK, BodyContains: []string{"Commissions"}, ArraySize: 8},
	}

	for _, tc := range testCases {
		response := e.GET("/api/commissions").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("Commissions").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

// createCommissionTest tests route is protected and sent commission is created.
func createCommissionTest(e *httpexpect.Expect, t *testing.T) int {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest, Sent: []byte(`{}`),
			BodyContains: []string{"Création de commission, champ manquant ou incorrect"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Sent: []byte(`{"name":"Test création commission", "date":"2018-04-01T20:00:00Z"}`),
			BodyContains: []string{"Commissions", `"name":"Test création commission"`, `"date":"2018-04-01T20:00:00Z"`}},
	}
	var coID int

	for _, tc := range testCases {
		response := e.POST("/api/commissions").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		if tc.Status == http.StatusOK {
			coID = int(response.JSON().Object().Value("Commissions").Object().Value("id").Number().Raw())
		}
		response.Status(tc.Status)
	}
	return coID
}

// modifyCommissionTest tests route is protected and modify work properly.
func modifyCommissionTest(e *httpexpect.Expect, t *testing.T, coID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusBadRequest,
			BodyContains: []string{"Modification de commission : introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(coID), Status: http.StatusOK,
			Sent:         []byte(`{"name":"Test modification commission","date":"2017-04-01T20:00:00Z"}`),
			BodyContains: []string{"Commissions", `"name":"Test modification commission"`, `"date":"2017-04-01T20:00:00Z"`}},
	}

	for _, tc := range testCases {
		response := e.PUT("/api/commissions/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}

// deleteCommissionTest tests route is protected and delete work properly.
func deleteCommissionTest(e *httpexpect.Expect, t *testing.T, coID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusNotFound,
			BodyContains: []string{"Suppression de commission : introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(coID), Status: http.StatusOK,
			BodyContains: []string{"Commission supprimée"}},
	}

	for _, tc := range testCases {
		response := e.DELETE("/api/commissions/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}
