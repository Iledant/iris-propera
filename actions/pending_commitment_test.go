package actions

import (
	"net/http"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func testPendingCommitment(t *testing.T) {
	t.Run("PendingCommitment", func(t *testing.T) {
		getPendingCommitmentsTest(testCtx.E, t)
		getUnlinkedPendingCommitmentsTest(testCtx.E, t)
		getLinkedPendingCommitmentsTest(testCtx.E, t)
		getOpPendingsTest(testCtx.E, t)
		linkPcToOpTest(testCtx.E, t)
		unlinkPCsTest(testCtx.E, t)
		batchPendingsTest(testCtx.E, t)
	})
}

// getPendingCommitmentsTest check route is protected and pending commitments correctly sent.
func getPendingCommitmentsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:         testCtx.User.Token,
			Status:        http.StatusOK,
			BodyContains:  []string{"PendingCommitments"},
			CountItemName: `"id"`,
			ArraySize:     51},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/pending_commitments").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetPendings") {
		t.Error(r)
	}
}

// getUnlinkedPendingCommitmentsTest check route is protected and pending commitments correctly sent.
func getUnlinkedPendingCommitmentsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:         testCtx.Admin.Token,
			Status:        http.StatusOK,
			BodyContains:  []string{"PendingCommitments"},
			CountItemName: `"id"`,
			ArraySize:     16},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/pending_commitments/unlinked").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetUnlinkedPendings") {
		t.Error(r)
	}
}

// getLinkedPendingCommitmentsTest check route is protected and pending commitments correctly sent.
func getLinkedPendingCommitmentsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:         testCtx.Admin.Token,
			Status:        http.StatusOK,
			BodyContains:  []string{"PendingCommitments", `"op_name"`, `"op_number"`},
			CountItemName: `"id"`,
			ArraySize:     35},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/pending_commitments/linked").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetLinkedPendings") {
		t.Error(r)
	}
}

// getOpPendingsTest check route is protected and pending commitments correctly sent.
func getOpPendingsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase, // 0 : bad token
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusOK,
			BodyContains: []string{"PendingCommitments", `"op_name"`, `"op_number"`,
				"UnlinkedPendingCommitments", "PhysicalOp"},
			CountItemName: `"id`,
			ArraySize:     670,
		}, // 1 : ok
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/pending_commitments/ops").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetOpPendings") {
		t.Error(r)
	}
}

// linkPcToOpTest check route is protected and pending commitments correctly sent.
func linkPcToOpTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"peIdList":[228, 14, 230, 231]}`),
			BodyContains: []string{"Rattachement d'engagement en cours, requête : pq"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           "12",
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"peIdList":[228, 14, 230, 231]}`),
			BodyContains: []string{"Rattachement d'engagement en cours, requête : Opération ou engagements en cours introuvables"}},
		{
			Token:         testCtx.Admin.Token,
			ID:            "12",
			Status:        http.StatusOK,
			Sent:          []byte(`{"peIdList":[228, 229, 230, 231]}`),
			BodyContains:  []string{"PendingCommitments"},
			CountItemName: `"id"`,
			ArraySize:     12},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/pending_commitments/physical_ops/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "LinkPcToOp") {
		t.Error(r)
	}
}

// unlinkPCsTest check route is protected and pending commitments correctly unlinked.
func unlinkPCsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{"peIdList":[228, 14, 230, 231]}`),
			BodyContains: []string{"Détachement d'engagement en cours, requête : Opération ou engagements en cours introuvables"}},
		{
			Token:         testCtx.Admin.Token,
			Status:        http.StatusOK,
			Sent:          []byte(`{"peIdList":[228, 229, 230, 231]}`),
			BodyContains:  []string{"PendingCommitments"},
			CountItemName: `"id"`,
			ArraySize:     35},
	}

	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/pending_commitments/unlink").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "UnlinkPcs") {
		t.Error(r)
	}
}

// batchPendingsTest check route is protected and return successful.
func batchPendingsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:        testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			Sent:         []byte(`{Pend}`),
			BodyContains: []string{"Batch d'engagements en cours, décodage :"}},
		{
			Token:  testCtx.Admin.Token,
			Status: http.StatusOK,
			//cSpell:disable
			Sent: []byte(`{"PendingCommitment": [
			{"iris_code":"18002306","name":"METRO LIGNE 11 - PROLONGEMENT A ROSNY BOIS PERRIER - CONVENTION DE FINANCEMENT TRAVAUX N°3","proposed_value":7501596200,"chapter":"908  ","action":"481006011 - Métro    ","commission_date":43250,"beneficiary":"RATP REGIE AUTONOME DES TRANSPORTS PARISIENS"},
			{"iris_code":"18002423","name":"VELO - ITINERAIRE CYCLABLE ENTRE LA GARE DE MENNECY ET L'AVENUE DE VILLEROY (91)","proposed_value":12375000,"chapter":"907  ","action":"17800101 - Réseaux verts et équipements cyclables   ","commission_date":43250,"beneficiary":"COMMUNE DE MENNECY"},
			{"iris_code":"18002451","name":"VELO - COMMUNAUTE D'AGGLOMERATION CERGY PONTOISE - PLAN TRIENNAL - ANNEE 1","proposed_value":25685000,"chapter":"907  ","action":"17800101 - Réseaux verts et équipements cyclables   ","commission_date":43250,"beneficiary":"COMMUNAUTE D'AGGLOMERATION CERGY PONTOISE"},
			{"iris_code":"18003295","name":"RESORPTION DES POINTS NOIRS BRUIT DU FERROVIAIRE - PONT METALLIQUE DES CHANTIERS A VERSAILLES - AVENANT N°1 A LA CONVENTION DE FINANCEMENT ETUDES DE PROJET ET TRAVAUX","proposed_value":19868800,"chapter":"907  ","action":"17700301 - Intégration environnementale des infrastructures de transport  ","commission_date":43250,"beneficiary":"RFF SNCF RESEAU"},
			{"iris_code":"18003447","name":"ROUTE - INNOVATION - OUTIL DE COORDINATION DES CHANTIERS (CD94)","proposed_value":29000000,"chapter":"908  ","action":"18100301 - Etudes et expérimentations    ","commission_date":43250,"beneficiary":"DEPARTEMENT DU VAL DE MARNE"},
			{"iris_code":"18003685","name":"PLD DU SYNDICAT DES TRANSPORTS DE MARNE-LA-VALLEE SECTEURS 3 ET 4 (77)","proposed_value":6876250,"chapter":"908  ","action":"18101401 - PDU : PLD et actions territoriales   ","commission_date":43250,"beneficiary":"TRANSPORTS SECTEUR 3 & 4"}]}`),
			//cSpell:enable
			BodyContains: []string{"Engagements en cours importés"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/pending_commitments").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "BatchPendings") {
		t.Error(r)
	}
}
