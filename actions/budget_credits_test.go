package actions

import (
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Iledant/iris_propera/models"
	"github.com/iris-contrib/httpexpect"
)

// TestBudgetCredit embeddes all tests for budget credits insuring the configuration and DB are properly initialized.
func TestBudgetCredit(t *testing.T) {
	TestCommons(t)
	t.Run("BudgetCredit", func(t *testing.T) {
		getBudgetCredits(testCtx.E, t)
		getLastBudgetCredits(testCtx.E, t)
		brID := createBudgetCreditTest(testCtx.E, t)
		t.Logf("brID : %d\n", brID)
		modifyBudgetCreditTest(testCtx.E, t, brID)
		deleteBudgetCreditTest(testCtx.E, t, brID)
		batchBudgetCreditTest(testCtx.E, t)
	})
}

// getBudgetCredits tests route is protected and all credits are sent back.
func getBudgetCredits(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Status       int
		BodyContains string
		ArraySize    int
	}{
		{Token: "", Status: http.StatusInternalServerError, BodyContains: "Token absent", ArraySize: 0},
		{Token: testCtx.User.Token, Status: http.StatusOK, BodyContains: "BudgetCredits", ArraySize: 75},
	}

	for _, tc := range testCases {
		response := e.GET("/api/budget_credits").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("BudgetCredits").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

// getLastBudgetCredits tests route is protected and all credits are sent back.
func getLastBudgetCredits(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Status       int
		BodyContains string
		ArraySize    int
	}{
		{Token: "", Status: http.StatusInternalServerError, BodyContains: "Token absent", ArraySize: 0},
		{Token: testCtx.User.Token, Status: http.StatusOK, BodyContains: "BudgetCredits", ArraySize: 3},
	}

	for _, tc := range testCases {
		response := e.GET("/api/budget_credits/year").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("BudgetCredits").Array().Length().Equal(tc.ArraySize)
			response.Body().Contains("2018-07-04")
		}
		response.Status(tc.Status)
	}
}

// createBudgetCreditTest tests route is protected and sent action is created.
func createBudgetCreditTest(e *httpexpect.Expect, t *testing.T) int {
	var p, r, f int64 = 123456, 123, 456
	testCases := []struct {
		Token        string
		Status       int
		BudgetCredit sentBrReq
		BodyContains string
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis"},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest, BudgetCredit: sentBrReq{}, BodyContains: "Création de crédits : champ manquant ou incorrect"},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest, BudgetCredit: sentBrReq{
			CommissionDate: models.NullTime{Valid: true, Time: time.Now()},
			ChapterID:      models.NullInt64{Valid: true, Int64: 1}}, BodyContains: "Création de crédits : champ manquant ou incorrect"},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, BudgetCredit: sentBrReq{
			CommissionDate: models.NullTime{Valid: true, Time: time.Now()},
			ChapterID:      models.NullInt64{Valid: true, Int64: 1}, PrimaryCommitment: &p,
			ReservedCommitment: &r, FrozenCommitment: &f}, BodyContains: "BudgetCredits"},
	}
	var brID int

	for _, tc := range testCases {
		response := e.POST("/api/budget_credits").WithHeader("Authorization", "Bearer "+tc.Token).WithJSON(tc.BudgetCredit).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("BudgetCredits").Object().Value("chapter_id").Number().Equal(tc.BudgetCredit.ChapterID.Int64)
			response.JSON().Object().Value("BudgetCredits").Object().Value("primary_commitment").Number().Equal(*tc.BudgetCredit.PrimaryCommitment)
			response.JSON().Object().Value("BudgetCredits").Object().Value("reserved_commitment").Number().Equal(*tc.BudgetCredit.ReservedCommitment)
			response.JSON().Object().Value("BudgetCredits").Object().Value("frozen_commitment").Number().Equal(*tc.BudgetCredit.FrozenCommitment)
			brID = int(response.JSON().Object().Value("BudgetCredits").Object().Value("id").Number().Raw())
		}
		response.Status(tc.Status)
	}
	return brID
}

// modifyBudgetCreditTest tests route is protected and modify work properly.
func modifyBudgetCreditTest(e *httpexpect.Expect, t *testing.T, brID int) {
	testCases := []struct {
		Token        string
		ID           int
		Status       int
		Sent         []byte
		BodyContains string
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis"},
		{Token: testCtx.Admin.Token, ID: 0, Status: http.StatusBadRequest, Sent: []byte("{\"chapter_id\":2}"), BodyContains: "Modification des crédits: introuvable"},
		{Token: testCtx.Admin.Token, ID: brID, Status: http.StatusOK, Sent: []byte("{\"chapter_id\":2}"), BodyContains: "BudgetCredits"},
		{Token: testCtx.Admin.Token, ID: brID, Status: http.StatusOK, Sent: []byte("{\"commission_date\":\"2018-08-13T09:21:56.132Z\"}"), BodyContains: "BudgetCredits"},
		{Token: testCtx.Admin.Token, ID: brID, Status: http.StatusOK, Sent: []byte("{\"primary_commitment\":999}"), BodyContains: "BudgetCredits"},
		{Token: testCtx.Admin.Token, ID: brID, Status: http.StatusOK, Sent: []byte("{\"frozen_commitment\":888}"), BodyContains: "BudgetCredits"},
		{Token: testCtx.Admin.Token, ID: brID, Status: http.StatusOK, Sent: []byte("{\"reserved_commitment\":777}"), BodyContains: "BudgetCredits"},
	}

	for _, tc := range testCases {
		response := e.PUT("/api/budget_credits/"+strconv.Itoa(tc.ID)).WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.Status == http.StatusOK {
			s := strings.Trim(strings.Trim(string(tc.Sent), "{"), "}")
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}

// deleteBudgetCreditTest tests route is protected and modify work properly.
func deleteBudgetCreditTest(e *httpexpect.Expect, t *testing.T, brID int) {
	testCases := []struct {
		Token        string
		Status       int
		ID           int
		BodyContains string
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis"},
		{Token: testCtx.Admin.Token, ID: 0, Status: http.StatusBadRequest, BodyContains: "Suppression de crédits: introuvable"},
		{Token: testCtx.Admin.Token, ID: brID, Status: http.StatusOK, BodyContains: "Crédits supprimés"},
	}

	for _, tc := range testCases {
		response := e.DELETE("/api/budget_credits/"+strconv.Itoa(tc.ID)).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		response.Body().Contains(tc.BodyContains)
		response.Status(tc.Status)
	}
}

//batchBudgetCreditTest tests route is protected and sent datas are correctly incorporated.
func batchBudgetCreditTest(e *httpexpect.Expect, t *testing.T) {
	s0 := []byte(`fake`)
	s1 := []byte(`{"BudgetCredits":[
		{"commission_date":"2018-04-01T00:00:00Z"}
		]}`)
	s2 := []byte(`{"BudgetCredits":[
		{"commission_date":"2018-07-04T20:00:00Z","chapter":907,
		 "primary_commitment":999,"reserved_commitment":888,"frozen_commitment":777},
		{"commission_date":"2018-04-01T00:00:00.000Z","chapter":907,
		 "primary_commitment":999,"reserved_commitment":888,"frozen_commitment":777}
		]}`)
	testCases := []struct {
		Token        string
		Status       int
		Sent         []byte
		BodyContains string
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis"},
		{Token: testCtx.Admin.Token, Sent: s0, Status: http.StatusInternalServerError, BodyContains: "Erreur de lecture du batch crédits"},
		{Token: testCtx.Admin.Token, Sent: s1, Status: http.StatusBadRequest, BodyContains: "Batch crédits, champs manquants"},
		{Token: testCtx.Admin.Token, Sent: s2, Status: http.StatusOK, BodyContains: "Credits importés"},
	}

	for _, tc := range testCases {
		response := e.POST("/api/budget_credits/array").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		response.Body().Contains(tc.BodyContains)
		response.Status(tc.Status)
	}
	// Check only one line has been created
	response := e.GET("/api/budget_credits").WithHeader("Authorization", "Bearer "+testCtx.Admin.Token).Expect()
	response.JSON().Object().Value("BudgetCredits").Array().Length().Equal(76)
}
