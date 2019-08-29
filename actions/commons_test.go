package actions

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/Iledant/iris_propera/config"
	"github.com/Iledant/iris_propera/models"
	"github.com/iris-contrib/httpexpect"

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
	DB     *sql.DB
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

func TestAll(t *testing.T) {
	testCommons(t)
	t.Run("Summaries", func(t *testing.T) { testSummaries(t) })
	t.Run("Scenarios", func(t *testing.T) { testScenario(t) })
	t.Run("Others", func(t *testing.T) {
		testBeneficiary(t)
		testBudgetAction(t)
		testBudgetChapter(t)
		testBudgetCredit(t)
		testBudgetProgram(t)
		testBudgetSector(t)
		testCategory(t)
		testCommission(t)
		testDocument(t)
		testEvent(t)
		testFinancialCommitment(t)
		testImportLog(t)
		testOpDptRatio(t)
		testPaymentRatio(t)
		testPaymentType(t)
		testPendingCommitment(t)
		testPayment(t)
		testPhysicalOps(t)
		testPlanLine(t)
		testPlan(t)
		testPreProgramming(t)
		testPrevCommitment(t)
		testProgramming(t)
		testRight(t)
		testSettings(t)
		testStep(t)
		testTodayMessage(t)
		testUser(t, &testCtx.Config.Users.User)
	})
}

// Init initialize the database for testing by creating a test database, connecting to it and launching
func testCommons(t *testing.T) {
	mutex := &sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()
	if testCtx == nil {
		app := iris.New().Configure(iris.WithConfiguration(iris.Configuration{DisablePathCorrection: true}))
		var cfg config.ProperaConf
		if _, err := cfg.Get(app); err != nil {
			t.Errorf("Configuration : %v\n", err)
			t.FailNow()
		}

		restoreTestDB(t, &cfg.Databases.Test)

		db, err := config.LaunchDB(&cfg.Databases.Test)
		if err != nil {
			t.Errorf("Erreur de connexion à postgres : %v\n", err)
			t.FailNow()
		}

		SetRoutes(app, db)
		e := httptest.New(t, app)
		admin := fetchLoginResponse(e, t, &cfg.Users.Admin, "ADMIN")
		user := fetchLoginResponse(e, t, &cfg.Users.User, "USER")

		t := TestContext{DB: db, App: app, E: e, Admin: admin, User: user, Config: &cfg}
		testCtx = &t
	}
}

// restoreTestDB executes the pg_restore command to restore a new database test.
// testing.FailNow is called if an error happens.
func restoreTestDB(t *testing.T, dbCfg *config.DBConf) {
	if dbCfg.UserName == "" || dbCfg.Password == "" || dbCfg.Host == "" ||
		dbCfg.Port == "" || dbCfg.Repository == "" {
		t.Errorf("Erreur de configuration de la base de test %v\n", *dbCfg)
		t.FailNow()
	}
	dbString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbCfg.UserName,
		dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.Name)
	cmd := exec.Command("pg_restore", "-cO", "-d", dbString, dbCfg.Repository)
	s, err := cmd.CombinedOutput()
	if err != nil && strings.Contains(string(s), "FATAL") {
		t.Errorf("Impossible de restaurer la base de test:\n%s\n", string(s))
		t.FailNow()
	}
}

// fetchTokens logins an user and send back the login response (token and user fiels)
func fetchLoginResponse(e *httpexpect.Expect, t *testing.T, c *config.Credentials, role string) *LoginResponse {
	response := e.POST("/api/user/signin").
		WithBytes([]byte(`{"email":"` + c.Email + `","password":"` + c.Password + `"}`)).
		Expect()

	lr := LoginResponse{}
	if err := json.Unmarshal(response.Content, &lr); err != nil {
		t.Errorf("Impossible de décoder la réponse du login %s sur réponse %s\n",
			role, string(response.Content))
		t.FailNow()
		return nil
	}
	response.Status(http.StatusOK).Body().Contains("token").Contains(role)

	return &lr
}
