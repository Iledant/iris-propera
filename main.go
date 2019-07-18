package main

import (
	stdContext "context"
	"log"
	"os"
	"time"

	"github.com/Iledant/iris_propera/actions"
	"github.com/Iledant/iris_propera/config"

	"github.com/kataras/iris"
)

func main() {
	app := iris.New().Configure(
		iris.WithConfiguration(iris.Configuration{DisablePathCorrection: true}))

	var cfg config.ProperaConf
	if err := cfg.Get(); err != nil {
		log.Fatal("Configuration : " + err.Error())
	}

	db, err := config.LaunchDB(&cfg.Databases.Development)
	if err != nil {
		log.Printf("Impossible de se connecter à la base de données : %s", err.Error())
		os.Exit(1)
	}
	defer db.Close()
	actions.SetRoutes(app, db)
	if cfg.App.LoggerLevel != "" {
		app.Logger().SetLevel(cfg.App.LoggerLevel)
	}

	if cfg.App.TokenFileName != "" {
		actions.TokenRecover(cfg.App.TokenFileName)
		iris.RegisterOnInterrupt(func() {
			timeout := 2 * time.Second
			ctx, cancel := stdContext.WithTimeout(stdContext.Background(), timeout)
			defer cancel()
			actions.TokenSave(cfg.App.TokenFileName)
			app.Shutdown(ctx)
		})
	}
	// Use port 5000 as Elastic beanstalk use it by default
	app.Run(iris.Addr(":5000"), iris.WithoutInterruptHandler)
}
