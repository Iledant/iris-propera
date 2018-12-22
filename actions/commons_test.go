package actions

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/Iledant/iris_propera/config"
	"github.com/Iledant/iris_propera/models"
	"github.com/iris-contrib/httpexpect"
	"github.com/jinzhu/gorm"

	"github.com/kataras/iris"
	"github.com/kataras/iris/httptest"
)

// testCase is the common structure for all case
type testCase struct {
	Token        string
	Status       int
	ID           string
	Param        string
	BodyContains []string
	Sent         []byte
	ArraySize    int
}

// TestContext contains all items for units tests in API.
type TestContext struct {
	DB     *gorm.DB
	App    *iris.Application
	E      *httpexpect.Expect
	Admin  *LoginResponse
	User   *LoginResponse
	Config *config.ProperaConf
}

// LoginResponse contains the response of a login i.e. token and most of users fields
type LoginResponse struct {
	Token string
	User  models.User
}

// Credentials are used for loging in
type Credentials struct {
	Email, Password string
}

var testCtx *TestContext

// Init initialize the database for testing by creating a test database, connecting to it and launching
func TestCommons(t *testing.T) {
	mutex := &sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()
	if testCtx == nil {
		restoreTestDB(t)

		app := iris.New().Configure(iris.WithConfiguration(iris.Configuration{DisablePathCorrection: true}))
		cfg := config.Get()
		if cfg == nil {
			t.Error("Impossible de récupérer la configuration")
			t.FailNow()
		}

		db, err := config.LaunchDB(&cfg.Databases.Test)
		if err != nil {
			t.Errorf("Erreur de connexion à postgres : %v\n", err.Error())
			t.FailNow()
		}

		SetRoutes(app, db)
		e := httptest.New(t, app)
		admin := fetchLoginResponse(e, t, &cfg.Users.Admin, "ADMIN")
		user := fetchLoginResponse(e, t, &cfg.Users.User, "USER")

		t := TestContext{DB: db, App: app, E: e, Admin: admin, User: user, Config: cfg}
		testCtx = &t
	}
}

// restoreTestDB executes the pg_restore command to restore a new database test. testing.FailNow is fired if an error happens.
func restoreTestDB(t *testing.T) {
	properaRep, ok := os.LookupEnv("PROPERAREPO")
	if !ok {
		t.Error("Variable PROPERAREPO introuvable")
		t.FailNow()
	}

	if _, ok = os.LookupEnv("PGPASSWORD"); !ok {
		t.Error("Variable PGPASSWORD introuvable")
		t.FailNow()
	}

	cmd := exec.Command("pg_restore", "-cOU", "postgres", "-d", "propera3_test", properaRep)
	s, err := cmd.CombinedOutput()
	if err != nil && strings.Contains(string(s), "FATAL") {
		t.Errorf("Impossible de restaurer la base de test:\n%s\n", string(s))
		t.FailNow()
	}
}

// fetchTokens logins an user and send back the login response (token and user fiels)
func fetchLoginResponse(e *httpexpect.Expect, t *testing.T, c *config.Credentials, role string) *LoginResponse {
	response := e.POST("/api/user/signin").WithBytes([]byte(`{"email":"` + c.Email + `","password":"` + c.Password + `"}`)).Expect()

	lr := LoginResponse{}
	if err := json.Unmarshal(response.Content, &lr); err != nil {
		t.Errorf("Impossible de décoder la réponse du login %s sur réponse %s\n", role, string(response.Content))
		t.FailNow()
		return nil
	}
	response.Status(http.StatusOK).Body().Contains("token").Contains(role)

	return &lr
}
