package actions

import (
	"net/http"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestFinancialCommitment embeddes all tests for financial commitment insuring the configuration and DB are properly initialized.
func testFinancialCommitment(t *testing.T) {
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
			BodyContains: []string{`"data":[],"current_page":0,"last_page":1`}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Page: 1, LinkType: "PlanLine", Search: "", MinYear: 0,
			BodyContains: []string{`"data":[{"id":1,"value":6000000,"iris_code":"R-2007-UAD-217075-1","name":"SEINE AVAL","date":"2007-10-11T00:00:00Z","beneficiary":"VNF VOIES NAVIGABLES DE FRANCE"},`,
				`"last_page":268`}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Page: 1, LinkType: "PlanLine", Search: "", MinYear: 2018,
			// cSpell:disable
			BodyContains: []string{`"data":[{"id":4264,"value":56000000,"iris_code":"18002216","name":"LIGNE RER D - REHAUSSEMENT DES QUAIS EN GARE DE VILLENEUVE SAINT-GEORGES EN LIEN AVEC LE DEPLOIEMENT DU RER NG - CONVENTION ETUDES EP/APO","date":"2018-03-16T00:00:00Z","beneficiary":"RFF SNCF RESEAU"},`,
				// cSpell:enable
				`"last_page":4`}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Page: 10, LinkType: "PlanLine", Search: "RATP", MinYear: 2010,
			// cSpell:disable
			BodyContains: []string{`{"data":[{"id":4042,"value":4216500000,"iris_code":"14014745","name":"PROLONGEMENT DE LA LIGNE 14 A MAIRIE DE SAINT-OUEN","date":"2014-11-20T00:00:00Z","beneficiary":"RATP REGIE AUTONOME DES TRANSPORTS PARISIENS"},`,
				// cSpell:enable
				`"last_page":8`, `"current_page":8`}},
	}

	for i, tc := range testCases {
		response := e.GET("/api/financial_commitments").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("page", tc.Page).WithQuery("LinkType", tc.LinkType).WithQuery("search", tc.Search).
			WithQuery("MinYear", tc.MinYear).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetUnlinkedFcs[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetUnlinkedFcs[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
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
			BodyContains: []string{"FinancialCommitmentsPerMonth", `"month":3`, `"value":9681497875`}},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"FinancialCommitmentsPerMonth", `"month":3`, `"value":10778560491`}},
	}

	for i, tc := range testCases {
		response := e.GET("/api/financial_commitments/month").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("year", tc.Year).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetMonthFCTest[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetMonthFCTest[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
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
			BodyContains: []string{`{"data":[{"fcId":1,"fcValue":6000000,"fcName":"SEINE AVAL","iris_code":"R-2007-UAD-217075-1","fcDate":"2007-10-11T00:00:00Z","opNumber":"18VN040","opName":"Hors fret - Aval - Autres ouvrages - Seine (78) (92)","fcBeneficiary":"VNF VOIES NAVIGABLES DE FRANCE"}`,
				`"last_page":285`}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Page: 1, LinkType: "PhysicalOp", Search: "", MinYear: 2016,
			// cSpell:disable
			BodyContains: []string{`{"data":[{"fcId":123,"fcValue":1136100000,"fcName":"SCHEMA DIRECTEUR DU RER B SUD - AMENAGEMENT DES GARES - PRO/REA DE LA  GARE DE LA CROIX DE BERNY","iris_code":"16007501","fcDate":"2016-10-12T00:00:00Z","opNumber":"15RE001","opName":"RER B - sud - Modernisation des gares","fcBeneficiary":"RATP REGIE AUTONOME DES TRANSPORTS PARISIENS"},`,
				// cSpell:enable
				`"last_page":28`}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Page: 50, LinkType: "PhysicalOp", Search: "SNCF", MinYear: 0,
			// cSpell:disable
			BodyContains: []string{`{"data":[{"fcId":4263,"fcValue":3000000000,"fcName":"SCHEMA DIRECTEUR D'ACCESSIBILITE - SNCF RESEAU (AFFECTATION 2017)","iris_code":"17011482","fcDate":"2017-11-23T00:00:00Z","opNumber":"11AC001","opName":"Accessibilité - SDA - tranche 1","fcBeneficiary":"RFF SNCF RESEAU"},`,
				// cSpell:enable
				`"last_page":39`, `"current_page":39`}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Page: 1, LinkType: "PlanLine", Search: "", MinYear: 0,
			// cSpell:disable
			BodyContains: []string{`{"data":[{"fcId":98,"fcValue":78750000,"fcName":"SCHEMA DIRECTEUR RER A - ETUDES D'AVANT-PROJET NIVEAU PROJET DE LA GARE D'AUBER","iris_code":"15014974","fcDate":"2015-10-08T00:00:00Z","plName":"RATP REGIE AUTONOME DES TRANSPORTS PARISIENS","fcBeneficiary":"CPER01 - Amélioration et modernisation des RER (schémas directeurs et gares)"},`,
				// cSpell:enable
				`"last_page":17`}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Page: 1, LinkType: "PlanLine", Search: "", MinYear: 2016,
			// cSpell:disable
			BodyContains: []string{`{"data":[{"fcId":123,"fcValue":1136100000,"fcName":"SCHEMA DIRECTEUR DU RER B SUD - AMENAGEMENT DES GARES - PRO/REA DE LA  GARE DE LA CROIX DE BERNY","iris_code":"16007501","fcDate":"2016-10-12T00:00:00Z","plName":"RATP REGIE AUTONOME DES TRANSPORTS PARISIENS","fcBeneficiary":"CPER01 - Amélioration et modernisation des RER (schémas directeurs et gares)"},`,
				// cSpell:enable
				`"last_page":15`}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Page: 50, LinkType: "PlanLine", Search: "SNCF", MinYear: 0,
			BodyContains: []string{`{"data":[{"fcId":4230,"fcValue":535367700,"fcName":"LIGNE TRANSILIEN J - ADAPTATION DES INFRASTRUCTURES - CONVENTION APO et REA","iris_code":"17012571","fcDate":"2017-10-18T00:00:00Z","plName":"RFF SNCF RESEAU","fcBeneficiary":"CPER02 - Schémas de secteurs Transilien"},`,
				`"last_page":6`, `"current_page":6`}},
	}

	for i, tc := range testCases {
		response := e.GET("/api/financial_commitments/linked").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("page", tc.Page).WithQuery("LinkType", tc.LinkType).WithQuery("search", tc.Search).
			WithQuery("MinYear", tc.MinYear).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetLinkedFcs[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetLinkedFcs[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
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
		{Token: testCtx.User.Token, Status: http.StatusOK,
			OpID: "0", BodyContains: []string{`"FinancialCommitment":[]`}},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			OpID: "12", BodyContains: []string{"FinancialCommitment"}, FcsCount: 8},
	}
	for i, tc := range testCases {
		response := e.GET("/api/physical_ops/"+tc.OpID+"/financial_commitments").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetOpFcs[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetOpFcs[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.FcsCount > 0 {
			count := strings.Count(content, `"coriolis_year"`)
			if count != tc.FcsCount {
				t.Errorf("\nGetOpFcs[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.FcsCount, count)
			}
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
			Sent: []byte(`{"linkType":"PhysicalOp","fcIdList":[2036, 2052, 2053, 3618, 2082]}`),
			BodyContains: []string{`{"data":[{"fcId":1,"fcValue":6000000,"fcName":"SEINE AVAL","iris_code":"R-2007-UAD-217075-1","fcDate":"2007-10-11T00:00:00Z","opNumber":"18VN040","opName":"Hors fret - Aval - Autres ouvrages - Seine (78) (92)","fcBeneficiary":"VNF VOIES NAVIGABLES DE FRANCE"},`,
				`"last_page":284`}, OpID: "12", FcsCount: 3},
		{Token: testCtx.Admin.Token, Status: http.StatusInternalServerError,
			Sent:         []byte(`{"linkType":"PhysicalOp","fcIdList":[0]}`),
			BodyContains: []string{"Détachement d'engagements, requête : Engagements incorrects"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent: []byte(`{"linkType":"PlanLine","fcIdList":[138,147,190,136,192]}`),
			// cSpell:disable
			BodyContains: []string{`{"data":[{"fcId":98,"fcValue":78750000,"fcName":"SCHEMA DIRECTEUR RER A - ETUDES D'AVANT-PROJET NIVEAU PROJET DE LA GARE D'AUBER","iris_code":"15014974","fcDate":"2015-10-08T00:00:00Z","plName":"RATP REGIE AUTONOME DES TRANSPORTS PARISIENS","fcBeneficiary":"CPER01 - Amélioration et modernisation des RER (schémas directeurs et gares)"},`,
				// cSpell:enable
				`"last_page":17`}},
		{Token: testCtx.Admin.Token, Status: http.StatusInternalServerError,
			Sent:         []byte(`{"linkType":"PlanLine","fcIdList":[0]}`),
			BodyContains: []string{"Détachement d'engagements, requête : Engagements incorrects"}},
	}

	for i, tc := range testCases {
		response := e.POST("/api/financial_commitments/unlink").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()

		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nUnlinkFcs[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nUnlinkFcs[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.OpID != "" {
			response := e.GET("/api/physical_ops/"+tc.OpID+"/financial_commitments").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
			count := strings.Count(string(response.Content), `"id"`)
			if count != tc.FcsCount {
				t.Errorf("\nUnlinkFcs[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.FcsCount, count)
			}
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
		{Token: testCtx.Admin.Token, OpID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Rattachement engagements / opération, requête : pq"},
			Sent:         []byte(`{"fcIdList":[2036, 2052, 2053, 3618, 2082]}`),
		},
		{Token: testCtx.Admin.Token, OpID: "12", Status: http.StatusOK,
			Sent:         []byte(`{"fcIdList":[2036, 2052, 2053, 3618, 2082]}`),
			BodyContains: []string{`{"data":[],"current_page":0,"last_page":1}`}, FcsCount: 8},
	}

	for i, tc := range testCases {
		response := e.POST("/api/financial_commitments/physical_ops/"+tc.OpID).WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()

		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nLinkFcToOp[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nLinkFcToOp[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.OpID != "0" {
			response := e.GET("/api/physical_ops/"+tc.OpID+"/financial_commitments").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
			count := strings.Count(string(response.Content), `"id"`)
			if count != tc.FcsCount {
				t.Errorf("\nUnlinkFcs[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.FcsCount, count)
			}
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
		{Token: testCtx.User.Token, PlID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, PlID: "0", Status: http.StatusInternalServerError,
			Sent:         []byte(`{"fcIdList":[138,147,190,136,192]}`),
			BodyContains: []string{"Rattachement engagements / ligne de plan, requête : pq"}},
		{Token: testCtx.Admin.Token, PlID: "23", Status: http.StatusOK,
			Sent: []byte(`{"fcIdList":[138,147,190,136,192]}`),
			BodyContains: []string{`{"data":[{"id":1,"value":6000000,"iris_code":"R-2007-UAD-217075-1","name":"SEINE AVAL","date":"2007-10-11T00:00:00Z","beneficiary":"VNF VOIES NAVIGABLES DE FRANCE"},`,
				`"last_page":268`}, FcsCount: 8},
	}

	for i, tc := range testCases {
		response := e.POST("/api/financial_commitments/plan_lines/"+tc.PlID).WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nLinkFcToPl[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nLinkFcToPl[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
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
				{"chapter":"907","action":"17700301 - Intégration environnementale des infrastructures de transport","iris_code":"18002439","coriolis_year":"2018","coriolis_egt_code":"IRIS","coriolis_egt_num":"553827","coriolis_egt_line":"1","name":"ROUTE - INNOVATION INFRASTRUCTURE ROUTIERE - VAL D'OISE","beneficiary":"DEPARTEMENT DU VAL D'OISE","beneficiary_code":2306,"date":43175,"value":3000000,"lapse_date":44271},
				{"chapter":"907","action":"17700301 - Intégration environnementale des infrastructures de transport","iris_code":"18003295","coriolis_year":"2018","coriolis_egt_code":"IRIS","coriolis_egt_num":"557246","coriolis_egt_line":"1","name":"RESORPTION DES POINTS NOIRS BRUIT DU FERROVIAIRE - PONT METALLIQUE DES CHANTIERS A VERSAILLES - AVENANT N°1 A LA CONVENTION DE FINANCEMENT ETUDES DE PROJET ET TRAVAUX","beneficiary":"RFF SNCF RESEAU","beneficiary_code":14154,"date":43250,"value":198688,"lapse_date":44346}]}`)},
		//cSpell:enable
	}

	for i, tc := range testCases {
		response := e.POST("/api/financial_commitments").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nBatchFcs[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nBatchFcs[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
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
	for i, tc := range testCases {
		response := e.POST("/api/financial_commitments/attachments").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nBatchOpFcs[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nBatchOpFcs[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.Status == http.StatusOK {
			body := e.GET("/api/physical_ops/17/financial_commitments").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
			content = string(body.Content)
			for _, s := range []string{"R-2007-UAD-217075-1", "R-2007-UAD-217078-1",
				"R-2008-P1215-241790-1", "R-2008-P1215-241792-1"} {
				if !strings.Contains(content, s) {
					t.Errorf("\nBatchOpFcs[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
				}
			}
		}
		response.Status(tc.Status)
	}
}
