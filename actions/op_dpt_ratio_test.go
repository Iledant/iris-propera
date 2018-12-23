package actions

import (
	"net/http"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func TestOpDptRatio(t *testing.T) {
	TestCommons(t)
	t.Run("OpDptRatio", func(t *testing.T) {
		getOpWithDptRatiosTest(testCtx.E, t)
		batchOpDptRatiosTest(testCtx.E, t)
		getFCPerDptTest(testCtx.E, t)
		getDetailedFCPerDptTest(testCtx.E, t)
		getDetailedPrgPerDptTest(testCtx.E, t)
	})
}

// getOpWithDptRatiosTest check route is protected and datas sent has got items and number of lines.
func getOpWithDptRatiosTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{"OpsWithDptRatios", "name", "number", "r75", "r77", "r78", "r91",
				"r92", "r93", "r94", "r95", "ProgrammingsYears"}},
	}
	for i, tc := range testCases {
		response := e.GET("/api/op_dpt_ratios/ops").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetOpWithDptRatios[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetOpWithDptRatios[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// batchOpDptRatiosTest check route is protected and datas sent back are similar to batch.
func batchOpDptRatiosTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent: []byte(`{"OpDptRatios":[{"physical_op_id":9,"r75":0.2,"r77":0.2,"r78":0.2,"r91":0.2,"r92":0.2,"r93":0,"r94":0,"r95":0}]}`),
			BodyContains: []string{"OpsWithDptRatios", "9", `"r75":0.2`, "r77", "r78", "r91",
				"r92", "r93", "r94", "r95", "ProgrammingsYears"}},
	}
	for i, tc := range testCases {
		response := e.POST("/api/op_dpt_ratios/upload").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nBatchOpDptRatios[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nBatchOpDptRatios[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// getFCPerDptTest check route is protected and datas sent has got items and number of lines.
func getFCPerDptTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{`"FinancialCommitmentPerDpt":[{"total":137901736223,"fc75":null,"fc77":null,"fc78":null,"fc91":null,"fc92":null,"fc93":null,"fc94":null,"fc95":null}]`}},
	}
	for i, tc := range testCases {
		response := e.GET("/api/op_dpt_ratios/financial_commitments").WithHeader("Authorization", "Bearer "+tc.Token).WithQuery("firstYear", 2016).WithQuery("lastYear", 2018).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetFCPerDpt[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetFCPerDpt[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// getDetailedFCPerDptTest check route is protected and datas sent has got items and number of lines.
func getDetailedFCPerDptTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{`"DetailedFinancialCommitmentPerDpt":[`,
				//cSpell:disable
				`{"total":1053500000,"fc75":null,"fc77":null,"fc78":null,"fc91":null,"fc92":null,"fc93":null,"fc94":null,"fc95":null,"id":13,"number":"01BU003","name":"Bus - Tzen5 - Paris-Choisy (94)"}`}},
		//cSpell:enable
	}
	for i, tc := range testCases {
		response := e.GET("/api/op_dpt_ratios/detailed_financial_commitments").WithHeader("Authorization", "Bearer "+tc.Token).WithQuery("firstYear", 2016).WithQuery("lastYear", 2018).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetDetailedFCPerDpt[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetDetailedFCPerDpt[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// getDetailedPrgPerDptTest check route is protected and datas sent has got items and number of lines.
func getDetailedPrgPerDptTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{`"DetailedProgrammingsPerDpt":[`,
				`{"date":"2018-03-16T00:00:00Z","id":37,"number":"02VE001","name":"Vélo - Toutes opérations","total":1081455457,"pr75":null,"pr77":null,"pr78":null,"pr91":null,"pr92":null,"pr93":null,"pr94":null,"pr95":null}`}},
	}
	for i, tc := range testCases {
		response := e.GET("/api/op_dpt_ratios/detailed_programmings").WithHeader("Authorization", "Bearer "+tc.Token).WithQuery("year", 2018).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetDetailedPrgPerDpt[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetDetailedPrgPerDpt[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}
