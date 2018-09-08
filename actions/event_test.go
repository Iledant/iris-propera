package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestEvent embeddes all tests for event insuring the configuration and DB are properly initialized.
func TestEvent(t *testing.T) {
	TestCommons(t)
	t.Run("Event", func(t *testing.T) {
		getEventTest(testCtx.E, t)
		evID := createEventTest(testCtx.E, t)
		modifyEventTest(testCtx.E, t, evID)
		deleteEventTest(testCtx.E, t, evID)
	})
}

// getEventTest tests route is protected and all events are sent back.
func getEventTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}, ArraySize: 0},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"Event"}, ArraySize: 1},
	}

	for _, tc := range testCases {
		response := e.GET("/api/physical_ops/9/events").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("Event").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

// getNextMonthEventTest tests route is protected and all events are sent back.
func getNextMonthEventTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}, ArraySize: 0},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"Event"}, ArraySize: 0},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{"Event"}, ArraySize: 0},
	}

	for _, tc := range testCases {
		response := e.GET("/api/events").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("Event").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

// createEventTest tests route is protected and sent event is created.
func createEventTest(e *httpexpect.Expect, t *testing.T) (evID int) {
	testCases := []testCase{
		{Token: "fake", ID: "9", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, ID: "9", Status: http.StatusBadRequest, Sent: []byte(`{}`),
			BodyContains: []string{"Création d'événement, champ manquant ou incorrect"}},
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusBadRequest, Sent: []byte(`{}`),
			BodyContains: []string{"Création d'événement : opération introuvable"}},
		{Token: testCtx.User.Token, ID: "9", Status: http.StatusOK,
			Sent: []byte(`{"name":"Test création événement", "date":"2018-04-01T20:00:00Z",
			"iscertain":true,"descript":"Test création événement description"}`),
			BodyContains: []string{"Event", `"name":"Test création événement"`, `"date":"2018-04-01T20:00:00Z"`,
				`"iscertain":true`, `"descript":"Test création événement description"`}},
	}

	for _, tc := range testCases {
		response := e.POST("/api/physical_ops/"+tc.ID+"/events").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		if tc.Status == http.StatusOK {
			evID = int(response.JSON().Object().Value("Event").Object().Value("id").Number().Raw())
		}
		response.Status(tc.Status)
	}
	return evID
}

// modifyEventTest tests route is protected and modify work properly.
func modifyEventTest(e *httpexpect.Expect, t *testing.T, evID int) {
	testCases := []testCase{
		{Token: "fake", ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusBadRequest,
			BodyContains: []string{"Modification d'événement : introuvable"}},
		{Token: testCtx.User.Token, ID: strconv.Itoa(evID), Status: http.StatusOK,
			Sent: []byte(`{"name":"Test modification événement", "date":"2017-04-01T20:00:00Z",
			"iscertain":false,"descript":"Test modification événement description"}`),
			BodyContains: []string{"Event", `"name":"Test modification événement"`,
				`"date":"2017-04-01T20:00:00Z"`, `"iscertain":false`,
				`"descript":"Test modification événement description"`}},
	}

	for _, tc := range testCases {
		response := e.PUT("/api/physical_ops/9/events/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}

// deleteEventTest tests route is protected and delete work properly.
func deleteEventTest(e *httpexpect.Expect, t *testing.T, evID int) {
	testCases := []testCase{
		{Token: "fake", ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusNotFound,
			BodyContains: []string{"Suppression d'événement : introuvable"}},
		{Token: testCtx.User.Token, ID: strconv.Itoa(evID), Status: http.StatusOK,
			BodyContains: []string{"Événement supprimé"}},
	}

	for _, tc := range testCases {
		response := e.DELETE("/api/physical_ops/9/events/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		for _, s := range tc.BodyContains {
			response.Body().Contains(s)
		}
		response.Status(tc.Status)
	}
}
