package config

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/kataras/iris"

	// Imported in config to avoid double import
	_ "github.com/lib/pq"

	yaml "gopkg.in/yaml.v2"
)

// ProperaConf includes all configuration datas from config.yml for production, development and tests.
type ProperaConf struct {
	Databases Databases
	Users     Users
	App       App
}

// Users includes users credentials for test purposes.
type Users struct {
	Admin Credentials
	User  Credentials
}

// Databases includes the 3 databases settings for production, development and tests.
type Databases struct {
	Prod        DBConf
	Development DBConf
	Test        DBConf
}

// App defines global values for the application
type App struct {
	Prod          bool
	LogFileName   string
	LoggerLevel   string
	TokenFileName string
}

// DBConf includes all informations for connecting to a database.
type DBConf struct {
	Name       string `yaml:"name"`
	Host       string `yaml:"host"`
	Port       string `yaml:"port"`
	UserName   string `yaml:"username"`
	Password   string `yaml:"password"`
	Repository string `yaml:"repository"`
	RestoreCmd string `yaml:"restoreCmd"`
}

// Credentials keep email ans password for a user.
type Credentials struct {
	Email, Password string
}

var config *ProperaConf

func logFileOpen(name string, app *iris.Application) (*os.File, error) {
	logFile, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	app.Logger().SetOutput(logFile)
	app.Logger().Infof("Fichier log configuré")
	return logFile, err
}

// Get fetches all parameters according to tne context : if proper environment
// variables are set, assumes beeing in prod, otherwise read the config.yml file
func (p *ProperaConf) Get(app *iris.Application) (logFile *os.File, err error) {
	if config != nil {
		p = config
		return nil, nil
	}

	// Configure the log file as first step to catch all messages
	p.App.LogFileName = os.Getenv("LOG_FILE_NAME")
	if p.App.LogFileName != "" {
		logFile, err = logFileOpen(p.App.LogFileName, app)
		if err != nil {
			return nil, err
		}
	}

	// Check if RDS environment variables are set
	name, okDbName := os.LookupEnv("RDS_DB_NAME")
	host, okHostName := os.LookupEnv("RDS_HOSTNAME")
	port, okPort := os.LookupEnv("RDS_PORT")
	username, okUserName := os.LookupEnv("RDS_USERNAME")
	password, okPwd := os.LookupEnv("RDS_PASSWORD")

	if okDbName && okHostName && okPort && okUserName && okPwd {
		app.Logger().Infof("Utilisation des variables d'environnement")
		p.Databases.Prod.Name = name
		p.Databases.Prod.Host = host
		p.Databases.Prod.Port = port
		p.Databases.Prod.UserName = username
		p.Databases.Prod.Password = password
		p.App.TokenFileName = os.Getenv("TOKEN_FILE_NAME")
		p.App.Prod = true
		p.App.LoggerLevel = "info"
		return logFile, nil
	}
	// Otherwise use database.yml
	cfgFile, err := ioutil.ReadFile("../config.yml")
	if err != nil {
		// Try to read directly
		cfgFile, err = ioutil.ReadFile("config.yml")
		if err != nil {
			return nil, fmt.Errorf("Erreur de lecture de config.yml : %v", err)
		}
	}
	if err = yaml.Unmarshal(cfgFile, p); err != nil {
		return nil, fmt.Errorf("Erreur lors du décodage de config.yml : %v", err)
	}
	if p.App.LoggerLevel != "" {
		app.Logger().SetLevel(p.App.LoggerLevel)
	}
	if logFile == nil && p.App.LogFileName != "" {
		logFile, err = logFileOpen(p.App.LogFileName, app)
	}
	app.Logger().Infof("Utilisation de config.yml")
	return logFile, nil
}

type mig struct {
	Batch int64
	Query string
}

var migrations = []mig{
	{
		Batch: 21,
		Query: `ALTER TABLE financial_commitment ADD COLUMN app boolean DEFAULT false`},
	{
		Batch: 22,
		Query: `ALTER TABLE temp_commitment ADD COLUMN app boolean`},
	{
		Batch: 23,
		Query: `update financial_commitment set coriolis_year='2019',coriolis_egt_code='IRIS',
  coriolis_egt_num='609297',coriolis_egt_line='1'  where id=4695`},
	{
		Batch: 24,
		Query: `update financial_commitment set coriolis_egt_num='609307', coriolis_year='2019' 
  where id=4697`},
	{
		Batch: 25,
		Query: `update financial_commitment set coriolis_egt_num='609308', coriolis_year='2019' 
  where id=4699`},
	{
		Batch: 26,
		Query: `update financial_commitment set coriolis_egt_num='609309', coriolis_year='2019' 
  where id=4701`},
	{
		Batch: 27,
		Query: `update financial_commitment set coriolis_egt_num='604865', coriolis_year='2019' 
  where id=4678`},
	{
		Batch: 28,
		Query: `CREATE EXTENSION IF NOT EXISTS fuzzystrmatch`},
	{
		Batch: 29,
		Query: `ALTER TABLE payment ADD COLUMN receipt_date date`},
	{
		Batch: 30,
		Query: `ALTER TABLE temp_payment ADD COLUMN receipt_date date`},
	{
		Batch: 31,
		Query: `ALTER TABLE temp_commitment ADD COLUMN op_name varchar(250)`},
	{
		Batch: 32,
		Query: `CREATE TABLE IF NOT EXISTS temp_payment_demands (
			iris_code varchar(32) NOT NULL,
			iris_name varchar(200) NOT NULL,
			commitment_date date NOT NULL,
			beneficiary_code int NOT NULL,
			demand_number int NOT NULL,
			demand_date	date NOT NULL,
			receipt_date date NOT NULL,
			demand_value bigint NOT NULL,
			csf_date date,
			csf_comment text,
			demand_status varchar(15),
			status_comment text
		)`},
	{
		Batch: 33,
		Query: `CREATE OR REPLACE VIEW imported_payment_demands AS
			SELECT iris_code,iris_name,MAX(commitment_date),beneficiary_code,
				demand_number,demand_date,receipt_date,demand_value,csf_date,csf_comment,
				demand_status,status_comment FROM temp_payment_demands
				GROUP BY 1,2,4,5,6,7,8,9,10,11,12`},
	{
		Batch: 34,
		Query: `CREATE TABLE IF NOT EXISTS payment_demands (
			id SERIAL PRIMARY KEY,
			import_date date NOT NULL,
			iris_code varchar(32) NOT NULL,
			iris_name varchar(200) NOT NULL,
			beneficiary_id int NOT NULL REFERENCES beneficiary(id),
			demand_number int NOT NULL,
			demand_date	date NOT NULL,
			receipt_date date NOT NULL,
			demand_value bigint NOT NULL,
			csf_date date,
			csf_comment text,
			demand_status varchar(15),
			status_comment text,
			excluded boolean,
			excluded_comment varchar(150),
			processed_date date
		)`},
	{
		Batch: 35,
		Query: `ALTER TABLE payment_demands 
			ALTER excluded SET DEFAULT FALSE,
			ALTER excluded SET NOT NULL`},
}

// handleMigrations checks against database if migrations queries must be executed
func handleMigrations(db *sql.DB) error {
	var bMax int64
	err := db.QueryRow(`SELECT max(batch) FROM migrations`).Scan(&bMax)
	if err != nil {
		return err
	}
	for _, b := range migrations {
		if b.Batch > bMax {
			if _, err = db.Exec(b.Query); err != nil {
				return fmt.Errorf("migration batch %d %v", b.Batch, err)
			}
			if _, err = db.Exec(`INSERT INTO migrations (migration,batch) VALUES($1,$2)`,
				fmt.Sprintf("%s", time.Now().Format("2006-01-02-150405")),
				b.Batch); err != nil {
				return fmt.Errorf("migration insert %d %v", b.Batch, err)
			}
		}
	}
	return err
}

// LaunchDB launch the DB with DBConf parameters
func LaunchDB(cfg *DBConf) (*sql.DB, error) {
	cfgStr := fmt.Sprintf(
		"sslmode=disable host=%s port=%s user=%s dbname=%s password=%s",
		cfg.Host, cfg.Port, cfg.UserName, cfg.Name, cfg.Password)
	db, err := sql.Open("postgres", cfgStr)
	if err != nil {
		return nil, err
	}
	err = handleMigrations(db)
	return db, err
}
