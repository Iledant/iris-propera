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

//TestPhysicalOps includes all tests for physical operation handler.
func TestPhysicalOps(t *testing.T) {
	TestCommons(t)
	t.Run("PhysicalOps", func(t *testing.T) {
		getPhysicalOpsTest(testCtx.E, t)
		opID := createPhysicalOpTest(testCtx.E, t)
		// updatePhysicalOpTest(testCtx.E, t)
		deletePhysicalOpTest(testCtx.E, t, opID)
		batchPhysicalOpsTest(testCtx.E, t)
	})
}

// getPhysicalOpsTest tests if route is protected and returned list properly formatted.
func getPhysicalOpsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Status       int
		BodyContains string
		ArraySize    int
	}{
		{Token: "", Status: http.StatusInternalServerError, BodyContains: "Token absent", ArraySize: 0},
		{Token: testCtx.User.Token, Status: http.StatusOK, BodyContains: "PhysicalOp", ArraySize: 3},
		{Token: testCtx.Admin.Token, Status: http.StatusOK, BodyContains: "PhysicalOp", ArraySize: 619},
	}

	for _, tc := range testCases {
		response := e.GET("/api/physical_ops").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.Status == http.StatusOK {
			response.JSON().Object().ContainsKey("PhysicalOp")
			response.JSON().Object().Value("PhysicalOp").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

//createPhysicalOpTest tests if route is protected, validations ok and number correctly computed.
func createPhysicalOpTest(e *httpexpect.Expect, t *testing.T) int {
	testCases := []struct {
		Token        string
		Op           models.PhysicalOp
		Status       int
		BodyContains string
	}{
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: "Droits administrateurs requis"},
		{Token: testCtx.Admin.Token, Status: http.StatusBadRequest, BodyContains: "Mauvais format de numéro d'opération"},
		{Token: testCtx.Admin.Token, Op: models.PhysicalOp{Number: "99XX001", Name: ""}, Status: http.StatusBadRequest, BodyContains: "Nom de l'opération absent"},
		{Token: testCtx.Admin.Token,
			Op: models.PhysicalOp{
				Number:    "18VN044",
				Name:      "Essai fluvial",
				Isr:       true,
				Descript:  models.NullString{String: "description", Valid: true},
				Value:     models.NullInt64{Int64: 123456, Valid: true},
				ValueDate: models.NullTime{Time: time.Now(), Valid: true},
				Length:    models.NullInt64{Int64: 123456, Valid: true},
				TRI:       models.NullInt64{Int64: 500, Valid: true},
				VAN:       models.NullInt64{Int64: 123456, Valid: true}},
			Status: http.StatusOK, BodyContains: "PhysicalOp"},
	}

	var opID int
	for _, tc := range testCases {
		response := e.POST("/api/physical_ops").WithHeader("Authorization", "Bearer "+tc.Token).WithJSON(tc.Op).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.Status == http.StatusOK {
			y, m, d := tc.Op.ValueDate.Time.Date()
			jsonDate, _ := time.Date(y, m, d, 0, 0, 0, 0, time.UTC).MarshalJSON()
			dateStr := strings.Trim(string(jsonDate), "\"")
			response.JSON().Object().ContainsKey("PhysicalOp")
			response.JSON().Object().Value("PhysicalOp").Object().ContainsKey("name")
			response.JSON().Object().Value("PhysicalOp").Object().Value("name").String().Equal(tc.Op.Name)
			response.JSON().Object().Value("PhysicalOp").Object().Value("number").String().Equal("18VN045")
			response.JSON().Object().Value("PhysicalOp").Object().ContainsKey("number")
			response.JSON().Object().Value("PhysicalOp").Object().ContainsKey("isr")
			response.JSON().Object().Value("PhysicalOp").Object().Value("isr").Boolean().Equal(tc.Op.Isr)
			response.JSON().Object().Value("PhysicalOp").Object().ContainsKey("descript")
			response.JSON().Object().Value("PhysicalOp").Object().Value("descript").String().Equal(tc.Op.Descript.String)
			response.JSON().Object().Value("PhysicalOp").Object().ContainsKey("value")
			response.JSON().Object().Value("PhysicalOp").Object().Value("value").Number().Equal(tc.Op.Value.Int64)
			response.JSON().Object().Value("PhysicalOp").Object().ContainsKey("valuedate")
			response.JSON().Object().Value("PhysicalOp").Object().Value("valuedate").String().Equal(dateStr)
			response.JSON().Object().Value("PhysicalOp").Object().ContainsKey("length")
			response.JSON().Object().Value("PhysicalOp").Object().Value("length").Number().Equal(tc.Op.Length.Int64)
			response.JSON().Object().Value("PhysicalOp").Object().ContainsKey("tri")
			response.JSON().Object().Value("PhysicalOp").Object().Value("tri").Number().Equal(tc.Op.TRI.Int64)
			response.JSON().Object().Value("PhysicalOp").Object().ContainsKey("van")
			response.JSON().Object().Value("PhysicalOp").Object().Value("van").Number().Equal(tc.Op.VAN.Int64)
			opID = int(response.JSON().Object().Value("PhysicalOp").Object().Value("id").Number().Raw())
		}
		response.Status(tc.Status)
	}

	return opID
}

// deletePhysicalOpTest tests if route is protected and destroy operation previously created.
func deletePhysicalOpTest(e *httpexpect.Expect, t *testing.T, opID int) {
	sOpID := strconv.Itoa(opID)
	testCases := []struct {
		Token        string
		OpID         string
		Status       int
		BodyContains string
	}{
		{Token: testCtx.User.Token, OpID: sOpID, Status: http.StatusUnauthorized, BodyContains: "Droits administrateurs requis"},
		{Token: testCtx.Admin.Token, OpID: "0", Status: http.StatusNotFound, BodyContains: "Opération introuvable"},
		{Token: testCtx.Admin.Token, OpID: sOpID, Status: http.StatusOK, BodyContains: "Opération supprimée"},
	}

	for _, tc := range testCases {
		response := e.DELETE("/api/physical_ops/"+tc.OpID).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		response.Body().Contains(tc.BodyContains)
		response.Status(tc.Status)
	}
}

//updatePhysicalOpTest tests if route is protected and fields properly updated according to role.
func updatePhysicalOpTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		opID         string
		Op           models.PhysicalOp
		Status       int
		BodyContains string
	}{
		{Token: "", opID: "0", Status: http.StatusInternalServerError, BodyContains: "Token absent"},
		{Token: testCtx.User.Token, opID: "0", Status: http.StatusNotFound, BodyContains: "Opération introuvable"},
		{Token: testCtx.User.Token, opID: "15", Status: http.StatusBadRequest, BodyContains: "L'utilisateur n'a pas de droits sur l'opération"},
		{Token: testCtx.Admin.Token, opID: "14",
			Op: models.PhysicalOp{
				Number: "01DI001"},
			Status: http.StatusBadRequest, BodyContains: "Numéro d'opération existant"},
		{Token: testCtx.User.Token, opID: "14",
			Op: models.PhysicalOp{
				Name:       "Nouveau nom",
				Isr:        true,
				Descript:   models.NullString{String: "Nouvelle description", Valid: true},
				Value:      models.NullInt64{Int64: 123456, Valid: true},
				ValueDate:  models.NullTime{Time: time.Now(), Valid: true},
				Length:     models.NullInt64{Int64: 123456, Valid: true},
				TRI:        models.NullInt64{Int64: 500, Valid: true},
				VAN:        models.NullInt64{Int64: 123456, Valid: true},
				PlanLineID: models.NullInt64{Int64: 34, Valid: true}},
			Status: http.StatusOK, BodyContains: "PhysicalOp"},
		{Token: testCtx.Admin.Token, opID: "14",
			Op: models.PhysicalOp{
				Name:       "Nom nouveau",
				Isr:        false,
				Descript:   models.NullString{String: "Description nouvelle", Valid: true},
				Value:      models.NullInt64{Int64: 546, Valid: true},
				ValueDate:  models.NullTime{Time: time.Now(), Valid: true},
				Length:     models.NullInt64{Int64: 546, Valid: true},
				TRI:        models.NullInt64{Int64: 300, Valid: true},
				VAN:        models.NullInt64{Int64: 100, Valid: true},
				PlanLineID: models.NullInt64{Int64: 34, Valid: true}},
			Status: http.StatusOK, BodyContains: "PhysicalOp"},
	}

	for _, tc := range testCases {
		response := e.PUT("/api/physical_ops/"+tc.opID).WithHeader("Authorization", "Bearer "+tc.Token).WithJSON(tc.Op).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.Status == http.StatusOK {
			y, m, d := tc.Op.ValueDate.Time.Date()
			jsonDate, _ := time.Date(y, m, d, 0, 0, 0, 0, time.UTC).MarshalJSON()
			dateStr := strings.Trim(string(jsonDate), "\"")
			response.JSON().Object().ContainsKey("PhysicalOp")
			response.JSON().Object().Value("PhysicalOp").Object().ContainsKey("name")
			if tc.Token == testCtx.Admin.Token {
				response.JSON().Object().Value("PhysicalOp").Object().Value("name").String().Equal(tc.Op.Name)
				response.JSON().Object().Value("PhysicalOp").Object().Value("plan_line_id").Number().Equal(tc.Op.PlanLineID.Int64)
			} else {
				response.JSON().Object().Value("PhysicalOp").Object().Value("name").String().NotEqual(tc.Op.Name)
				response.JSON().Object().Value("PhysicalOp").Object().Value("plan_line_id").Number().NotEqual(tc.Op.PlanLineID.Int64)
			}
			response.JSON().Object().Value("PhysicalOp").Object().ContainsKey("number")
			response.JSON().Object().Value("PhysicalOp").Object().ContainsKey("isr")
			response.JSON().Object().Value("PhysicalOp").Object().Value("isr").Boolean().Equal(tc.Op.Isr)
			response.JSON().Object().Value("PhysicalOp").Object().ContainsKey("descript")
			response.JSON().Object().Value("PhysicalOp").Object().Value("descript").String().Equal(tc.Op.Descript.String)
			response.JSON().Object().Value("PhysicalOp").Object().ContainsKey("value")
			response.JSON().Object().Value("PhysicalOp").Object().Value("value").Number().Equal(tc.Op.Value.Int64)
			response.JSON().Object().Value("PhysicalOp").Object().ContainsKey("valuedate")
			response.JSON().Object().Value("PhysicalOp").Object().Value("valuedate").String().Equal(dateStr)
			response.JSON().Object().Value("PhysicalOp").Object().ContainsKey("length")
			response.JSON().Object().Value("PhysicalOp").Object().Value("length").Number().Equal(tc.Op.Length.Int64)
			response.JSON().Object().Value("PhysicalOp").Object().ContainsKey("tri")
			response.JSON().Object().Value("PhysicalOp").Object().Value("tri").Number().Equal(tc.Op.TRI.Int64)
			response.JSON().Object().Value("PhysicalOp").Object().ContainsKey("van")
			response.JSON().Object().Value("PhysicalOp").Object().Value("van").Number().Equal(tc.Op.VAN.Int64)
		}
		response.Status(tc.Status)
	}
}

type opIncomplete struct {
	Number string
	Isr    bool
}

type batchIncomplete struct {
	PhysicalOps []opIncomplete `json:"PhysicalOp"`
}
type opPartial struct {
	Name   string
	Number string
	Isr    bool
}

type batchPartial struct {
	PhysicalOps []opPartial `json:"PhysicalOp"`
}

type opComplete struct {
	Number        string
	Name          string
	Descript      string
	Isr           bool
	Value         int64
	Valuedate     time.Time
	Length        int64
	Step          string
	Category      string
	Tri           int64
	Van           int64
	Action        string
	PaymentTypeID int64 `json:"payment_types_id"`
	PlanLineID    int64 `json:"plan_line_id"`
}

type batchComplete struct {
	PhysicalOps []opComplete `json:"PhysicalOp"`
}

// batchPhysicalOpsTest tests if route is protected and import passed.
func batchPhysicalOpsTest(e *httpexpect.Expect, t *testing.T) {
	inc := []opPartial{{Name: "Essai batch1", Number: "20XX001", Isr: true},
		{Name: "Essai batch2", Number: "18DI999", Isr: true}}
	com := []opComplete{{Name: "Essai batch3", Number: "20XX003", Isr: true, Descript: "Description batch3", Value: 123, Valuedate: time.Now(), Length: 123, Step: "Protocole", Category: "Route", Tri: 500, Van: 123, Action: "17700101", PaymentTypeID: 4, PlanLineID: 20}}
	testCases := []struct {
		Token        string
		Incomplete   *batchIncomplete
		Partial      *batchPartial
		Complete     *batchComplete
		Status       int
		BodyContains string
	}{
		{Token: testCtx.User.Token, Incomplete: nil, Partial: nil, Complete: nil, Status: http.StatusUnauthorized, BodyContains: "Droits administrateurs requis"},
		{Token: testCtx.Admin.Token, Incomplete: &batchIncomplete{PhysicalOps: []opIncomplete{{Number: "20XX999", Isr: true}}}, Partial: nil, Complete: nil, Status: http.StatusInternalServerError, BodyContains: "Erreur d'insertion"},
		{Token: testCtx.Admin.Token, Incomplete: nil, Partial: &batchPartial{PhysicalOps: inc}, Complete: nil, Status: http.StatusOK, BodyContains: "Terminé"},
		{Token: testCtx.Admin.Token, Incomplete: nil, Partial: nil, Complete: &batchComplete{PhysicalOps: com}, Status: http.StatusOK, BodyContains: "Terminé"},
	}

	for _, tc := range testCases {
		req := e.POST("/api/physical_ops/array").WithHeader("Authorization", "Bearer "+tc.Token)
		if tc.Incomplete != nil {
			req = req.WithJSON(*tc.Incomplete)
		} else {
			if tc.Partial != nil {
				req = req.WithJSON(*tc.Partial)
			} else {
				if tc.Complete != nil {
					req = req.WithJSON(*tc.Complete)
				}
			}
		}
		response := req.Expect()
		response.Body().Contains(tc.BodyContains)
		response.Status(tc.Status)
	}

	response := e.GET("/api/physical_ops").WithHeader("Authorization", "Bearer "+testCtx.Admin.Token).Expect()
	response.Body().Contains("Essai batch1")
	response.Body().Contains("Essai batch2")
	response.Body().Contains("Essai batch3")
	response.Body().Contains("Description batch3")
	response.Body().NotContains("20XX999")
}
