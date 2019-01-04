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
	app.Logger().SetLevel("debug")

	var cfg config.ProperaConf
	if err := cfg.Get(); err != nil {
		log.Fatal("Configuration : " + err.Error())
	}

	db, err := config.LaunchDB(&cfg.Databases.Development)
	if err != nil {
		log.Printf("Impossible de se connecter à la base de données : %s", err.Error())
		os.Exit(1)
	}
	actions.TokenRecover(cfg.TokenFileName)
	defer db.Close()
	actions.SetRoutes(app, db)

	iris.RegisterOnInterrupt(func() {
		timeout := 2 * time.Second
		ctx, cancel := stdContext.WithTimeout(stdContext.Background(), timeout)
		defer cancel()
		actions.TokenSave(cfg.TokenFileName)
		app.Shutdown(ctx)
	})
	// Use port 5000 as Elastic beanstalk use it by default
	app.Run(iris.Addr(":5000"), iris.WithoutInterruptHandler)
}
