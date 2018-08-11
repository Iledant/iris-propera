package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/iris-contrib/httpexpect"
)

type rightTest struct {
	Right []int `json:"Right"`
}

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
	testCases := []struct {
		UserID, Token string
		Status        int
		BodyContains  string
		ArraySize     int
	}{
		{UserID: "26", Token: testCtx.User.Token, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis", ArraySize: 0},
		{UserID: "0", Token: testCtx.Admin.Token, Status: http.StatusBadRequest, BodyContains: "Utilisateur introuvable", ArraySize: 0},
		{UserID: "26", Token: testCtx.Admin.Token, Status: http.StatusOK, BodyContains: "Right", ArraySize: 49},
	}

	for _, tc := range testCases {
		response := e.GET("/api/users/"+tc.UserID+"/rights").WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.ArraySize > 0 {
			response.JSON().Object().ContainsKey("Right")
			response.JSON().Object().ContainsKey("User")
			response.JSON().Object().ContainsKey("PhysicalOp")
			response.JSON().Object().Value("Right").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

// setRightTest tests route is protected and returned list is the same as sent one
func setRightsTest(e *httpexpect.Expect, t *testing.T) {
	userID := strconv.Itoa(testCtx.User.User.ID)
	testCases := []struct {
		UserID, Token string
		Right         rightTest
		Status        int
		BodyContains  string
		ArraySize     int
	}{
		{UserID: "26", Token: testCtx.User.Token, Right: rightTest{}, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis", ArraySize: 0},
		{UserID: "0", Token: testCtx.Admin.Token, Right: rightTest{}, Status: http.StatusBadRequest, BodyContains: "Utilisateur introuvable", ArraySize: 0},
		{UserID: userID, Token: testCtx.Admin.Token, Right: rightTest{[]int{}}, Status: http.StatusOK, BodyContains: "Right", ArraySize: 0},
		{UserID: userID, Token: testCtx.Admin.Token, Right: rightTest{[]int{0}}, Status: http.StatusBadRequest, BodyContains: "Mauvais identificateur d'opération", ArraySize: 0},
		{UserID: userID, Token: testCtx.Admin.Token, Right: rightTest{[]int{9, 10, 11, 12}}, Status: http.StatusOK, BodyContains: "Right", ArraySize: 4},
	}

	for _, tc := range testCases {
		response := e.POST("/api/users/"+tc.UserID+"/rights").WithHeader("Authorization", "Bearer "+tc.Token).WithJSON(tc.Right).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("Right").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}
}

// inheritRightsTest tests route is protected and returned list is compliant.
func inheritsRightsTest(e *httpexpect.Expect, t *testing.T) {
	userID := strconv.Itoa(testCtx.User.User.ID)
	testCases := []struct {
		UserID, Token string
		Right         rightTest
		Status        int
		BodyContains  string
		ArraySize     int
	}{
		{UserID: "26", Token: testCtx.User.Token, Right: rightTest{}, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis", ArraySize: 0},
		{UserID: "0", Token: testCtx.Admin.Token, Right: rightTest{}, Status: http.StatusBadRequest, BodyContains: "Utilisateur introuvable", ArraySize: 0},
		{UserID: userID, Token: testCtx.Admin.Token, Right: rightTest{[]int{}}, Status: http.StatusOK, BodyContains: "Right", ArraySize: 4},
		{UserID: userID, Token: testCtx.Admin.Token, Right: rightTest{[]int{0}}, Status: http.StatusBadRequest, BodyContains: "Mauvais identificateur d'utilisateur", ArraySize: 0},
		{UserID: userID, Token: testCtx.Admin.Token, Right: rightTest{[]int{26}}, Status: http.StatusOK, BodyContains: "Right", ArraySize: 53},
	}

	for _, tc := range testCases {
		response := e.POST("/api/users/"+tc.UserID+"/inherits").WithHeader("Authorization", "Bearer "+tc.Token).WithJSON(tc.Right).Expect()
		response.Body().Contains(tc.BodyContains)
		if tc.ArraySize > 0 {
			response.JSON().Object().Value("Right").Array().Length().Equal(tc.ArraySize)
		}
		response.Status(tc.Status)
	}

}
