package actions

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/Iledant/iris_propera/config"
	"github.com/iris-contrib/httpexpect"
)

type SentUserCase struct {
	Sent         []byte
	Status       int
	BodyContains string
}

// UserTest includes all tests for users
func testUser(t *testing.T, userCredentials *config.Credentials) {
	t.Run("User", func(t *testing.T) {
		getUsers(testCtx.E, t)
		createdUID := createUser(testCtx.E, t)
		updateUser(testCtx.E, t, createdUID)
		chgPwd(testCtx.E, t)
		deleteUser(testCtx.E, t, createdUID)
		signupTest(testCtx.E, t)
		logoutTest(testCtx.E, t, userCredentials)
	})
}

// getUsers test list returned
func getUsers(e *httpexpect.Expect, t *testing.T) {
	response := e.GET("/api/user").
		WithHeader("Authorization", "Bearer "+testCtx.Admin.Token).Expect()
	content := string(response.Content)
	if !strings.Contains(content, "User") {
		t.Errorf("\nGetUsers[admin] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", "user", content)
	}
	statusCode := response.Raw().StatusCode
	if statusCode != http.StatusOK {
		t.Errorf("\nGetUsers[admin],statut :  attendu ->%v  reçu <-%v", http.StatusOK, statusCode)
	}
	response = e.GET("/api/user").
		WithHeader("Authorization", "Bearer "+testCtx.User.Token).Expect()
	statusCode = response.Raw().StatusCode
	if statusCode != http.StatusUnauthorized {
		t.Errorf("\nGetUsers[user],statut :  attendu ->%v  reçu <-%v", http.StatusUnauthorized, statusCode)
	}
}

//createUser test user creation and get userID for other tests
func createUser(e *httpexpect.Expect, t *testing.T) (createdUID string) {
	cts := []SentUserCase{
		{Sent: []byte(`{"name":"Essai6","email":"essai5@iledefrance.fr","password":"toto","role":"USER","active":false}`),
			Status: http.StatusBadRequest, BodyContains: "existant"},
		{Sent: []byte(`{"name":"","email":"essai@iledefrance.fr","password":"toto","role":"USER","active":false}`),
			Status: http.StatusBadRequest, BodyContains: "Champ manquant ou incorrect"},
		{Sent: []byte(`{"name":"essai4","email":"essai@iledefrance.fr","password":"toto","role":"FALSE","active":false}`),
			Status: http.StatusBadRequest, BodyContains: "Champ manquant ou incorrect"},
		{Sent: []byte(`{"name":"essai","email":"essai@iledefrance.fr","password":"toto","role":"USER","active":false}`),
			Status: http.StatusCreated, BodyContains: "essai"},
	}

	var response *httpexpect.Response
	for i, ct := range cts {
		response = e.POST("/api/user").WithHeader("Authorization", "Bearer "+testCtx.Admin.Token).WithBytes(ct.Sent).Expect()
		content := string(response.Content)
		if !strings.Contains(content, ct.BodyContains) {
			t.Errorf("\nCreateUser[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, ct.BodyContains, content)
		}
		statusCode := response.Raw().StatusCode
		if statusCode != ct.Status {
			t.Errorf("\nCreateUser[%d],statut :  attendu ->%v  reçu <-%v", i, ct.Status, statusCode)
		}
		if ct.Status == http.StatusCreated {
			createdUID = strconv.Itoa(int(response.JSON().Object().Value("User").Object().Value("id").Number().Raw()))
		}
	}
	return createdUID
}

// updateUser tests changing properties of previously created user
func updateUser(e *httpexpect.Expect, t *testing.T, createdUID string) {
	cts := []SentUserCase{
		{Sent: []byte(`{"name":"","email":"","password":"","role":"","active":false}`),
			Status: http.StatusOK, BodyContains: "essai"},
		{Sent: []byte(`{"name":"Essai2","email":"","password":"","role":"","active":true}`),
			Status: http.StatusOK, BodyContains: `"name":"Essai2"`},
		{Sent: []byte(`{"name":"","email":"essai2@iledefrance.fr","password":"","role":"","active":true}`),
			Status: http.StatusOK, BodyContains: `"email":"essai2@iledefrance.fr"`},
		{Sent: []byte(`{"name":"","email":"","password":"","role":"ADMIN","active":true}`),
			Status: http.StatusOK, BodyContains: `"role":"ADMIN"`},
		{Sent: []byte(`{"name":"","email":"","password":"","role":"FAIL","active":true}`),
			Status: http.StatusBadRequest, BodyContains: "Modification d'utilisateur, rôle incorrect"},
	}

	for i, ct := range cts {
		response := e.PUT("/api/user/"+createdUID).
			WithHeader("Authorization", "Bearer "+testCtx.Admin.Token).WithBytes(ct.Sent).Expect()
		content := string(response.Content)
		if !strings.Contains(content, ct.BodyContains) {
			t.Errorf("\nUpdateUser[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, ct.BodyContains, content)
		}
		statusCode := response.Raw().StatusCode
		if statusCode != ct.Status {
			t.Errorf("\nUpdateUser[%d],statut :  attendu ->%v  reçu <-%v", i, ct.Status, statusCode)
		}
	}

	response := e.PUT("/post/users/0").WithHeader("Authorization", "Bearer "+testCtx.Admin.Token).
		Expect()
	statusCode := response.Raw().StatusCode
	if statusCode != http.StatusNotFound {
		t.Errorf("\nUpdateUser[final],statut :  attendu ->%v  reçu <-%v", http.StatusNotFound, statusCode)
	}
}

// chgPwd tests the request for the connected user
func chgPwd(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Old, New, BodyContains string
		Status                 int
	}{{"fake", "tutu", "Erreur de mot de passe", http.StatusBadRequest},
		{"toto", "tutu", "Mot de passe changé", http.StatusOK},
		{"tutu", "toto", "Mot de passe changé", http.StatusOK}}

	for i, tc := range testCases {
		response := e.POST("/api/user/password").WithQuery("current_password", tc.Old).WithQuery("password", tc.New).
			WithHeader("Authorization", "Bearer "+testCtx.User.Token).Expect()
		content := string(response.Content)
		if !strings.Contains(content, tc.BodyContains) {
			t.Errorf("\nChgPwd[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, tc.BodyContains, content)
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nChgPwd[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// deleteUser test admin deleting of an user
func deleteUser(e *httpexpect.Expect, t *testing.T, createdUID string) {
	testCases := []struct {
		Token, UserID string
		Status        int
		BodyContains  string
	}{
		{Token: testCtx.User.Token, UserID: createdUID, Status: http.StatusUnauthorized, BodyContains: "Droits administrateur requis"},
		{Token: testCtx.Admin.Token, UserID: "", Status: http.StatusNotFound, BodyContains: ""},
		{Token: testCtx.Admin.Token, UserID: strconv.Itoa(0), Status: http.StatusInternalServerError, BodyContains: "Utilisateur introuvable"},
		{Token: testCtx.Admin.Token, UserID: createdUID, Status: http.StatusOK, BodyContains: "Utilisateur supprimé"},
	}

	for i, tc := range testCases {
		response := e.DELETE("/api/user/"+tc.UserID).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		content := string(response.Content)
		if tc.BodyContains != "" {
			if !strings.Contains(content, tc.BodyContains) {
				t.Errorf("\nDeleteUser[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, tc.BodyContains, content)
			}
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nDeleteUser[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// signupTest check functionality for a user signing up
func signupTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Payload      string
		Status       int
		BodyContains string
	}{
		{
			Payload:      `{"Name": "Nouveau", "Email": "", "Password": ""}`,
			Status:       http.StatusBadRequest,
			BodyContains: "Inscription d'utilisateur : nom, email ou mot de passe manquant"},
		{
			Payload:      `{"Name": "Nouveau", "Email": "nouveau@iledefrance.fr", "Password": ""}`,
			Status:       http.StatusBadRequest,
			BodyContains: "Inscription d'utilisateur : nom, email ou mot de passe manquant"},
		{
			Payload:      `{"Name": "Nouveau", "Email": "essai5@iledefrance.fr", "Password": "nouveau"}`,
			Status:       http.StatusBadRequest,
			BodyContains: "Utilisateur existant"},
		{
			Payload:      `{"Name": "Nouveau", "Email": "nouveau@iledefrance.fr", "Password": "nouveau"}`,
			Status:       http.StatusCreated,
			BodyContains: "Utilisateur créé"},
	}

	for i, tc := range testCases {
		response := e.POST("/api/user/signup").WithBytes([]byte(tc.Payload)).Expect()
		content := string(response.Content)
		if !strings.Contains(content, tc.BodyContains) {
			t.Errorf("\nSignUp[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", i, tc.BodyContains, content)
		}
		statusCode := response.Raw().StatusCode
		if statusCode != tc.Status {
			t.Errorf("\nSignUp[%d],statut :  attendu ->%v  reçu <-%v", i, tc.Status, statusCode)
		}
	}
}

// logoutTest for a connected user
func logoutTest(e *httpexpect.Expect, t *testing.T, userCredentials *config.Credentials) {
	response := e.POST("/api/user/logout").WithHeader("Authorization", "Bearer "+testCtx.User.Token).Expect()
	content := string(response.Content)
	if !strings.Contains(content, "Utilisateur déconnecté") {
		t.Errorf("\nLogOut[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", 0, "Utilisateur déconnecté", content)
	}
	statusCode := response.Raw().StatusCode
	if statusCode != http.StatusOK {
		t.Errorf("\nLogOut[%d],statut :  attendu ->%v  reçu <-%v", 0, http.StatusOK, statusCode)
	}

	response = e.POST("/api/user/logout").WithHeader("Authorization", "Bearer "+testCtx.User.Token).Expect()
	content = string(response.Content)
	if !strings.Contains(content, "Token invalide") {
		t.Errorf("\nLogOut[%d] :\n  attendu ->\"%s\"\n  reçu <-\"%s\"", 1, "Token invalide", content)
	}
	statusCode = response.Raw().StatusCode
	if statusCode != http.StatusInternalServerError {
		t.Errorf("\nLogOut[%d],statut :  attendu ->%v  reçu <-%v", 1, http.StatusInternalServerError, statusCode)
	}
	newLRUser := fetchLoginResponse(e, t, userCredentials, "USER")
	if newLRUser != nil {
		testCtx.User = newLRUser
	}
}
