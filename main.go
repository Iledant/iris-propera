package main

import (
	"log"
	"os"

	"github.com/Iledant/iris_propera/actions"
	"github.com/Iledant/iris_propera/config"

	"github.com/kataras/iris"
)

func main() {
	app := iris.New().Configure(iris.WithConfiguration(iris.Configuration{DisablePathCorrection: true}))
	app.Logger().SetLevel("debug")

	cfg := config.Get()
	if cfg == nil {
		log.Print("Erreur ===> impossible de récupérer la configuration")
		os.Exit(1)
	}

	db, err := config.LaunchDB(&cfg.Databases.Development)
	if err != nil {
		log.Printf("Erreur ===> impossible de se connecter à la base de données : %s", err.Error())
		os.Exit(1)
	}

	defer db.Close()

	actions.SetRoutes(app, db)

	// Use port 5000 as Elastic beanstalk use it by default
	app.Run(iris.Addr(":5000"))
}
