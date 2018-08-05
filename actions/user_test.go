package actions

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/Iledant/iris_propera/config"
	"github.com/Iledant/iris_propera/models"
	"github.com/iris-contrib/httpexpect"
)

type SentUser struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role,omitempty"`
	Active   bool   `json:"active,omitempty"`
}

type jsonUser struct {
	User struct {
		ID       int             `json:"id"`
		Name     string          `json:"name"`
		Email    string          `json:"email"`
		Password string          `json:"password"`
		Role     string          `json:"role"`
		Active   bool            `json:"active"`
		Created  models.NullTime `json:"created_at"`
		Updated  models.NullTime `json:"updated_at"`
	}
}

type SentUserCase struct {
	User         SentUser
	Status       int
	BodyContains string
	Describe     string
}

// Stored for all tests
var createdID int
var adminToken, userToken string

// UserTest includes all tests for users
func TestUser(t *testing.T) {
	TestCommons(t)
	adminToken = testCtx.Admin.Token
	userToken = testCtx.User.Token
	t.Run("User", func(t *testing.T) {
		getUsers(testCtx.E, t)
		createUser(testCtx.E, t)
		updateUser(testCtx.E, t)
		chgPwd(testCtx.E, t)
		deleteUser(testCtx.E, t)
		signupTest(testCtx.E, t)
		logoutTest(testCtx.E, t)
	})
}

// getUsers test list returned
func getUsers(e *httpexpect.Expect, t *testing.T) {
	response := e.GET("/api/users").WithHeader("Authorization", "Bearer "+adminToken).
		Expect()
	response.Status(http.StatusOK)
	response.JSON().Object().ContainsKey("user")
	e.GET("/api/users").WithHeader("Authorization", "Bearer "+userToken).
		Expect().Status(http.StatusUnauthorized)
}

//createUser test user creation and get userID for other tests
func createUser(e *httpexpect.Expect, t *testing.T) {
	cts := []SentUserCase{
		{User: SentUser{Name: "Essai6", Email: "essai5@iledefrance.fr", Password: "toto", Role: "USER", Active: false},
			Status: http.StatusBadRequest, BodyContains: "existant", Describe: "Email déjà présent"},
		{User: SentUser{Name: "", Email: "essai@iledefrance.fr", Password: "toto", Role: "USER", Active: false},
			Status: http.StatusBadRequest, BodyContains: "Champ manquant ou incorrect"},
		{User: SentUser{Name: "essai4", Email: "essai@iledefrance.fr", Password: "toto", Role: "FALSE", Active: false},
			Status: http.StatusBadRequest, BodyContains: "Champ manquant ou incorrect"},
		{User: SentUser{Name: "essai", Email: "essai@iledefrance.fr", Password: "toto", Role: "USER", Active: false},
			Status: http.StatusCreated, BodyContains: "essai"},
	}

	var response *httpexpect.Response
	for _, ct := range cts {
		response = e.POST("/api/users").WithHeader("Authorization", "Bearer "+adminToken).WithJSON(ct.User).Expect()
		response.Body().Contains(ct.BodyContains)
		response.Status(ct.Status)
	}

	createdUser := jsonUser{}
	if err := json.Unmarshal(response.Content, &createdUser); err != nil {
		t.Error("Impossible de décoder la réponse de l'utilisateur créé")
		t.FailNow()
	}
	createdID = createdUser.User.ID
}

// updateUser tests changing properties of previously created user
func updateUser(e *httpexpect.Expect, t *testing.T) {
	cts := []SentUserCase{
		{User: SentUser{Name: "", Email: "", Password: "", Role: "", Active: false},
			Status: http.StatusOK, BodyContains: "essai"},
		{User: SentUser{Name: "Essai2", Email: "", Password: "", Role: "", Active: true},
			Status: http.StatusOK, BodyContains: `"name":"Essai2"`},
		{User: SentUser{Name: "", Email: "essai2@iledefrance.fr", Password: "", Role: "", Active: true},
			Status: http.StatusOK, BodyContains: `"email":"essai2@iledefrance.fr"`},
		{User: SentUser{Name: "", Email: "", Password: "", Role: "ADMIN", Active: true},
			Status: http.StatusOK, BodyContains: `"role":"ADMIN"`},
		{User: SentUser{Name: "", Email: "", Password: "", Role: "FAIL", Active: true},
			Status: http.StatusBadRequest, BodyContains: "Rôle différent"},
	}

	for _, ct := range cts {
		response := e.PUT("/api/users/"+strconv.Itoa(createdID)).
			WithHeader("Authorization", "Bearer "+adminToken).WithJSON(ct.User).Expect()
		response.Status(ct.Status).Body().Contains(ct.BodyContains)
	}

	e.PUT("/post/users/0").WithHeader("Authorization", "Bearer "+adminToken).
		Expect().Status(http.StatusNotFound)
}

// chgPwd tests the request for the connected user
func chgPwd(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Old, New, BodyContains string
		Status                 int
	}{{"fake", "tutu", "Erreur de mot de passe", http.StatusBadRequest},
		{"toto", "tutu", "Mot de passe changé", http.StatusOK},
		{"tutu", "toto", "Mot de passe changé", http.StatusOK}}

	for _, tc := range testCases {
		response := e.POST("/api/user/password").WithQuery("current_password", tc.Old).WithQuery("password", tc.New).
			WithHeader("Authorization", "Bearer "+userToken).Expect()
		response.Body().Contains(tc.BodyContains)
		response.Status(tc.Status)
	}
}

// deleteUser test admin deleting of an user
func deleteUser(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Token, UserID string
		Status        int
		BodyContains  string
	}{
		{Token: userToken, UserID: strconv.Itoa(createdID), Status: http.StatusUnauthorized, BodyContains: "Droits administrateurs requis"},
		{Token: adminToken, UserID: "", Status: http.StatusNotFound, BodyContains: ""},
		{Token: adminToken, UserID: strconv.Itoa(0), Status: http.StatusNotFound, BodyContains: "Utilisateur introuvable"},
		{Token: adminToken, UserID: strconv.Itoa(createdID), Status: http.StatusOK, BodyContains: "Utilisateur supprimé"},
	}

	for _, tc := range testCases {
		request := e.DELETE("/api/users/"+tc.UserID).WithHeader("Authorization", "Bearer "+tc.Token).Expect()
		if tc.BodyContains != "" {
			request.Body().Contains(tc.BodyContains)
		}
		request.Status(tc.Status)
	}
}

// signupTest check functionality for a user signing up
func signupTest(e *httpexpect.Expect, t *testing.T) {
	testCases := []struct {
		Name, Email, Password string
		Status                int
		BodyContains          string
	}{
		{Name: "Nouveau", Email: "", Password: "", Status: http.StatusBadRequest, BodyContains: "Champ manquant ou incorrect"},
		{Name: "Nouveau", Email: "nouveau@iledefrance.fr", Password: "", Status: http.StatusBadRequest, BodyContains: "Champ manquant ou incorrect"},
		{Name: "Nouveau", Email: "essai5@iledefrance.fr", Password: "nouveau", Status: http.StatusBadRequest, BodyContains: "Utilisateur existant"},
		{Name: "Nouveau", Email: "nouveau@iledefrance.fr", Password: "nouveau", Status: http.StatusCreated, BodyContains: "Utilisateur créé"},
	}

	for _, tc := range testCases {
		request := e.POST("/users/signup").WithQuery("name", tc.Name).WithQuery("email", tc.Email).WithQuery("password", tc.Password).Expect()
		request.Body().Contains(tc.BodyContains)
		request.Status(tc.Status)
	}
}

// logoutTest for a connected user
func logoutTest(e *httpexpect.Expect, t *testing.T) {
	request := e.POST("/api/logout").WithHeader("Authorization", "Bearer "+userToken).Expect()
	request.Body().Contains("Utilisateur déconnecté")
	request.Status(http.StatusOK)

	request = e.POST("/api/logout").WithHeader("Authorization", "Bearer "+userToken).Expect()
	request.Body().Contains("Token invalide")
	request.Status(http.StatusInternalServerError)

	cfg := config.Get()
	newLRUser := fetchLoginResponse(e, t, &cfg.Users.User, "USER")
	if newLRUser != nil {
		testCtx.User = newLRUser
	}
}
