package actions

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testPaymentNeed(t *testing.T) {
	t.Run("PaymentNeed", func(t *testing.T) {
		pnID := createPaymentNeedTest(testCtx.E, t)
		if pnID == 0 {
			t.Errorf("Impossible de créer le besoin de paiement")
			t.FailNow()
		}
		modifyPaymentNeedTest(testCtx.E, t, pnID)
		getPaymentNeedsTest(testCtx.E, t)
		deletePaymentNeedTest(testCtx.E, t, pnID)
	})
}

// createPaymentNeedTest tests route is protected and sent PaymentNeed is created.
func createPaymentNeedTest(e *httpexpect.Expect, t *testing.T) (pnID int) {
	testCases := []testCase{
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"},
		}, // 0 : user unauthorized
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{`),
			BodyContains: []string{"Création d'un besoin de paiement, décodage : "},
		}, // 1 : bad json
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{"PaymentNeed":{"Date":"2019-10-25T20:00:00Z","Value":100000,"Comment":null}}`),
			BodyContains: []string{`Création d'un besoin de paiement, requête : beneficiary ID nul`},
		}, // 2 : beneficiary ID nul
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{"PaymentNeed":{"BeneficiaryID":8,"Date":"2019-10-25T20:00:00Z","Comment":null}}`),
			BodyContains: []string{`Création d'un besoin de paiement, requête : value nul`},
		}, // 3 : value null
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusCreated,
			Sent:         []byte(`{"PaymentNeed":{"BeneficiaryID":8,"Date":"2019-10-25T20:00:00Z","Value":100000,"Comment":null}}`),
			BodyContains: []string{"PaymentNeed", `RATP`, `"Date":"2019-10-25T20:00:00Z"`, `"Value":100000`, `"Comment":""`},
		}, // 4 : ok
	}
	for i, tc := range testCases {
		response := e.POST("/api/payment_need").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nCreatePaymentNeed[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nCreatePaymentNeed[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.Status == http.StatusCreated {
			pnID = int(response.JSON().Object().Value("PaymentNeed").Object().Value("ID").Number().Raw())
		}
	}
	return pnID
}

// modifyPaymentNeedTest tests route is protected and modify work properly.
func modifyPaymentNeedTest(e *httpexpect.Expect, t *testing.T, pnID int) {
	testCases := []testCase{
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"},
		}, // 0 unauthorized
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{`),
			BodyContains: []string{`Modification d'un besoin de paiement, décodage :`},
		}, // 1 bad json
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusInternalServerError,
			Sent: []byte(`{"PaymentNeed":{"ID":` + strconv.Itoa(pnID) +
				`,"BeneficiaryID":10,"Date":"2019-10-23T20:00:00Z","Comment":"commentaire"}}`),
			BodyContains: []string{`Modification d'un besoin de paiement, requête : value nul`},
		}, // 2 bad value
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusInternalServerError,
			Sent: []byte(`{"PaymentNeed":{"ID":` + strconv.Itoa(pnID) +
				`,"Date":"2019-10-23T20:00:00Z","Value":30000,"Comment":"commentaire"}}`),
			BodyContains: []string{`Modification d'un besoin de paiement, requête : beneficiary ID nul`},
		}, // 3 bad beneficiary ID
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"PaymentNeed":{"ID":0,"BeneficiaryID":10,"Date":"2019-10-23T20:00:00Z","Value":30000,"Comment":"commentaire"}}`),
			BodyContains: []string{`Modification d'un besoin de paiement, requête :`},
		}, // 4 bad ID
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusOK,
			Sent: []byte(`{"PaymentNeed":{"ID":` + strconv.Itoa(pnID) +
				`,"BeneficiaryID":10,"Date":"2019-10-23T20:00:00Z","Value":30000,"Comment":"commentaire"}}`),
			BodyContains: []string{"PaymentNeed", `STIF`, `"BeneficiaryID":10`,
				`"Date":"2019-10-23`, `"Value":30000`, `"Comment":"commentaire"`},
		}, // 5 ok
	}
	for i, tc := range testCases {
		response := e.PUT("/api/payment_need").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nModifyPaymentNeed[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nModifyPaymentNeed[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// getPaymentNeedsTest tests route is protected and all PaymentNeeds are sent back.
func getPaymentNeedsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{
			Token:        "fake",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"},
		},
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusOK,
			Param:        "Year=2019&PaymentTypeID=4",
			BodyContains: []string{`"PaymentNeed":[`},
			ArraySize:    2,
		},
	}

	for i, tc := range testCases {
		response := e.GET("/api/payment_need").
			WithHeader("Authorization", "Bearer "+tc.Token).WithQueryString(tc.Param).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetPaymentNeeds[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"",
					i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetPaymentNeeds[%d],statut :  attendu ->%v  reçu <-%v", i,
				tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"Need"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetPaymentNeeds[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d",
					i, tc.ArraySize, count)
			}
		}
	}
}

// deletePaymentNeedTest tests route is protected and delete work properly.
func deletePaymentNeedTest(e *httpexpect.Expect, t *testing.T, pnID int) {
	testCases := []testCase{
		{
			Token:        testCtx.User.Token,
			ID:           "0",
			Status:       http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"},
		}, // 0 unauthorized
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Suppression d'un besoin de paiement, requête : payment need introuvable"},
		}, // 1 bad ID
		{
			Token:        testCtx.Admin.Token,
			ID:           strconv.Itoa(pnID),
			Status:       http.StatusOK,
			BodyContains: []string{"Besoin de paiement supprimé"},
		}, // 2 ok
	}
	for i, tc := range testCases {
		response := e.DELETE("/api/payment_need/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nDeletePaymentNeed[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nDeletePaymentNeed[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}
