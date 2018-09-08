package actions

import (
	"net/http"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func TestPendingCommitment(t *testing.T) {
	TestCommons(t)
	t.Run("PendingCommitment", func(t *testing.T) {
		getPendingCommitmentsTest(testCtx.E, t)
		getUnlinkedPendingCommitmentsTest(testCtx.E, t)
		getLinkedPendingCommitmentsTest(testCtx.E, t)
		linkPcToOpTest(testCtx.E, t)
		unlinkPCsTest(testCtx.E, t)
		batchPendingsTest(testCtx.E, t)
	})
}

// getPendingCommitmentsTest check route is protected and pending commitments correctly sent.
func getPendingCommitmentsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"PendingCommitments"}, ArraySize: 51},
	}
	for i, tc := range testCases {
		response := e.GET("/api/pending_commitments").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("GetPendings[%d] : contenu incorrect, attendu \"%s\" et reçu \"%s\"", i, tc.BodyContains, content)
			}
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("PendingCommitments").Array().Length().Equal(tc.ArraySize)
		}
	}
}

// getUnlinkedPendingCommitmentsTest check route is protected and pending commitments correctly sent.
func getUnlinkedPendingCommitmentsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{"PendingCommitments"}, ArraySize: 16},
	}
	for i, tc := range testCases {
		response := e.GET("/api/pending_commitments/unlinked").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("GetPendings[%d] : contenu incorrect, attendu \"%s\" et reçu \"%s\"", i, tc.BodyContains, content)
			}
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("PendingCommitments").Array().Length().Equal(tc.ArraySize)
		}
	}
}

// getLinkedPendingCommitmentsTest check route is protected and pending commitments correctly sent.
func getLinkedPendingCommitmentsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{"PendingCommitments"}, ArraySize: 35},
	}
	for i, tc := range testCases {
		response := e.GET("/api/pending_commitments/linked").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("GetPendings[%d] : contenu incorrect, attendu \"%s\" et reçu \"%s\"", i, tc.BodyContains, content)
			}
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("PendingCommitments").Array().Length().Equal(tc.ArraySize)
		}
	}
}

// linkPcToOpTest check route is protected and pending commitments correctly sent.
func linkPcToOpTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusBadRequest,
			BodyContains: []string{"Rattachement d'engagement en cours : opération introuvable"}},
		{Token: testCtx.Admin.Token, ID: "12", Status: http.StatusBadRequest,
			Sent: []byte(`{"peIdList":[228, 14, 230, 231]}`), ArraySize: 16,
			BodyContains: []string{"Rattachement d'engagement en cours, identifiant introuvable"}},
		{Token: testCtx.Admin.Token, ID: "12", Status: http.StatusOK,
			Sent:         []byte(`{"peIdList":[228, 229, 230, 231]}`),
			BodyContains: []string{"PendingCommitments"}, ArraySize: 12},
	}

	for _, tc := range testCases {
		response := e.POST("/api/pending_commitments/physical_ops/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("PendingCommitments").Array().Length().Equal(tc.ArraySize)
		}
	}
}

// unlinkPCsTest check route is protected and pending commitments correctly unlinked.
func unlinkPCsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest,
			Sent: []byte(`{"peIdList":[228, 14, 230, 231]}`), ArraySize: 39,
			BodyContains: []string{"Détachement d'engagement en cours, identifiant introuvable"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent:         []byte(`{"peIdList":[228, 229, 230, 231]}`),
			BodyContains: []string{"PendingCommitments"}, ArraySize: 35},
	}

	for _, tc := range testCases {
		response := e.POST("/api/pending_commitments/unlink").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("PendingCommitments").Array().Length().Equal(tc.ArraySize)
		}
	}
}

// batchPendingsTest check route is protected and return successful.
func batchPendingsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusInternalServerError, Sent: []byte(`{Pend}`),
			BodyContains: []string{"Batch d'engagements en cours, erreur de lecture"}},
		//cSpell:disable
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Sent: []byte(`{"PendingCommitment": [
			{"iris_code":"18002306","name":"METRO LIGNE 11 - PROLONGEMENT A ROSNY BOIS PERRIER - CONVENTION DE FINANCEMENT TRAVAUX N°3","proposed_value":7501596200,"chapter":"908  ","action":"481006011 - Métro    ","commission_date":"2018-05-30T00:00:00Z","beneficiary":"RATP REGIE AUTONOME DES TRANSPORTS PARISIENS"},
			{"iris_code":"18002423","name":"VELO - ITINERAIRE CYCLABLE ENTRE LA GARE DE MENNECY ET L'AVENUE DE VILLEROY (91)","proposed_value":12375000,"chapter":"907  ","action":"17800101 - Réseaux verts et équipements cyclables   ","commission_date":"2018-05-30T00:00:00Z","beneficiary":"COMMUNE DE MENNECY"},
			{"iris_code":"18002451","name":"VELO - COMMUNAUTE D'AGGLOMERATION CERGY PONTOISE - PLAN TRIENNAL - ANNEE 1","proposed_value":25685000,"chapter":"907  ","action":"17800101 - Réseaux verts et équipements cyclables   ","commission_date":"2018-05-30T00:00:00Z","beneficiary":"COMMUNAUTE D'AGGLOMERATION CERGY PONTOISE"},
			{"iris_code":"18003295","name":"RESORPTION DES POINTS NOIRS BRUIT DU FERROVIAIRE - PONT METALLIQUE DES CHANTIERS A VERSAILLES - AVENANT N°1 A LA CONVENTION DE FINANCEMENT ETUDES DE PROJET ET TRAVAUX","proposed_value":19868800,"chapter":"907  ","action":"17700301 - Intégration environnementale des infrastructures de transport  ","commission_date":"2018-05-30T00:00:00Z","beneficiary":"RFF SNCF RESEAU"},
			{"iris_code":"18003447","name":"ROUTE - INNOVATION - OUTIL DE COORDINATION DES CHANTIERS (CD94)","proposed_value":29000000,"chapter":"908  ","action":"18100301 - Etudes et expérimentations    ","commission_date":"2018-05-30T00:00:00Z","beneficiary":"DEPARTEMENT DU VAL DE MARNE"},
			{"iris_code":"18003685","name":"PLD DU SYNDICAT DES TRANSPORTS DE MARNE-LA-VALLEE SECTEURS 3 ET 4 (77)","proposed_value":6876250,"chapter":"908  ","action":"18101401 - PDU : PLD et actions territoriales   ","commission_date":"2018-05-30T00:00:00Z","beneficiary":"TRANSPORTS SECTEUR 3 & 4"}]}`),
			BodyContains: []string{"Engagements en cours importés"}},
	}
	//cSpell:enable
	for _, tc := range testCases {
		response := e.POST("/api/pending_commitments").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}
