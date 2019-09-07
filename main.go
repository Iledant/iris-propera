package main

import (
	stdContext "context"
	"log"
	"os"
	"time"

	"github.com/Iledant/iris-propera/actions"
	"github.com/Iledant/iris-propera/config"

	"github.com/kataras/iris"
)

func main() {
	app := iris.New().Configure(
		iris.WithConfiguration(iris.Configuration{DisablePathCorrection: true}))

	var cfg config.ProperaConf
	logFile, err := cfg.Get(app)
	if logFile != nil {
		defer logFile.Close()
	}
	if err != nil {
		log.Fatal("Configuration : " + err.Error())
	}
	var dbConf *config.DBConf
	if cfg.App.Prod {
		dbConf = &cfg.Databases.Prod
	} else {
		dbConf = &cfg.Databases.Development
	}

	db, err := config.LaunchDB(dbConf)
	if err != nil {
		log.Printf("Impossible de se connecter à la base de données : %v", err)
		os.Exit(1)
	}
	app.Logger().Infof("Base de données connectée et initialisée")
	defer db.Close()

	actions.SetRoutes(app, db)
	app.StaticWeb("/", "./dist")
	app.Logger().Infof("Routes et serveur statique configurés")

	if cfg.App.TokenFileName != "" {
		actions.TokenRecover(cfg.App.TokenFileName)
		iris.RegisterOnInterrupt(func() {
			timeout := 2 * time.Second
			ctx, cancel := stdContext.WithTimeout(stdContext.Background(), timeout)
			defer cancel()
			actions.TokenSave(cfg.App.TokenFileName)
			app.Shutdown(ctx)
		})
		app.Logger().Infof("Fichier de sauvegarde des tokens configuré")
	}

	// Use port 5000 as Elastic beanstalk use it by default
	app.Run(iris.Addr(":5000"), iris.WithoutInterruptHandler)
	app.Logger().Fatalf("Erreur de serveur run %v", err)
}
