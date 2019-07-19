package config

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/kataras/iris"

	"github.com/jinzhu/gorm"
	// Only place where the postgres is initialize to avoid duplication in tests
	_ "github.com/jinzhu/gorm/dialects/postgres"
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
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	UserName string `yaml:"username"`
	Password string `yaml:"password"`
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

// Get fetches all parameters according to tne context : if proper environment variables are set, assumes beeing in prod, otherwise read the config.yml file
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
			return nil, errors.New("Erreur lors de la lecture de config.yml : " + err.Error())
		}
	}
	if err = yaml.Unmarshal(cfgFile, p); err != nil {
		return nil, errors.New("Erreur lors du décodage de config.yml : " + err.Error())
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

// LaunchDB launch the DB with DBConf parameters
func LaunchDB(cfg *DBConf) (*gorm.DB, error) {
	return gorm.Open("postgres", "sslmode=disable host="+cfg.Host+" port="+cfg.Port+
		" user="+cfg.UserName+" dbname="+cfg.Name+" password= "+cfg.Password)
}
