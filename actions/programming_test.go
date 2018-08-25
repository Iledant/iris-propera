package actions

import (
	"net/http"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func TestProgramming(t *testing.T) {
	TestCommons(t)
	t.Run("Programming", func(t *testing.T) {
		getProgrammingsTest(testCtx.E, t)
		getProgrammingsYearsTest(testCtx.E, t)
		batchProgrammingsTest(testCtx.E, t)
	})
}

// getProgrammingsTest check route is protected and programmings correctly sent.
func getProgrammingsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Status       int
		BodyContains []string
		Year         string
		Count        int
	}{
		{Token: "fake", Year: "2018", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.Admin.Token, Year: "2018", Status: http.StatusOK,
			BodyContains: []string{"Programmings"}, Count: 623},
	}
	for i, tc := range testCases {
		response := e.GET("/api/programmings").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQuery("year", tc.Year).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("GetProgrammings[%d] : attendu \"%s\" et reçu \"%s\"", i, s, content)
			}
		}
		response.Status(tc.Status)
		if tc.Status == http.StatusOK {
			response.JSON().Object().Value("Programmings").Array().Length().Equal(tc.Count)
		}
	}
}

// getProgrammingsYearsTest check route is protected and programmings correctly sent.
func getProgrammingsYearsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Status       int
		BodyContains []string
	}{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{`"ProgrammingsYear":[2018]`}},
	}
	for i, tc := range testCases {
		response := e.GET("/api/programmings/years").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("GetProgrammingsYears[%d] : attendu \"%s\" et reçu \"%s\"", i, s, content)
			}
		}
		response.Status(tc.Status)
	}
}

// batchProgrammingsTest check route is protected and return successful.
func batchProgrammingsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token        string
		Sent         []byte
		Status       int
		BodyContains []string
	}{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized, Sent: []byte(`{Pend}`),
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusInternalServerError, Sent: []byte(`{Pend}`),
			BodyContains: []string{"Batch programmation, décodage impossible"}},
		//cSpell:disable
		{Token: testCtx.Admin.Token, Status: http.StatusOK, Sent: []byte(`{"Programmings": [
			{"physical_op_id":9,"year":2018,"value":100000000,
			"commission_id":7,"total_value":null,"state_ratio":null},
			{"physical_op_id":10,"year":2018,"value":200000000,
			"commission_id":8,"total_value":400000000,"state_ratio":null},
			{"physical_op_id":14 ,"year":2018,"value":300000000,
			"commission_id":3,"total_value":600000000,"state_ratio":0.35}],
			"year":2018}`),
			BodyContains: []string{"Programmngs", `"physical_op_id":9`, `"physical_op_id":10`,
				`"physical_op_id":14`, `"value":200000000`, `"commission_id":8`, `"total_value":400000000`,
				`"total_value":null`, `"state_ratio":null`, `"state_ratio":0.35`}},
	}
	//cSpell:enable
	for i, tc := range testCases {
		response := e.POST("/api/programmings").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("BatchProgrammings[%d] : attendu \"%s\" et reçu \"%s\"", i, s, content)
			}
		}
		response.Status(tc.Status)
	}
}
