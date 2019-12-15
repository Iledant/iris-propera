package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestEvent embeddes all tests for event insuring the configuration and DB are properly initialized.
func testEvent(t *testing.T) {
	t.Run("Event", func(t *testing.T) {
		getEventTest(testCtx.E, t)
		evID := createEventTest(testCtx.E, t)
		if evID == 0 {
			t.Fatal("Impossible de créer l'événement")
		}
		modifyEventTest(testCtx.E, t, evID)
		deleteEventTest(testCtx.E, t, evID)
	})
}

// getEventTest tests route is protected and all events are sent back.
func getEventTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			Status:       http.StatusOK,
			BodyContains: []string{"Event"},
			IDName:       `"id"`,
			ArraySize:    1},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/physical_ops/9/events").
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetEvent") {
		t.Error(r)
	}
}

// getNextMonthEventTest tests route is protected and all events are sent back.
func getNextMonthEventTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:         testCtx.User.Token,
			Status:        http.StatusOK,
			BodyContains:  []string{"Event"},
			CountItemName: `"id"`,
			ArraySize:     0},
		{
			Token:         testCtx.Admin.Token,
			Status:        http.StatusOK,
			BodyContains:  []string{"Event"},
			CountItemName: `"id"`,
			ArraySize:     0},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/events").WithHeader("Authorization", "Bearer "+tc.Token).
			Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetNextMonthEvent") {
		t.Error(r)
	}
}

// createEventTest tests route is protected and sent event is created.
func createEventTest(e *httpexpect.Expect, t *testing.T) (evID int) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			ID:           "9",
			Status:       http.StatusBadRequest,
			Sent:         []byte(`{}`),
			BodyContains: []string{"Création d'un événement : PhysicalOpID, Name ou Date incorrect"}},
		{
			Token:  testCtx.User.Token,
			ID:     "0",
			Status: http.StatusBadRequest,
			Sent: []byte(`{"name":"Test création événement", "date":"2018-04-01T20:00:00Z",
		"iscertain":true,"descript":"Test création événement description"}`),
			BodyContains: []string{"Création d'un événement : PhysicalOpID, Name ou Date incorrect"}},
		{
			Token:  testCtx.User.Token,
			ID:     "9",
			Status: http.StatusInternalServerError,
			Sent: []byte(`{"name":"Test création événement", "date":"2018-04-01T20:00:00Z",
			"iscertain":true,"descript":"Test création événement description"`),
			BodyContains: []string{"Création d'un événement, décodage : "}},
		{
			Token:  testCtx.User.Token,
			ID:     "9",
			Status: http.StatusCreated,
			IDName: `"id"`,
			Sent: []byte(`{"name":"Test création événement", "date":"2018-04-01T20:00:00Z",
			"iscertain":true,"descript":"Test création événement description"}`),
			BodyContains: []string{"Event", `"name":"Test création événement"`, `"date":"2018-04-01T20:00:00Z"`,
				`"iscertain":true`, `"descript":"Test création événement description"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/physical_ops/"+tc.ID+"/events").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "CreateEvent", &evID) {
		t.Error(r)
	}
	return evID
}

// modifyEventTest tests route is protected and modify work properly.
func modifyEventTest(e *httpexpect.Expect, t *testing.T, evID int) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:  testCtx.User.Token,
			ID:     "0",
			Status: http.StatusInternalServerError,
			Sent: []byte(`{"name":"Test modification événement", "date":"2017-04-01T20:00:00Z",
		"iscertain":false,"descript":"Test modification événement description"}`),
			BodyContains: []string{"Modification d'un événement, requête : Événement introuvable"}},
		{
			Token:  testCtx.User.Token,
			ID:     strconv.Itoa(evID),
			Status: http.StatusInternalServerError,
			Sent: []byte(`{"name":"Test modification événement", "date":"2017-04-01T20:00:00Z",
			"iscertain":false,"descript":"Test modification événement description"`),
			BodyContains: []string{"Modification d'un événement, décodage :"}},
		{
			Token:  testCtx.User.Token,
			ID:     strconv.Itoa(evID),
			Status: http.StatusOK,
			Sent: []byte(`{"name":"Test modification événement", "date":"2017-04-01T20:00:00Z",
			"iscertain":false,"descript":"Test modification événement description"}`),
			BodyContains: []string{"Event", `"name":"Test modification événement"`,
				`"date":"2017-04-01T20:00:00Z"`, `"iscertain":false`,
				`"descript":"Test modification événement description"`}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.PUT("/api/physical_ops/9/events/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "ModifyEvent") {
		t.Error(r)
	}
}

// deleteEventTest tests route is protected and delete work properly.
func deleteEventTest(e *httpexpect.Expect, t *testing.T, evID int) {
	testCases := []testCase{
		notLoggedTestCase,
		{
			Token:        testCtx.User.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Suppression d'un événement, requête : Événement introuvable"}},
		{
			Token:        testCtx.User.Token,
			ID:           strconv.Itoa(evID),
			Status:       http.StatusOK,
			BodyContains: []string{"Événement supprimé"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.DELETE("/api/physical_ops/9/events/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DeleteEvent") {
		t.Error(r)
	}
}
