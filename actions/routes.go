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

	adminParty.Get("/financial_commitments", GetUnlinkedFcs)                      // changed, before financialcommitments
	adminParty.Get("/financial_commitments/linked", GetLinkedFcs)                 // changed, before financialcommitments
	adminParty.Post("/financial_commitments/physical_ops/{opID:int}", LinkFcToOp) // changed, before financialcommitments
	adminParty.Post("/financial_commitments/plan_lines/{plID:int}", LinkFcToPl)   // changed, before financialcommitments
	adminParty.Post("/financial_commitments/unlink", UnlinkFcs)                   // changed, before financialcommitments
	adminParty.Post("/financial_commitments", BatchFcs)                           // changed, before financialcommitments
	adminParty.Post("/financial_commitments/attachments", BatchOpFcs)             // changed, before financialcommitments

	adminParty.Post("/payment_types/{ptID:int}/payment_ratios", SetPtRatios) // changed, put strictly identical to post is no longer implemented
	adminParty.Delete("/payment_types/{ptID:int}/payment_ratios", DeleteRatios)

	adminParty.Post("/payment_types", CreatePaymentType)
	adminParty.Put("/payment_types/{ptID:int}", ModifyPaymentType)
	adminParty.Delete("/payment_types/{ptID:int}", DeletePaymentType)

	adminParty.Get("/pending_commitments/unlinked", GetUnlinkedPendings)
	adminParty.Get("/pending_commitments/linked", GetLinkedPendings)
	adminParty.Post("/pending_commitments/physical_ops/{opID:int}", LinkPcToOp)
	adminParty.Post("/pending_commitments/physical_ops/{opID:int}", LinkPcToOp)
	adminParty.Post("/pending_commitments/unlink", UnlinkPCs)
	adminParty.Post("/pending_commitments", BatchPendings)

	adminParty.Post("/payments", BatchPayments)

	adminParty.Post("/plans/{pID:int}/planlines", CreatePlanLine)
	adminParty.Put("/plans/{pID:int}/planlines/{plID:int}", ModifyPlanLine)
	adminParty.Delete("/plans/{pID:int}/planlines/{plID:int}", DeletePlanLine)

	userParty := api.Party("", ActiveMiddleware)
	userParty.Post("/logout", Logout) // changed, before located at /user/logout
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

	userParty.Get("/physical_ops/{opID:int}/events", GetEvents)
	userParty.Get("/physical_ops/{opID:int}/financial_commitments", GetOpFcs)                         // changed, before financialcommitments
	userParty.Get("/physical_ops/{opID:int}/financial_commitments/{fcID:int}/payments", GetFcPayment) // changed, before financialcommitments
	userParty.Get("/events", GetNextMonthEvent)
	userParty.Post("/physical_ops/{opID:int}/events", CreateEvent)
	userParty.Put("/physical_ops/{opID:int}/events/{evID:int}", ModifyEvent)
	userParty.Delete("/physical_ops/{opID:int}/events/{evID:int}", DeleteEvent)

	userParty.Get("/financial_commitments/month", GetMonthFC) // changed, before financialcommitments
	userParty.Get("/import_log", GetImportLogs)

	userParty.Get("/payment_ratios", GetRatios)
	userParty.Get("/payment_types/{ptID:int}/payment_ratios", GetPtRatios)
	userParty.Get("/payment_ratios/year", GetYearRatios)

	userParty.Get("/payment_types", GetPaymentTypes)
	userParty.Get("/payments/month", GetPaymentsPerMonth)
	userParty.Get("/payments/prevision_realized", GetPrevisionRealized)
	userParty.Get("/payments/cumulated", GetCumulatedMonthPayment)

	userParty.Get("/pending_commitments", GetPendings)

	userParty.Get("/plans/{pID:int}/planlines", GetPlanLines)
	userParty.Get("/plans/{pID:int}/planlines/detailed", GetDetailedPlanLines)
}

// setDBMiddleware return a middleware to add db to context values
func setDBMiddleware(db *gorm.DB) func(iris.Context) {
	return func(ctx iris.Context) {
		ctx.Values().Set("db", db)
		ctx.Next()
	}
}
