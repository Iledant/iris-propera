package actions

import (
	"database/sql"

	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris"
)

// SetRoutes initialize all routes for the application
func SetRoutes(app *iris.Application, db *sql.DB) {
	crs := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	crsParty := app.Party("/api", crs).AllowMethods(iris.MethodOptions)

	crsParty.Post("/user/signup", setDBMiddleware(db), SignUp)
	crsParty.Post("/user/signin", setDBMiddleware(db), Login)
	api := crsParty.Party("", setDBMiddleware(db))

	adminParty := api.Party("", AdminMiddleware)

	adminParty.Get("/user", GetUsers)
	adminParty.Post("/user", CreateUser)
	adminParty.Put("/user/{userID:int}", UpdateUser)
	adminParty.Delete("/user/{userID:int}", DeleteUser)
	adminParty.Get("/user/{userID:int}/rights", GetRight)
	adminParty.Post("/user/{userID:int}/rights", SetRight)
	adminParty.Post("/user/{userID:int}/inherit", InheritRight)

	adminParty.Post("/physical_ops", CreatePhysicalOp)
	adminParty.Post("/physical_ops/array", BatchPhysicalOps)
	adminParty.Delete("/physical_ops/{opID:int}", DeletePhysicalOp)
	adminParty.Get("/physical_ops/financial_commitments", GetOpsAndFCs)

	adminParty.Get("/beneficiaries", GetBeneficiaries)
	adminParty.Put("/beneficiaries/{beneficiaryID:int}", UpdateBeneficiary)

	adminParty.Get("/budget_chapters", GetBudgetChapters)
	adminParty.Post("/budget_chapters", CreateBudgetChapter)
	adminParty.Put("/budget_chapters/{bcID:int}", ModifyBudgetChapter)
	adminParty.Delete("/budget_chapters/{bcID:int}", DeleteBudgetChapter)

	adminParty.Post("/budget_sectors", CreateBudgetSector)
	adminParty.Put("/budget_sectors/{bsID:int}", ModifyBudgetSector)
	adminParty.Delete("/budget_sectors/{bsID:int}", DeleteBudgetSector)

	adminParty.Post("/budget_chapters/{chpID:int}/programs", CreateBudgetProgram)
	adminParty.Put("/budget_chapters/{chpID:int}/programs/{bpID:int}", ModifyBudgetProgram)
	adminParty.Delete("/budget_chapters/{chpID:int}/programs/{bpID:int}", DeleteBudgetProgram)
	adminParty.Post("/budget_programs", BatchBudgetProgram)

	adminParty.Post("/budget_credits", CreateBudgetCredit)
	adminParty.Put("/budget_credits/{brID:int}", ModifyBudgetCredit)
	adminParty.Post("/budget_credits/array", BatchBudgetCredits)
	adminParty.Delete("/budget_credits/{brID:int}", DeleteBudgetCredit)

	adminParty.Get("/budget_chapters/{chpID:int}/programs/{prgID:int}/actions",
		GetProgramBudgetActions)
	adminParty.Post("/budget_chapters/{chpID:int}/programs/{prgID:int}/actions",
		CreateBudgetAction)
	adminParty.Post("/budget_actions", BatchBudgetActions)
	adminParty.Put("/budget_chapters/{chpID:int}/programs/{prgID:int}/actions/{baID:int}",
		ModifyBudgetAction)
	adminParty.Delete("/budget_chapters/{chpID:int}/programs/{prgID:int}/actions/{baID:int}",
		DeleteBudgetAction)

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
	adminParty.Post("/pending_commitments/unlink", UnlinkPCs)
	adminParty.Post("/pending_commitments", BatchPendings)

	adminParty.Post("/payments", BatchPayments)

	adminParty.Post("/plans/{pID:int}/planlines", CreatePlanLine)
	adminParty.Put("/plans/{pID:int}/planlines/{plID:int}", ModifyPlanLine)
	adminParty.Delete("/plans/{pID:int}/planlines/{plID:int}", DeletePlanLine)
	adminParty.Post("/plans/{pID:int}/planlines/array", BatchPlanLines)

	adminParty.Post("/plans", CreatePlan)
	adminParty.Put("/plans/{pID:int}", ModifyPlan)
	adminParty.Delete("/plans/{pID:int}", DeletePlan)

	adminParty.Post("/prev_commitments", BatchPrevCommitments)

	adminParty.Post("/programmings/array", BatchProgrammings)

	adminParty.Post("/steps", CreateStep)
	adminParty.Put("/steps/{stID:int}", ModifyStep)
	adminParty.Delete("/steps/{stID:int}", DeleteStep)

	adminParty.Get("/settings", getSettings)
	adminParty.Get("/budget_tables", getBudgetTables)

	adminParty.Post("/today_message", SetTodayMessage)

	adminParty.Get("/scenarios", GetScenarios)
	adminParty.Post("/scenarios", CreateScenario)
	adminParty.Put("/scenarios/{sID:int}", ModifyScenario)
	adminParty.Delete("/scenarios/{sID:int}", DeleteScenario)
	adminParty.Get("/scenarios/{sID:int}", GetScenarioDatas)
	adminParty.Post("/scenarios/{sID:int}/offsets", SetScenarioOffsets)
	adminParty.Get("/scenarios/{sID:int}/payment_per_budget_action",
		GetScenarioActionPayments)
	adminParty.Get("/scenarios/{sID:int}/statistical_payment_per_budget_action",
		GetScenarioStatActionPayments)
	adminParty.Get("/scenarios/{sID:int}/budget", GetMultiAnnualScenario)

	userParty := api.Party("", ActiveMiddleware)
	userParty.Post("/user/logout", Logout)
	userParty.Get("/physical_ops", GetPhysicalOps)
	userParty.Put("/physical_ops/{opID:int}", UpdatePhysicalOp)
	userParty.Post("/user/password", ChangeUserPwd)

	userParty.Get("/budget_actions", GetAllBudgetActions)

	userParty.Get("/budget_credits/year", GetLastBudgetCredits)
	userParty.Get("/budget_credits", GetBudgetCredits)

	userParty.Get("/budget_programs", GetAllBudgetPrograms)
	userParty.Get("/budget_chapters/{chpID:int}/programs", GetChapterBudgetPrograms)

	userParty.Get("/budget_sectors", GetBudgetSectors)

	userParty.Get("/commissions", GetCommissions)

	userParty.Get("/physical_ops/{opID:int}/documents", GetDocuments)
	userParty.Post("/physical_ops/{opID:int}/documents", CreateDocument)
	userParty.Put("/physical_ops/{opID:int}/documents/{doID:int}", ModifyDocument)
	userParty.Delete("/physical_ops/{opID:int}/documents/{doID:int}", DeleteDocument)

	userParty.Get("/physical_ops/{opID:int}/events", GetEvents)
	userParty.Get("/physical_ops/{opID:int}/financial_commitments", GetOpFcs) // changed, before financialcommitments
	userParty.Get("/physical_ops/{opID:int}/financial_commitments/{fcID:int}/payments",
		GetFcPayment) // changed, before financialcommitments
	userParty.Get("/events", GetNextMonthEvent)
	userParty.Post("/physical_ops/{opID:int}/events", CreateEvent)
	userParty.Put("/physical_ops/{opID:int}/events/{evID:int}", ModifyEvent)
	userParty.Delete("/physical_ops/{opID:int}/events/{evID:int}", DeleteEvent)

	userParty.Get("/physical_ops/{opID:int}/previsions", GetOpPrevisions)
	userParty.Get("/physical_ops/{opID:int}/only_previsions", GetOpOnlyPrevisions)
	userParty.Post("/physical_ops/{opID:int}/previsions", SetOpPrevisions)

	userParty.Get("/financial_commitments/month", GetMonthFC)
	userParty.Get("/import_log", GetImportLogs)

	userParty.Get("/payment_ratios", GetRatios)
	userParty.Get("/payment_types/{ptID:int}/payment_ratios", GetPtRatios)
	userParty.Get("/payment_ratios/year", GetYearRatios)

	userParty.Get("/payment_types", GetPaymentTypes)
	userParty.Get("/payments/month", GetPaymentsPerMonth)
	userParty.Get("/payments/prevision_realized", GetPrevisionRealized)
	userParty.Get("/payments/month_cumulated", GetCumulatedMonthPayment)
	userParty.Get("/payments", GetAllPayments)

	userParty.Get("/pending_commitments", GetPendings)

	userParty.Get("/plans/{pID:int}/planlines", GetPlanLines)
	userParty.Get("/plans/{pID:int}/planlines/detailed", GetDetailedPlanLines)

	userParty.Get("/plans", GetPlans)

	userParty.Get("/pre_programmings", GetPreProgrammings)
	userParty.Post("/pre_programmings", BatchPreProgrammings)

	userParty.Get("/programmings", GetProgrammings)
	userParty.Get("/programmings/years", GetProgrammingsYear)

	userParty.Get("/summaries/multiannual_programmation", GetMultiannualProg)
	userParty.Get("/summaries/annual_programmation", GetAnnualProgrammation)
	userParty.Get("/summaries/programmation_prevision", GetProgrammingAndPrevisions)
	userParty.Get("/summaries/budget_action_programmation", GetActionProgrammation)
	userParty.Get("/summaries/budget_action_programmation_years",
		GetActionProgrammationAndYears)
	userParty.Get("/summaries/commitment_per_budget_action", GetActionCommitment)
	userParty.Get("/summaries/detailed_commitment_per_budget_action",
		GetDetailedActionCommitment)
	userParty.Get("/summaries/payment_per_budget_action", GetActionPayment)
	userParty.Get("/summaries/statistical_payment_per_budget_action", GetStatActionPayment)
	userParty.Get("/summaries/detailed_payment_per_budget_action", GetDetailedActionPayment)
	userParty.Get("/summaries/statistical_detailed_payment_per_budget_action",
		GetStatDetailedActionPayment)
	userParty.Get("/summaries/statistical_current_year_payment_per_budget_action",
		GetStatCurrentYearPayment)

	userParty.Get("/today_message", GetTodayMessage)

	userParty.Get("/op_dpt_ratios/ops", GetOpWithDptRatios)
	userParty.Post("/op_dpt_ratios/upload", BatchOpDptRatios)
	userParty.Get("/op_dpt_ratios/financial_commitments", GetFCPerDpt)
	userParty.Get("/op_dpt_ratios/detailed_financial_commitments", GetDetailedFCPerDpt)
	userParty.Get("/op_dpt_ratios/detailed_programmings", GetDetailedPrgPerDpt)

	userParty.Get("/home", GetHomeDatas)

	userParty.Get("/steps", GetSteps)
}

// setDBMiddleware return a middleware to add db to context values
func setDBMiddleware(db *sql.DB) func(iris.Context) {
	return func(ctx iris.Context) {
		ctx.Values().Set("db", db)
		ctx.Next()
	}
}
