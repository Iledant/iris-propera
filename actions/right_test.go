package actions

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

// TestRight implements tests for users right handlers.
func TestRight(t *testing.T) {
	TestCommons(t)
	t.Run("Right", func(t *testing.T) {
		getRightsTest(testCtx.E, t)
		setRightsTest(testCtx.E, t)
		inheritsRightsTest(testCtx.E, t)
	})
}

// getRightTest tests list returned correctly and route is protected
func getRightsTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{ID: "26", Token: testCtx.User.Token, Status: http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}, ArraySize: 0},
		{ID: "26", Token: testCtx.Admin.Token, Status: http.StatusOK,
			BodyContains: []string{`"Right":[536,25,41,54,56,477,140,444,446,447,448,449,450,451,452,453,19,372,389,391,445,454,478,479,481,482,483,164,165,166,182,194,197,220,294,300,321,333,340,349,357,480,484,486,487,488,489,390,551]`, "User", "PhysicalOp"}},
	}
	for i, tc := range testCases {
		response := e.GET("/api/user/"+tc.ID+"/rights").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nGetRights[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nGetRights[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// setRightTest tests route is protected and returned list is the same as sent one
func setRightsTest(e *httpexpect.Expect, t *testing.T) {
	userID := strconv.Itoa(testCtx.User.User.ID)
	testCases := []testCase{
		{ID: "26", Token: testCtx.User.Token, Sent: []byte(`{"Right":[]}`),
			Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{ID: userID, Token: testCtx.Admin.Token, Sent: []byte(`{"Right":[]}`),
			Status: http.StatusOK, BodyContains: []string{"Right"}},
		{ID: userID, Token: testCtx.Admin.Token, Sent: []byte(`{"Right":[0]}`),
			Status: http.StatusInternalServerError, BodyContains: []string{"Fixation des droits, requête :"}},
		{ID: userID, Token: testCtx.Admin.Token, Sent: []byte(`{"Right":[9, 10, 11, 12]}`),
			Status: http.StatusOK, BodyContains: []string{`"Right":[9,10,11,12]`}},
	}
	for i, tc := range testCases {
		response := e.POST("/api/user/"+tc.ID+"/rights").
			WithHeader("Authorization", "Bearer "+tc.Token).WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nSetRights[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nSetRights[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// inheritRightsTest tests route is protected and returned list is compliant.
func inheritsRightsTest(e *httpexpect.Expect, t *testing.T) {
	userID := strconv.Itoa(testCtx.User.User.ID)
	testCases := []testCase{
		{ID: "26", Token: testCtx.User.Token, Sent: []byte(`{"Right":[]}`),
			Status: http.StatusUnauthorized, BodyContains: []string{"Droits administrateur requis"}},
		{ID: userID, Token: testCtx.Admin.Token, Sent: []byte(`{"Right":[]}`),
			Status: http.StatusOK, BodyContains: []string{"Right"}},
		{ID: userID, Token: testCtx.Admin.Token, Sent: []byte(`{"Right":[35]}`),
			Status: http.StatusOK, BodyContains: []string{"Right", "37", "541", "543"}},
	}
	for i, tc := range testCases {
		response := e.POST("/api/user/"+tc.ID+"/inherits").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
		content := string(response.Content)
		for _, s := range tc.BodyContains {
			if !strings.Contains(content, s) {
				t.Errorf("\nInheritsRights[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, s, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nInheritsRights[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}
