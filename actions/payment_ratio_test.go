package actions

import (
	"net/http"
	"strings"
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
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"PaymentRatio"}, ArraySize: 26},
	}
	for i, tc := range testCases {
		response := e.GET("/api/payment_ratios").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetRatios[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetRatios[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetRatios[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// getPtRatiosTest check route is protected and ratios correctly sent.
func getPtRatiosTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			ID: "0", BodyContains: []string{`"PaymentRatio":null`}},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			ID: "5", BodyContains: []string{"PaymentRatio"}, ArraySize: 8},
	}
	for i, tc := range testCases {
		response := e.GET("/api/payment_types/"+tc.ID+"/payment_ratios").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetPtRatios[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetPtRatios[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetPtRatios[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// setPtRatiosTest check route is protected and ratios correctly set.
func setPtRatiosTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			Sent: []byte(`{"PaymentRatio":[{"ratio":0.05,"index":0},
		{"ratio":0.1,"index":1},{"ratio":0.15,"index":2},{"ratio":0.25,"index":3},
		{"ratio":0.45,"index":4}]}`),
			BodyContains: []string{"Ratios d'une chronique, requête : pq"}},
		{Token: testCtx.Admin.Token, ID: "5", Status: http.StatusOK,
			Sent: []byte(`{"PaymentRatio":[{"ratio":0.05,"index":0},
			{"ratio":0.1,"index":1},{"ratio":0.15,"index":2},{"ratio":0.25,"index":3},
			{"ratio":0.45,"index":4}]}`),
			BodyContains: []string{"PaymentRatio", `"ratio":0.05,"index":0`}, ArraySize: 5},
	}
	for i, tc := range testCases {
		response := e.POST("/api/payment_types/"+tc.ID+"/payment_ratios").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nSetPtRatios[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nSetPtRatios[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nSetPtRatios[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// deletePtRatiosTest check route is protected and ratios correctly deleted
func deletePtRatiosTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Suppression des ratios d'une chronique, requête : Ratios de paiement introuvables"}},
		{Token: testCtx.Admin.Token, ID: "5", Status: http.StatusOK,
			BodyContains: []string{"Ratios supprimés"}},
	}
	for i, tc := range testCases {
		response := e.DELETE("/api/payment_types/"+tc.ID+"/payment_ratios").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nDeletePtRatios[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nDeletePtRatios[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.Status == http.StatusOK {
			resp := e.GET("/api/payment_types/"+tc.ID+"/payment_ratios").
				WithHeader("Authorization", "Bearer "+tc.Token).Expect()
			content = string(resp.Content)
			expected := `"PaymentRatio":null`
			if !strings.Contains(content, expected) {
				t.Errorf("\nDeletePtRatios[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, expected, content)
			}
		}
	}
}

// getYearRatiosTest check route is protected and ratios correctly calculated
func getYearRatiosTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusBadRequest,
			BodyContains: []string{"Ratios annuels : année manquante"}},
		{Token: testCtx.User.Token, Status: http.StatusOK, Param: "2011",
			BodyContains: []string{"Ratios", `"ratio":0.108592`}, ArraySize: 8},
	}
	for i, tc := range testCases {
		response := e.GET("/api/payment_ratios/year").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("Year", tc.Param).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetYearRatios[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetYearRatios[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"index"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetYearRatios[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}
