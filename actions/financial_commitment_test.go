package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestFinancialCommitment embeddes all tests for financial commitment insuring the configuration and DB are properly initialized.
func TestFinancialCommitment(t *testing.T) {
	TestCommons(t)
	t.Run("FinancialCommitment", func(t *testing.T) {
		getUnlinkedFcsTest(testCtx.E, t)
		getMonthFCTest(testCtx.E, t)
		getLinkedFcsTest(testCtx.E, t)
		getOpFcsTest(testCtx.E, t)
		unlinkFcsTest(testCtx.E, t)
		linkFcToOpTest(testCtx.E, t)
		linkFcToPlTest(testCtx.E, t)
		batchFcsTest(testCtx.E, t)
		batchOpFcsTest(testCtx.E, t)
	})
}

// getUnlinkedFcsTest tests route is protected and all financial commitments are sent back.
func getUnlinkedFcsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Page         int
		Search       string
		LinkType     string
		MinYear      int
		Status       int
		BodyContains []string
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Page: 1, LinkType: "PhysicalOp", Search: "", MinYear: 0,
			BodyContains: []string{"FinancialCommitment", `"last_page":1`}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Page: 1, LinkType: "PlanLine", Search: "", MinYear: 0,
			BodyContains: []string{"FinancialCommitment", `"last_page":268`}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Page: 1, LinkType: "PlanLine", Search: "", MinYear: 2018,
			BodyContains: []string{"FinancialCommitment", `"last_page":4`}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Page: 1, LinkType: "PlanLine", Search: "", MinYear: 2018,
			BodyContains: []string{"FinancialCommitment", `"last_page":4`}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Page: 10, LinkType: "PlanLine", Search: "RATP", MinYear: 2010,
			BodyContains: []string{"FinancialCommitment", `"last_page":8`, `"current_page":8`}},
	}

	for _, tc := range testCases {
		response := e.GET("/api/financial_commitments").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("page", tc.Page).WithQuery("LinkType", tc.LinkType).WithQuery("search", tc.Search).
			WithQuery("MinYear", tc.MinYear).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}

// getMonthFCTest tests route is protected and all financial commitments are sent back.
func getMonthFCTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Year         int
		Status       int
		BodyContains []string
	}{
		{Token: "", Status: http.StatusInternalServerError, BodyContains: []string{"Token absent"}},
		{Token: testCtx.User.Token, Status: http.StatusOK, Year: 2017,
			BodyContains: []string{"FinancialCommitmentPerMonth", `"month":3`, `"value":9681497875`}},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"FinancialCommitmentPerMonth", `"month":3`, `"value":10778560491`}},
	}

	for _, tc := range testCases {
		response := e.GET("/api/financial_commitments/month").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("year", tc.Year).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}

// getLinkedFcsTest tests route is protected and all financial commitments are sent back.
func getLinkedFcsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Page         int
		Search       string
		LinkType     string
		MinYear      int
		Status       int
		BodyContains []string
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Page: 1, LinkType: "PhysicalOp", Search: "", MinYear: 0,
			BodyContains: []string{"FinancialCommitment", `"last_page":285`}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Page: 1, LinkType: "PhysicalOp", Search: "", MinYear: 2016,
			BodyContains: []string{"FinancialCommitment", `"last_page":28`}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Page: 50, LinkType: "PhysicalOp", Search: "SNCF", MinYear: 0,
			BodyContains: []string{"FinancialCommitment", `"last_page":39`, `"current_page":39`}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Page: 1, LinkType: "PlanLine", Search: "", MinYear: 0,
			BodyContains: []string{"FinancialCommitment", `"last_page":17`}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Page: 1, LinkType: "PlanLine", Search: "", MinYear: 2016,
			BodyContains: []string{"FinancialCommitment", `"last_page":15`}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Page: 50, LinkType: "PlanLine", Search: "SNCF", MinYear: 0,
			BodyContains: []string{"FinancialCommitment", `"last_page":6`, `"current_page":6`}},
	}

	for _, tc := range testCases {
		response := e.GET("/api/financial_commitments/linked").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("page", tc.Page).WithQuery("LinkType", tc.LinkType).WithQuery("search", tc.Search).
			WithQuery("MinYear", tc.MinYear).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}

// getOpFcsTest check route is protected and number of financial commitments sent back is good.
func getOpFcsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		OpID         string
		Status       int
		BodyContains []string
		FcsCount     int
	}{
		{Token: "fake", OpID: "0", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusBadRequest,
			OpID: "0", BodyContains: []string{"Liste des engagements : opération introuvable"}},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			OpID: "12", BodyContains: []string{"FinancialCommitment"}, FcsCount: 8},
	}
	for _, tc := range testCases {
		response := e.GET("/api/physical_ops/"+tc.OpID+"/financial_commitments").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("FinancialCommitment").Array().Length().Equal(tc.FcsCount)
		}
	}
}

// unlinkFcs check if route is protected and links are correctly removed.
func unlinkFcsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Sent         []byte
		Status       int
		BodyContains []string
		OpID         string
		FcsCount     int
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent:         []byte(`{"linkType":"PhysicalOp","fcIdList":[2036, 2052, 2053, 3618, 2082]}`),
			BodyContains: []string{"FinancialCommitment", `"last_page":284`}, OpID: "12", FcsCount: 3},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest,
			Sent:         []byte(`{"linkType":"PhysicalOp","fcIdList":[0]}`),
			BodyContains: []string{"Détachement d'engagement : mauvais identificateur d'engagement"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent:         []byte(`{"linkType":"PlanLine","fcIdList":[138,147,190,136,192]}`),
			BodyContains: []string{"FinancialCommitment", `"last_page":17`}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest,
			Sent:         []byte(`{"linkType":"PlanLine","fcIdList":[0]}`),
			BodyContains: []string{"Détachement d'engagement : mauvais identificateur d'engagement"}},
	}

	for _, tc := range testCases {
		response := e.POST("/api/financial_commitments/unlink").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
		if tc.OpID != "" {
			response := e.GET("/api/physical_ops/"+tc.OpID+"/financial_commitments").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
			response.JSON().Object().Value("FinancialCommitment").Array().Length().Equal(tc.FcsCount)
		}
		// TODO implement plan line get test
	}
}

// linkFcToOpTest check if route is protected and links are correctly done.
func linkFcToOpTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Sent         []byte
		Status       int
		BodyContains []string
		OpID         string
		FcsCount     int
	}{
		{Token: testCtx.User.Token, OpID: "0", Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, OpID: "0", Status: http.StatusBadRequest, BodyContains: []string{"Rattachement d'engagement : opération introuvable"}},
		{Token: testCtx.Admin.Token, OpID: "12", Status: http.StatusOK,
			Sent:         []byte(`{"fcIdList":[2036, 2052, 2053, 3618, 2082]}`),
			BodyContains: []string{"FinancialCommitment", `"last_page":1`}, FcsCount: 8},
	}

	for _, tc := range testCases {
		response := e.POST("/api/financial_commitments/physical_ops/"+tc.OpID).WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
		if tc.OpID != "0" {
			response := e.GET("/api/physical_ops/"+tc.OpID+"/financial_commitments").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
			response.JSON().Object().Value("FinancialCommitment").Array().Length().Equal(tc.FcsCount)
		}
	}
}

// linkFcToPlTest check if route is protected and links are correctly done.
func linkFcToPlTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Sent         []byte
		Status       int
		BodyContains []string
		PlID         string
		FcsCount     int
	}{
		{Token: testCtx.User.Token, PlID: "0", Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, PlID: "0", Status: http.StatusBadRequest, BodyContains: []string{"Rattachement d'engagement : ligne de plan introuvable"}},
		{Token: testCtx.Admin.Token, PlID: "23", Status: http.StatusOK,
			Sent:         []byte(`{"fcIdList":[138,147,190,136,192]}`),
			BodyContains: []string{"FinancialCommitment", `"last_page":268`}, FcsCount: 8},
	}

	for _, tc := range testCases {
		response := e.POST("/api/financial_commitments/plan_lines/"+tc.PlID).WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
		// TODO implement get plan line FCs test
		// if tc.PlID != "0" {
		// 	response := e.GET("/api/physical_ops/"+tc.PlID+"/financial_commitments").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		// 	response.JSON().Object().Value("FinancialCommitment").Array().Length().Equal(tc.FcsCount)
		// }
	}
}

// batchFcsTest check if route is protected and no error encounters when pattern is good.
func batchFcsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Sent         []byte
		Status       int
		BodyContains []string
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusInternalServerError, BodyContains: []string{"JSON"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, BodyContains: []string{"Engagements importés et mis à jour"},
			//cSpell:disable
			Sent: []byte(`{"FinancialCommitment":[
				{"chapter":"907","action":"17700301 - Intégration environnementale des infrastructures de transport","iris_code":"18002439","coriolis_year":"2018","coriolis_egt_code":"IRIS","coriolis_egt_num":"553827","coriolis_egt_line":"1","name":"ROUTE - INNOVATION INFRASTRUCTURE ROUTIERE - VAL D'OISE","beneficiary":"DEPARTEMENT DU VAL D'OISE","beneficiary_code":2306,"date":"2018-03-16T00:00:00Z","value":3000000,"lapse_date":"2021-03-16T00:00:00Z"},
				{"chapter":"907","action":"17700301 - Intégration environnementale des infrastructures de transport","iris_code":"18003295","coriolis_year":"2018","coriolis_egt_code":"IRIS","coriolis_egt_num":"557246","coriolis_egt_line":"1","name":"RESORPTION DES POINTS NOIRS BRUIT DU FERROVIAIRE - PONT METALLIQUE DES CHANTIERS A VERSAILLES - AVENANT N°1 A LA CONVENTION DE FINANCEMENT ETUDES DE PROJET ET TRAVAUX","beneficiary":"RFF SNCF RESEAU","beneficiary_code":14154,"date":"2018-05-30T00:00:00Z","value":198688,"lapse_date":"2021-05-30T00:00:00Z"}]}`)},
		//cSpell:enable
	}

	for _, tc := range testCases {
		response := e.POST("/api/financial_commitments").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}

// batchOpFcsTest check if route is protected and no error encounters when pattern is good.
func batchOpFcsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Sent         []byte
		Status       int
		BodyContains []string
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusInternalServerError, BodyContains: []string{"JSON"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, BodyContains: []string{"Rattachements importés et réalisés"},
			Sent: []byte(`{"Attachment":[{"op_number":"18FF005","coriolis_year":"2007","coriolis_egt_code":"UAD","coriolis_egt_num":"217075","coriolis_egt_line":"1"},
				{"op_number":"18FF005","coriolis_year":"2007","coriolis_egt_code":"UAD","coriolis_egt_num":"217078","coriolis_egt_line":"1"},
				{"op_number":"18FF005","coriolis_year":"2008","coriolis_egt_code":"P1215","coriolis_egt_num":"241790","coriolis_egt_line":"1"},
				{"op_number":"18FF005","coriolis_year":"2008","coriolis_egt_code":"P1215","coriolis_egt_num":"241792","coriolis_egt_line":"1"}]}`)},
	}
	for _, tc := range testCases {
		response := e.POST("/api/financial_commitments/attachments").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		if tc.Status == http.StatusOK {
			body := e.GET("/api/physical_ops/17/financial_commitments").WithHeader("Authorization", "Bearer "+tc.Token).Expect().Body()
			body.Contains("R-2007-UAD-217075-1")
			body.Contains("R-2007-UAD-217078-1")
			body.Contains("R-2008-P1215-241790-1")
			body.Contains("R-2008-P1215-241792-1")
		}
		response.Status(tc.Status)
	}
}
