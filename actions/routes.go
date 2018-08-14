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
	adminParty.Get("/users/{userID:int}/rights", GetRight)
	adminParty.Post("/users/{userID:int}/rights", SetRight)
	adminParty.Post("/users/{userID:int}/inherits", InheritRight)

	adminParty.Post("/physical_ops", CreatePhysicalOp)
	adminParty.Post("/physical_ops/array", BatchPhysicalOps)
	adminParty.Delete("/physical_ops/{opID:int}", DeletePhysicalOp)

	adminParty.Get("/beneficiaries", GetBeneficiaries)
	adminParty.Put("/beneficiaries/{beneficiaryID:int}", UpdateBeneficiary)

	adminParty.Get("/budget_chapters", GetBudgetChapters)
	adminParty.Post("/budget_chapters", CreateBudgetChapter)
	adminParty.Put("/budget_chapters/{bcID:int}", ModifyBudgetChapter)
	adminParty.Delete("/budget_chapters/{bcID:int}", DeleteBudgetChapter)

	adminParty.Post("/budget_sectors", CreateBudgetSector)
	adminParty.Put("/budget_sectors/{bsID:int}", ModifyBudgetSector)
	adminParty.Delete("/budget_sectors/{bsID:int}", DeleteBudgetSector)

	adminParty.Post("/budget_chapters/{chpID:int}/budget_programs", CreateBudgetProgram)
	adminParty.Put("/budget_chapters/{chpID:int}/budget_programs/{bpID:int}", ModifyBudgetProgram)
	adminParty.Delete("/budget_chapters/{chpID:int}/budget_programs/{bpID:int}", DeleteBudgetProgram)

	adminParty.Post("/budget_credits", CreateBudgetCredit)
	adminParty.Put("/budget_credits/{brID:int}", ModifyBudgetCredit)
	adminParty.Post("/budget_credits/array", BatchBudgetCredits)
	adminParty.Delete("/budget_credits/{brID:int}", DeleteBudgetCredit)

	adminParty.Get("/budget_chapters/{chpID:int}/budget_programs/{prgID:int}/budget_actions", GetProgramBudgetActions)
	adminParty.Post("/budget_chapters/{chpID:int}/budget_programs/{prgID:int}/budget_actions", CreateBudgetAction)
	adminParty.Post("/budget_actions", BatchBudgetActions)
	adminParty.Put("/budget_chapters/{chpID:int}/budget_programs/{prgID:int}/budget_actions/{baID:int}", ModifyBudgetAction)
	adminParty.Delete("/budget_chapters/{chpID:int}/budget_programs/{prgID:int}/budget_actions/{baID:int}", DeleteBudgetAction)

	adminParty.Get("/categories", GetCategories)
	adminParty.Post("/categories", CreateCategory)
	adminParty.Put("/categories/{caID:int}", ModifyCategory)
	adminParty.Delete("/categories/{caID:int}", DeleteCategory)

	adminParty.Post("/commissions", CreateCommission)
	adminParty.Put("/commissions/{coID:int}", ModifyCommission)
	adminParty.Delete("/commissions/{coID:int}", DeleteCommission)

	userParty := api.Party("", ActiveMiddleware)
	userParty.Post("/logout", Logout) // change, before located at /user/logout
	userParty.Get("/physical_ops", GetPhysicalOps)
	userParty.Put("/physical_ops/{opID:int}", UpdatePhysicalOp)
	userParty.Post("/user/password", ChangeUserPwd)

	userParty.Get("/budget_actions", GetAllBudgetActions)

	userParty.Get("/budget_credits/year", GetLastBudgetCredits)
	userParty.Get("/budget_credits", GetBudgetCredits)

	userParty.Get("/budget_programs", GetAllBudgetPrograms)
	userParty.Get("/budget_chapters/{chpID:int}/budget_programs", GetChapterBudgetPrograms)

	userParty.Get("/budget_sectors", GetBudgetSectors)

	userParty.Get("/commissions", GetCommissions)

	userParty.Get("/physical_ops/{opID:int}/documents", GetDocuments)
	userParty.Post("/physical_ops/{opID:int}/documents", CreateDocument)
	userParty.Put("/physical_ops/{opID:int}/documents/{doID:int}", ModifyDocument)
	userParty.Delete("/physical_ops/{opID:int}/documents/{doID:int}", DeleteDocument)

}

// setDBMiddleware return a middleware to add db to context values
func setDBMiddleware(db *gorm.DB) func(iris.Context) {
	return func(ctx iris.Context) {
		ctx.Values().Set("db", db)
		ctx.Next()
	}
}
