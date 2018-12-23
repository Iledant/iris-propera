package actions

import (
	"net/http"
	"strconv"
	"strings"
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

	for i, tc := range testCases {
		response := e.GET("/api/physical_ops/9/events").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetEvent[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetEvent[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetEvent[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
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

	for i, tc := range testCases {
		response := e.GET("/api/events").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetNextMonthEvent[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetNextMonthEvent[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.ArraySize > 0 {
			count := strings.Count(content, `"id"`)
			if count != tc.ArraySize {
				t.Errorf("\nGetNextMonthEvent[%d] :\n  nombre attendu -> %d\n  nombre reçu <-%d", i, tc.ArraySize, count)
			}
		}
	}
}

// createEventTest tests route is protected and sent event is created.
func createEventTest(e *httpexpect.Expect, t *testing.T) (evID int) {
	testCases := []testCase{
		{Token: "fake", ID: "9", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, ID: "9", Status: http.StatusBadRequest, Sent: []byte(`{}`),
			BodyContains: []string{"Création d'un événement : PhysicalOpID, Name ou Date incorrect"}},
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusBadRequest,
			Sent: []byte(`{"name":"Test création événement", "date":"2018-04-01T20:00:00Z",
		"iscertain":true,"descript":"Test création événement description"}`),
			BodyContains: []string{"Création d'un événement : PhysicalOpID, Name ou Date incorrect"}},
		{Token: testCtx.User.Token, ID: "9", Status: http.StatusOK,
			Sent: []byte(`{"name":"Test création événement", "date":"2018-04-01T20:00:00Z",
			"iscertain":true,"descript":"Test création événement description"}`),
			BodyContains: []string{"Event", `"name":"Test création événement"`, `"date":"2018-04-01T20:00:00Z"`,
				`"iscertain":true`, `"descript":"Test création événement description"`}},
	}

	for i, tc := range testCases {
		response := e.POST("/api/physical_ops/"+tc.ID+"/events").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nCreateEvent[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nCreateEvent[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
		if tc.Status == http.StatusOK {
			evID = int(response.JSON().Object().Value("Event").Object().Value("id").Number().Raw())
		}
	}
	return evID
}

// modifyEventTest tests route is protected and modify work properly.
func modifyEventTest(e *httpexpect.Expect, t *testing.T, evID int) {
	testCases := []testCase{
		{Token: "fake", ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusInternalServerError,
			Sent: []byte(`{"name":"Test modification événement", "date":"2017-04-01T20:00:00Z",
		"iscertain":false,"descript":"Test modification événement description"}`),
			BodyContains: []string{"Modification d'un événement, requête : Événement introuvable"}},
		{Token: testCtx.User.Token, ID: strconv.Itoa(evID), Status: http.StatusOK,
			Sent: []byte(`{"name":"Test modification événement", "date":"2017-04-01T20:00:00Z",
			"iscertain":false,"descript":"Test modification événement description"}`),
			BodyContains: []string{"Event", `"name":"Test modification événement"`,
				`"date":"2017-04-01T20:00:00Z"`, `"iscertain":false`,
				`"descript":"Test modification événement description"`}},
	}

	for i, tc := range testCases {
		response := e.PUT("/api/physical_ops/9/events/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nModifyEvent[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nModifyEvent[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// deleteEventTest tests route is protected and delete work properly.
func deleteEventTest(e *httpexpect.Expect, t *testing.T, evID int) {
	testCases := []testCase{
		{Token: "fake", ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, ID: "0", Status: http.StatusInternalServerError,
			BodyContains: []string{"Suppression d'un événement, requête : Événement introuvable"}},
		{Token: testCtx.User.Token, ID: strconv.Itoa(evID), Status: http.StatusOK,
			BodyContains: []string{"Événement supprimé"}},
	}

	for i, tc := range testCases {
		response := e.DELETE("/api/physical_ops/9/events/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nDeleteEvent[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nDeleteEvent[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}
