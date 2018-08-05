package actions

import (
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// SetRoutes initialize all routes for the application
func SetRoutes(app *iris.Application, db *gorm.DB) {
	app.Post("/users/signup", setDBMiddleware(db), SignUp)
	app.Post("/users/signin", setDBMiddleware(db), Login)
	api := app.Party("/api", setDBMiddleware(db))
	adminParty := api.Party("", AdminMiddleware)
	adminParty.Get("/users", GetUsers)
	adminParty.Post("/users", CreateUser)
	adminParty.Put("/users/{userID:int}", UpdateUser)
	adminParty.Delete("/users/{userID:int}", DeleteUser)
	userParty := api.Party("", ActiveMiddleware)
	userParty.Post("/logout", Logout) // change, before located at /user/logout
	userParty.Post("/user/password", ChangeUserPwd)
}

// setDBMiddleware return a middleware to add db to context values
func setDBMiddleware(db *gorm.DB) func(iris.Context) {
	return func(ctx iris.Context) {
		ctx.Values().Set("db", db)
		ctx.Next()
	}
}
