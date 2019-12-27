package actions

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/Iledant/iris-propera/config"
	"github.com/iris-contrib/httpexpect"
)

// UserTest includes all tests for users
func testUser(t *testing.T, userCredentials *config.Credentials) {
	t.Run("User", func(t *testing.T) {
		getUsers(testCtx.E, t)
		createdUID := createUser(testCtx.E, t)
		if createdUID == "0" {
			t.Fatal("Impossible de créer l'utilisateur")
		}
		updateUser(testCtx.E, t, createdUID)
		chgPwd(testCtx.E, t)
		deleteUser(testCtx.E, t, createdUID)
		signupTest(testCtx.E, t)
		logoutTest(testCtx.E, t, userCredentials)
	})
}

// getUsers test list returned
func getUsers(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Token:         testCtx.Admin.Token,
			Status:        http.StatusOK,
			BodyContains:  []string{`"User":[`},
			CountItemName: `"id"`,
			ArraySize:     43,
		},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.GET("/api/user").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "GetUsers") {
		t.Error(r)
	}
}

//createUser test user creation and get userID for other tests
func createUser(e *httpexpect.Expect, t *testing.T) string {
	testCases := []testCase{
		notAdminTestCase,
		{
			Sent: []byte(`{"name":"Essai6","email":"essai5@iledefrance.fr",` +
				`"password":"toto","role":"USER","active":false}`),
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			BodyContains: []string{"existant"}},
		{
			Sent: []byte(`{"name":"","email":"essai@iledefrance.fr","password":` +
				`"toto","role":"USER","active":false}`),
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			BodyContains: []string{"Champ manquant ou incorrect"}},
		{
			Sent: []byte(`{"name":"essai4","email":"essai@iledefrance.fr",` +
				`"password":"toto","role":"FALSE","active":false}`),
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			BodyContains: []string{"Champ manquant ou incorrect"}},
		{
			Sent: []byte(`{"name":"essai","email":"essai@iledefrance.fr",` + `
			"password":"toto","role":"USER","active":false}`),
			Token:        testCtx.Admin.Token,
			Status:       http.StatusCreated,
			IDName:       `"id"`,
			BodyContains: []string{"essai"}},
	}
	var userID int
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/user").WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "CreateUser", &userID) {
		t.Error(r)
	}
	return strconv.Itoa(userID)
}

// updateUser tests changing properties of previously created user
func updateUser(e *httpexpect.Expect, t *testing.T, createdUID string) {
	testCases := []testCase{
		notAdminTestCase,
		{
			Sent:         []byte(`{"name":"","email":"","password":"","role":"","active":false}`),
			Token:        testCtx.Admin.Token,
			Status:       http.StatusInternalServerError,
			ID:           "0",
			BodyContains: []string{"Modification d'utilisateur, requête get"}},
		{
			Sent:         []byte(`{"name":"","email":"","password":"","role":"","active":false}`),
			Token:        testCtx.Admin.Token,
			Status:       http.StatusOK,
			ID:           createdUID,
			BodyContains: []string{"essai"}},
		{
			Sent:         []byte(`{"name":"Essai2","email":"","password":"","role":"","active":true}`),
			Token:        testCtx.Admin.Token,
			Status:       http.StatusOK,
			ID:           createdUID,
			BodyContains: []string{`"name":"Essai2"`}},
		{
			Sent:         []byte(`{"name":"","email":"essai2@iledefrance.fr","password":"","role":"","active":true}`),
			Token:        testCtx.Admin.Token,
			Status:       http.StatusOK,
			ID:           createdUID,
			BodyContains: []string{`"email":"essai2@iledefrance.fr"`}},
		{
			Sent:         []byte(`{"name":"","email":"","password":"","role":"ADMIN","active":true}`),
			Token:        testCtx.Admin.Token,
			Status:       http.StatusOK,
			ID:           createdUID,
			BodyContains: []string{`"role":"ADMIN"`}},
		{
			Sent:         []byte(`{"name":"","email":"","password":"","role":"FAIL","active":true}`),
			Token:        testCtx.Admin.Token,
			Status:       http.StatusBadRequest,
			ID:           createdUID,
			BodyContains: []string{"Modification d'utilisateur, rôle incorrect"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.PUT("/api/user/"+tc.ID).WithHeader("Authorization", "Bearer "+tc.Token).
			WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "UpdateUser") {
		t.Error(r)
	}
}

// chgPwd tests the request for the connected user
func chgPwd(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{
			Token:        testCtx.User.Token,
			Sent:         []byte("current_password=fake&password=tutu"),
			Status:       http.StatusBadRequest,
			BodyContains: []string{"Erreur de mot de passe"}},
		{
			Token:        testCtx.User.Token,
			Sent:         []byte("current_password=toto&password=tutu"),
			Status:       http.StatusOK,
			BodyContains: []string{"Mot de passe changé"}},
		{
			Token:        testCtx.User.Token,
			Sent:         []byte("current_password=tutu&password=toto"),
			Status:       http.StatusOK,
			BodyContains: []string{"Mot de passe changé"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/user/password").WithHeader("Authorization", "Bearer "+tc.Token).
			WithQueryString(string(tc.Sent)).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "ChangePassword") {
		t.Error(r)
	}
}

// deleteUser test admin deleting of an user
func deleteUser(e *httpexpect.Expect, t *testing.T, createdUID string) {
	testCases := []testCase{
		{
			Token:        testCtx.User.Token,
			ID:           createdUID,
			Status:       http.StatusUnauthorized,
			BodyContains: []string{"Droits administrateur requis"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           "",
			Status:       http.StatusNotFound,
			BodyContains: []string{""}},
		{
			Token:        testCtx.Admin.Token,
			ID:           "0",
			Status:       http.StatusInternalServerError,
			BodyContains: []string{"Utilisateur introuvable"}},
		{
			Token:        testCtx.Admin.Token,
			ID:           createdUID,
			Status:       http.StatusOK,
			BodyContains: []string{"Utilisateur supprimé"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.DELETE("/api/user/"+tc.ID).
			WithHeader("Authorization", "Bearer "+tc.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "DeleteUser") {
		t.Error(r)
	}
}

// signupTest check functionality for a user signing up
func signupTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []testCase{
		{
			Sent:         []byte(`{"Name": "Nouveau", "Email": "", "Password": ""}`),
			Status:       http.StatusBadRequest,
			BodyContains: []string{"Inscription d'utilisateur : nom, email ou mot de passe manquant"}},
		{
			Sent:         []byte(`{"Name": "Nouveau", "Email": "nouveau@iledefrance.fr", "Password": ""}`),
			Status:       http.StatusBadRequest,
			BodyContains: []string{"Inscription d'utilisateur : nom, email ou mot de passe manquant"}},
		{
			Sent:         []byte(`{"Name": "Nouveau", "Email": "essai5@iledefrance.fr", "Password": "nouveau"}`),
			Status:       http.StatusBadRequest,
			BodyContains: []string{"Utilisateur existant"}},
		{
			Sent:         []byte(`{"Name": "Nouveau", "Email": "nouveau@iledefrance.fr", "Password": "nouveau"}`),
			Status:       http.StatusCreated,
			BodyContains: []string{"Utilisateur créé"}},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/user/signup").WithBytes(tc.Sent).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "SignUp") {
		t.Error(r)
	}
}

// logoutTest for a connected user
func logoutTest(e *httpexpect.Expect, t *testing.T, userCredentials *config.Credentials) {
	testCases := []testCase{
		{
			Token:        testCtx.User.Token,
			BodyContains: []string{"Utilisateur déconnecté"},
			Status:       http.StatusOK,
		},
		{
			Token:        testCtx.User.Token,
			BodyContains: []string{"Token invalide"},
			Status:       http.StatusInternalServerError,
		},
	}
	f := func(tc testCase) *httpexpect.Response {
		return e.POST("/api/user/logout").
			WithHeader("Authorization", "Bearer "+testCtx.User.Token).Expect()
	}
	for _, r := range chkTestCases(testCases, f, "LogOut") {
		t.Error(r)
	}

	newLRUser := fetchLoginResponse(e, t, userCredentials, "USER")
	if newLRUser != nil {
		testCtx.User = newLRUser
	}
}
