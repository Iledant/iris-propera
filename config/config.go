package config

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/jinzhu/gorm"
	// Only place where the postgres is initialize to avoid duplication in tests
	_ "github.com/jinzhu/gorm/dialects/postgres"
	yaml "gopkg.in/yaml.v2"
)

// ProperaConf includes all configuration datas from config.yml for production, development and tests.
type ProperaConf struct {
	Databases Databases
	Users     Users
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

// Get fetches all parameters according to tne context : if proper environment variables are set, assumes beeing in prod, otherwise read the config.yml file
func (p *ProperaConf) Get() error {
	if config == nil {
		// Check if RDS environment variables are set otherwise use database.yml
		name, okDbName := os.LookupEnv("RDS_DB_NAME")
		host, okHostName := os.LookupEnv("RDS_HOSTNAME")
		port, okPort := os.LookupEnv("RDS_PORT")
		username, okUserName := os.LookupEnv("RDS_USERNAME")
		password, okPwd := os.LookupEnv("RDS_PASSWORD")

		if okDbName && okHostName && okPort && okUserName && okPwd {
			p = &ProperaConf{Databases: Databases{Prod: DBConf{
				Name:     name,
				Host:     host,
				Port:     port,
				UserName: username,
				Password: password}}}
			return nil
		}

		cfgFile, err := ioutil.ReadFile("../config.yml")
		if err != nil {
			// Try to read directly
			cfgFile, err = ioutil.ReadFile("config.yml")
			if err != nil {
				return errors.New("Erreur lors de la lecture de config.yml : " + err.Error())
			}
		}
		if err = yaml.Unmarshal(cfgFile, p); err != nil {
			return errors.New("Erreur lors du décodage de config.yml : " + err.Error())
		}
	} else {
		p = config
	}
	return nil
}

// LaunchDB launch the DB with DBConf parameters
func LaunchDB(cfg *DBConf) (*gorm.DB, error) {
	return gorm.Open("postgres", "sslmode=disable host="+cfg.Host+" port="+cfg.Port+
		" user="+cfg.UserName+" dbname="+cfg.Name+" password= "+cfg.Password)
}
