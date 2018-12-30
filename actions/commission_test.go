package actions

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestCommission embeddes all tests for category insuring the configuration and DB are properly initialized.
func testCommission(t *testing.T) {
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

	for i, tc := range testCases {
		response := e.GET("/api/commissions").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetCommission[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetCommission[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetCommission[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// createCommissionTest tests route is protected and sent commission is created.
func createCommissionTest(e *httpexpect.Expect, t *testing.T) (coID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest, Sent: []byte(`{}`),
			BodyContains: []string{"Création d'une commission : Name ou date incorrect"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Sent: []byte(`{"name":"Test création commission", "date":"2018-04-01T20:00:00Z"}`),
			BodyContains: []string{"Commissions", `"name":"Test création commission"`, `"date":"2018-04-01T20:00:00Z"`}},
	}
	for i, tc := range testCases {
		response := e.POST("/api/commissions").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nCreateCommission[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nCreateCommission[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.Status == http.StatusOK {
			coID = int(response.JSON().Object().Value("Commissions").Object().Value("id").Number().Raw())
		}
	}
	return coID
}

// modifyCommissionTest tests route is protected and modify work properly.
func modifyCommissionTest(e *httpexpect.Expect, t *testing.T, coID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			Sent:         []byte(`{"name":"Test modification commission","date":"2017-04-01T20:00:00Z"}`),
			BodyContains: []string{"Modification d'une commission, requête : Commission introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(coID), Status: http.StatusOK,
			Sent:         []byte(`{"name":"Test modification commission","date":"2017-04-01T20:00:00Z"}`),
			BodyContains: []string{"Commissions", `"name":"Test modification commission"`, `"date":"2017-04-01T20:00:00Z"`}},
	}
	for i, tc := range testCases {
		response := e.PUT("/api/commissions/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nModifyCommission[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nModifyCommission[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// deleteCommissionTest tests route is protected and delete work properly.
func deleteCommissionTest(e *httpexpect.Expect, t *testing.T, coID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Suppression d'une commission, requête : Commission introuvable"}},
		{Token: testCtx.Admin.Token, ID: strconv.Itoa(coID), Status: http.StatusOK,
			BodyContains: []string{"Commission supprimée"}},
	}
	for i, tc := range testCases {
		response := e.DELETE("/api/commissions/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nDeleteCommission[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nDeleteCommission[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}
