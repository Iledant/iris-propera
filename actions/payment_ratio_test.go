package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func TestPaymentRatio(t *testing.T) {
	TestCommons(t)
	t.Run("PaymentRatios", func(t *testing.T) {
		getRatiosTest(testCtx.E, t)
		getPtRatiosTest(testCtx.E, t)
		setPtRatiosTest(testCtx.E, t)
		deletePtRatiosTest(testCtx.E, t)
		getYearRatiosTest(testCtx.E, t)
	})
}

// getRatiosTest check route is protected and ratios correctly sent.
func getRatiosTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Status       int
		BodyContains []string
		Count        int
	}{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK, BodyContains: []string{"PaymentRatio"}, Count: 26},
	}
	for _, tc := range testCases {
		response := e.GET("/api/payment_ratios").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("PaymentRatio").Array().Length().Equal(tc.Count)
		}
	}
}

// getPtRatiosTest check route is protected and ratios correctly sent.
func getPtRatiosTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		PtID         string
		Status       int
		BodyContains []string
		Count        int
	}{
		{Token: "fake", PtID: "0", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusBadRequest,
			PtID: "0", BodyContains: []string{"Liste des ratios : chronique introuvable"}},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			PtID: "5", BodyContains: []string{"PaymentRatio"}, Count: 8},
	}
	for _, tc := range testCases {
		response := e.GET("/api/payment_types/"+tc.PtID+"/payment_ratios").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("PaymentRatio").Array().Length().Equal(tc.Count)
		}
	}
}

// setPtRatiosTest check route is protected and ratios correctly set.
func setPtRatiosTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		PtID         string
		Status       int
		Sent         []byte
		BodyContains []string
		Count        int
	}{
		{Token: testCtx.User.Token, PtID: "0", Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, PtID: "0", Status: http.StatusBadRequest, BodyContains: []string{"Ratios : chronique introuvable"}},
		{Token: testCtx.Admin.Token, PtID: "5", Status: http.StatusOK,
			Sent:         []byte(`{"PaymentRatio":[{"ratio":0.05,"index":0},{"ratio":0.1,"index":1},{"ratio":0.15,"index":2},{"ratio":0.25,"index":3},{"ratio":0.45,"index":4}]}`),
			BodyContains: []string{"PaymentRatio", `"ratio":0.05,"index":0`}, Count: 5},
	}
	for _, tc := range testCases {
		response := e.POST("/api/payment_types/"+tc.PtID+"/payment_ratios").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("PaymentRatio").Array().Length().Equal(tc.Count)
		}
	}
}

// deletePtRatiosTest check route is protected and ratios correctly deleted
func deletePtRatiosTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		PtID         string
		Status       int
		BodyContains []string
	}{
		{Token: testCtx.User.Token, PtID: "0", Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, PtID: "0", Status: http.StatusBadRequest, BodyContains: []string{"Suppression de ratios : introuvable"}},
		{Token: testCtx.Admin.Token, PtID: "5", Status: http.StatusOK, BodyContains: []string{"Ratios supprimés"}},
	}
	for _, tc := range testCases {
		response := e.DELETE("/api/payment_types/"+tc.PtID+"/payment_ratios").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			resp := e.GET("/api/payment_types/"+tc.PtID+"/payment_ratios").
				WithHeader("Authorization", "Bearer "+tc.Token).Expect()
			resp.JSON().Object().Value("PaymentRatio").Array().Length().Equal(0)
		}
	}
}

// getYearRatiosTest check route is protected and ratios correctly calculated
func getYearRatiosTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Year         string
		Status       int
		BodyContains []string
		Count        int
	}{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusBadRequest, BodyContains: []string{"Ratios annuels : année manquante"}, Count: 8},
		{Token: testCtx.User.Token, Status: http.StatusOK, Year: "2011", BodyContains: []string{"Ratios", `"ratio":0.108592`}, Count: 8},
	}
	for _, tc := range testCases {
		response := e.GET("/api/payment_ratios/year").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("Year", tc.Year).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("Ratios").Array().Length().Equal(tc.Count)
		}
	}
}
