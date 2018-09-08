package actions

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

//TestPhysicalOps includes all tests for physical operation handler.
func TestPhysicalOps(t *testing.T) {
	TestCommons(t)
	t.Run("PhysicalOps", func(t *testing.T) {
		getPhysicalOpsTest(testCtx.E, t)
		ID := createPhysicalOpTest(testCtx.E, t)
		updatePhysicalOpTest(testCtx.E, t)
		deletePhysicalOpTest(testCtx.E, t, ID)
		batchPhysicalOpsTest(testCtx.E, t)
		getPrevisionsTests(testCtx.E, t)
		setOpPrevisionsTests(testCtx.E, t)
	})
}

// getPhysicalOpsTest tests if route is protected and returned list properly formatted.
func getPhysicalOpsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token absent"}, ArraySize: 0},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"PhysicalOp"}, ArraySize: 3},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{"PhysicalOp"}, ArraySize: 619},
	}

	for i, tc := range testCases {
		response := e.GET("/api/physical_ops").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("GetPhysicalOps[%d] : contenu incorrect, attendu \"%s\" et reçu\n\"%s\"", i, tc.BodyContains, content)
			}
		}
		if tc.Status == http.StatusOK {
			response.JSON().Object().ContainsKey("PhysicalOp")
			response.JSON().Object().Value("PhysicalOp").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

//createPhysicalOpTest tests if route is protected, validations ok and number correctly computed.
func createPhysicalOpTest(e *httpexpect.Expect, t *testing.T) (ID int) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest, Sent: []byte(`{}`),
			BodyContains: []string{"Mauvais format de numéro d'opération"}},
		{Token: testCtx.Admin.Token, Sent: []byte(`{"number":"99XX001","name":""}`),
			Status: http.StatusBadRequest, BodyContains: []string{"Nom de l'opération absent"}},
		{Token: testCtx.Admin.Token, Sent: []byte(`{"number":"18VN044","name":"Essai fluvial","isr":true,
		"descript":"description","value":123456,"valuedate":"2018-08-21T02:00:00Z","length":123456,"tri":500,"van":123456}`),
			Status: http.StatusOK, BodyContains: []string{"PhysicalOp", `"number":"18VN045"`,
				`"name":"Essai fluvial"`, `"isr":true`, `"descript":"description"`, `"value":123456`,
				`"valuedate":"2018-08-21T00:00:00Z"`, `"length":123456`, `"tri":500`, `"van":123456`}},
	}

	for i, tc := range testCases {
		response := e.POST("/api/physical_ops").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, bc := range tc.BodyContains {
			if !strings.Contains(content, bc) {
				t.Errorf("CreatePhysicalOp[%d] : contenu incorrect, attendu \"%s\" et reçu\n\"%s\"", i, bc, content)
			}
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			ID = int(response.JSON().Object().Value("PhysicalOp").Object().Value("id").Number().Raw())
		}
	}
	return ID
}

// deletePhysicalOpTest tests if route is protected and destroy operation previously created.
func deletePhysicalOpTest(e *httpexpect.Expect, t *testing.T, ID int) {
	sID := strconv.Itoa(ID)
	testCases := []testCase{
		{Token: testCtx.User.Token, ID: sID, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, ID: "0", Status: http.StatusNotFound,
			BodyContains: []string{"Opération introuvable"}},
		{Token: testCtx.Admin.Token, ID: sID, Status: http.StatusOK,
			BodyContains: []string{"Opération supprimée"}},
	}

	for _, tc := range testCases {
		response := e.DELETE("/api/physical_ops/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}

//updatePhysicalOpTest tests if route is protected and fields properly updated according to role.
func updatePhysicalOpTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "", ID: "0", Status: http.StatusInternalServerError,
			Sent: []byte(`{}`), BodyContains: []string{"Token absent"}},
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusBadRequest,
			Sent: []byte(`{}`), BodyContains: []string{"Modification d'opération, introuvable"}},
		{Token: testCtx.User.Token, ID: "15", Status: http.StatusBadRequest,
			Sent: []byte(`{}`), BodyContains: []string{"L'utilisateur n'a pas de droits sur l'opération"}},
		{Token: testCtx.Admin.Token, ID: "14", Sent: []byte(`{"number":"01DI001"}`),
			Status: http.StatusBadRequest, BodyContains: []string{"Numéro d'opération existant"}},
		{Token: testCtx.User.Token, ID: "14", Sent: []byte(`{"name":"Nouveau nom","isr":true,"descript":"Nouvelle description","value":123456,"valuedate":"2018-08-17T00:00:00Z","length":123456,"tri":500,"van":123456,"plan_line_id":34}`),
			Status: http.StatusOK, BodyContains: []string{"PhysicalOp", `"name":"Bus - voirie - aménagement"`,
				`"isr":true`, `"descript":"Nouvelle description"`, `"value":123456`, `"valuedate":"2018-08-17T00:00:00Z"`,
				`"length":123456`, `"tri":500`, `"van":123456`, `"plan_line_id":32`}},
		{Token: testCtx.Admin.Token, ID: "14", Sent: []byte(`{"name":"Nom nouveau","isr":false,"descript":"Description nouvelle","value":546,"valuedate":"2018-08-16T00:00:00Z","length":546,"tri":300,"van":100,"plan_line_id":34}`),
			Status: http.StatusOK, BodyContains: []string{"PhysicalOp", `"name":"Nom nouveau"`, `"isr":false`,
				`"descript":"Description nouvelle"`, `"value":546`, `"valuedate":"2018-08-16T00:00:00Z"`,
				`"length":546`, `"tri":300`, `"van":100`, `"plan_line_id":34`}},
	}

	for i, tc := range testCases {
		response := e.PUT("/api/physical_ops/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, bc := range tc.BodyContains {
			if !strings.Contains(content, bc) {
				t.Errorf("UpdatePhysicalOp[%d] : contenu incorrect, attendu \"%s\" et reçu\n\"%s\"", i, bc, content)
			}
		}
		response.Status(tc.Status)
	}
}

// batchPhysicalOpsTest tests if route is protected and import passed.
func batchPhysicalOpsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: testCtx.User.Token, Sent: []byte(`{"PhysicalOp":[]}`),
			Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Sent: []byte(`{"PhysicalOp":[{"number":"20XX999","isr":true}]}`),
			Status: http.StatusInternalServerError, BodyContains: []string{"Erreur d'insertion"}},
		{Token: testCtx.Admin.Token, Sent: []byte(`{"PhysicalOp":[{"number":"20XX001",
		"name":"Essai batch1","isr":true},{"number":"18DI999","name":"Essai batch2","isr":true}]}`),
			Status: http.StatusOK, BodyContains: []string{"Terminé"}},
		{Token: testCtx.Admin.Token, Sent: []byte(`{"PhysicalOp":[{"number":"20XX003",
		"name":"Essai batch3","isr":true,"descript":"Description batch3","value":123,"valuedate":"2018-08-15T00:00:00Z",
		"length":123,"step":"Protocole","category":"Route","tri":500,"van":123,"action":"17700101","payment_type_id":4,"plan_line_id":20}]}`),
			Status: http.StatusOK, BodyContains: []string{"Terminé"}},
	}

	for i, tc := range testCases {
		response := e.POST("/api/physical_ops/array").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("Batch physical_ops[%d] : contenu incorrect, attendu \"%s\" et reçu\n\"%s\"", i, tc.BodyContains, content)
			}
		}
		response.Status(tc.Status)
	}

	response := e.GET("/api/physical_ops").WithHeader("Authorization", "Bearer "+testCtx.Admin.Token).Expect()
	ee := []string{"Essai batch1", "Essai batch2", "Essai batch3", "Description batch3", "20XX001"}
	content := string(response.Content)
	for _, e := range ee {
		if !strings.Contains(content, e) {
			t.Errorf("Batch physical_ops[GET] : attendu \"%s\" et reçu\n\"%s\"", e, content)
		}
	}
}

// getPrevisionsTests check route is protected and datas sent are correct
func getPrevisionsTests(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusBadRequest,
			BodyContains: []string{"Prevision d'opération, opération introuvable"}},
		{Token: testCtx.User.Token, ID: "10", Status: http.StatusOK,
			BodyContains: []string{"PrevCommitment", "PrevPayment", "FinancialCommitment", "PendingCommitment",
				"Payment", "PaymentPerBeneficiary", "FinancialCommitmentPerBeneficiary", "ImportLog"}},
	}
	for i, tc := range testCases {
		response := e.GET("/api/physical_ops/"+tc.ID+"/previsions").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("GetOpPrevisions[%d] : attendu %s et reçu\n%s", i, s, content)
			}
		}
		response.Status(tc.Status)
	}
}

// setOpPrevisionsTests check if route is protected and datas sent back correspond to post ones.
func setOpPrevisionsTests(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", ID: "0", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, ID: "0", Sent: []byte(`{Prev}`), Status: http.StatusInternalServerError,
			BodyContains: []string{"Fixation prévision d'opération, erreur décodage payload"}},
		{Token: testCtx.User.Token, ID: "0", Sent: []byte(`{"PrevCommitment":[],"PrevPayment":[]}`), Status: http.StatusBadRequest,
			BodyContains: []string{"Fixation prévision d'opération, opération introuvable"}},
		{Token: testCtx.User.Token, ID: "10", Status: http.StatusOK,
			Sent: []byte(`{"PrevCommitment":[{"year":2019,"value":100000000,"descript":null,"total_value":null,"state_ratio":null},
		{"year":2020,"value":200000000,"descript":"essai de description","total_value":400000000,"state_ratio":0.5}],
		"PrevPayment":[{"year":2019,"value":3000000,"descript":null},{"year":2020,"value":5000000,"descript":"autre essai description"}]}`),
			BodyContains: []string{"PrevCommitment", "PrevPayment", `"year":2020`, `"year":2019`, `"value":100000000`, `"descript":null`,
				`"descript":"autre essai description"`, `"value":200000000`, `"descript":"essai de description"`, `"total_value":400000000`,
				`"total_value":null`, `"value":3000000`}},
	}
	for i, tc := range testCases {
		response := e.POST("/api/physical_ops/"+tc.ID+"/previsions").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("SetOpPrevisions[%d] : attendu %s et reçu\n%s", i, s, content)
			}
		}
		response.Status(tc.Status)
	}
}
