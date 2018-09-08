package actions

import (
	"net/http"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

func TestTodayMessage(t *testing.T) {
	TestCommons(t)
	t.Run("TodayMessage", func(t *testing.T) {
		getTodayMessageTest(testCtx.E, t)
		setTodayMessageTest(testCtx.E, t)
	})
}

// getTodayMessageTest check route is protected and datas sent has got items and number of lines.
func getTodayMessageTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError, BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusOK,
			BodyContains: []string{"TodayMessage", "title", "text"}},
	}
	for i, tc := range testCases {
		response := e.GET("/api/today_message").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("GetTodayMessage[%d] : attendu %s et reçu \n%s", i, s, content)
			}
		}
		response.Status(tc.Status)
	}
}

// setTodayMessageTest check route is protected and datas sent has got items and number of lines.
func setTodayMessageTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{Token: "fake", Status: http.StatusInternalServerError,
			BodyContains: []string{"Token invalide"}},
		{Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{Token: testCtx.Admin.Token, Status: http.StatusOK,
			Sent:         []byte(`{"title":"Essai de titre","text":"Essai de texte"}`),
			BodyContains: []string{"TodayMessage", `"title":"Essai de titre"`, `"text":"Essai de texte"`}},
	}
	for i, tc := range testCases {
		response := e.POST("/api/today_message").WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("SetTodayMessage[%d] : attendu %s et reçu \n%s", i, s, content)
			}
		}
		response.Status(tc.Status)
	}
}
