package actions

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Iledant/iris_propera/actions/queries"
	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// paymentResp embeddes an array of payment sent back
type paymentResp struct {
	Payments []models.Payment `json:"Payment"`
}

// GetFcPayment handles the get request fetching all payments of a financial commitment.
func GetFcPayment(ctx iris.Context) {
	fcID, err := ctx.Params().GetInt("fcID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiements d'un engagement : " + err.Error()})
		return
	}

	fc, db := models.FinancialCommitment{}, ctx.Values().Get("db").(*gorm.DB)
	if err := db.Find(&fc, fcID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Paiements d'un engagement : introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiements d'un engagement : " + err.Error()})
		return
	}

	resp := paymentResp{}
	if err := db.Where("financial_commitment_id = ?", fcID).Find(&resp.Payments).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiements d'un engagement : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// paymentPerMonth is used to fetch results for the query calculating it.
type paymentPerMonth struct {
	Year  int64 `json:"year"`
	Month int64 `json:"month"`
	Value int64 `json:"value"`
}

// paymentPerMonthResp embeddes the response for payments per month request.
type paymentPerMonthResp struct {
	Payments []paymentPerMonth `json:"PaymentsPerMonth"`
}

// GetPaymentsPerMonth handles the get request fetching payments per month of a given year and the precedent.
func GetPaymentsPerMonth(ctx iris.Context) {
	y, err := ctx.URLParamInt("year")
	if err != nil {
		y = time.Now().Year()
	}

	year := time.Date(y-1, 1, 1, 0, 0, 0, 0, time.UTC)
	db := ctx.Values().Get("db").(*gorm.DB)

	rows, err := db.Raw(queries.GetPaymentsPerMonth, year).Rows()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Paiements par mois : " + err.Error()})
		return
	}
	defer rows.Close()

	resp, ppm := paymentPerMonthResp{}, paymentPerMonth{}
	for rows.Next() {
		db.ScanRows(rows, &ppm)
		resp.Payments = append(resp.Payments, ppm)
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// paymentReq is used to decode a line of payment batch payload.
type paymentReq struct {
	CoriolisYear    string    `json:"coriolis_year"`
	CoriolisEgtCode string    `json:"coriolis_egt_code"`
	CoriolisEgtNum  string    `json:"coriolis_egt_num"`
	CoriolisEgtLine string    `json:"coriolis_egt_line"`
	Date            time.Time `json:"date"`
	Number          string    `json:"number"`
	Value           float64   `json:"value"`
	CancelledValue  float64   `json:"cancelled_value"`
	BeneficiaryCode int       `json:"beneficiary_code"`
}

// batchPaymentReq is used to decode batch payment payload.
type batchPaymentReq struct {
	Pp []paymentReq `json:"Payment"`
}

// BatchPayments handles the request sending an array of payments.
func BatchPayments(ctx iris.Context) {
	req := batchPaymentReq{}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch de paiements, lecture du payload : " + err.Error()})
		return
	}

	db := ctx.Values().Get("db").(*gorm.DB)
	tx := db.Begin()

	if err := tx.Exec(queries.DeleteTempPayment).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch de paiements, vidage table temporaire : " + err.Error()})
		tx.Rollback()
		return
	}

	for _, p := range req.Pp {
		if err := tx.Exec(queries.InsertTempPayment, p.CoriolisYear, p.CoriolisEgtCode, p.CoriolisEgtNum,
			p.CoriolisEgtLine, p.BeneficiaryCode, p.Date, p.Value*100, p.CancelledValue*100, p.Number).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Batch de paiements insertion : " + err.Error()})
			tx.Rollback()
			return
		}
	}

	qry := []string{queries.UpdatePaymentWithTemp, queries.InsertTempIntoPayment, queries.CalculatePaymentFcID}
	for _, q := range qry {
		if err := tx.Exec(q).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Batch de paiements requêtes : " + err.Error()})
			tx.Rollback()
			return
		}
	}

	if err := tx.Exec(queries.DeleteTempPayment).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch de paiements, vidage table temporaire : " + err.Error()})
		tx.Rollback()
		return
	}

	if err := tx.Exec("UPDATE import_logs SET last_date = ? WHERE category = 'Payments'", time.Now()).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	tx.Commit()
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Paiements importés"})
}

// prevRealized is used to fetch a row of the query result.
type prevRealized struct {
	Name        string `json:"name"`
	PrevPayment int64  `json:"prev_payment"`
	Payment     int64  `json:"payment"`
}

// prevRealizedResp embeddes the array of prevRealized for the response.
type prevRealizedResp struct {
	Prr []prevRealized `json:"PaymentPrevisionAndRealized"`
}

// GetPrevisionRealized handles the request to the payment prevision and real payments for the given year and beneficiary.
func GetPrevisionRealized(ctx iris.Context) {
	year, err := ctx.URLParamInt64("year")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Prévu réalisé erreur sur year : " + err.Error()})
		return
	}

	ptID, err := ctx.URLParamInt64("paymentTypeId")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévu réalisé erreur sur paymentTypeId : " + err.Error()})
		return
	}

	db := ctx.Values().Get("db").(*gorm.DB)
	pt := models.PaymentType{ID: int(ptID)}
	if err := db.First(&pt).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Prévu réalisé : chronique introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévu réalisé erreur de requête : " + err.Error()})
		return
	}

	rows, err := db.Raw(queries.PrevisionRealized, ptID, year, year, year).Rows()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévu réalisé erreur de requête : " + err.Error()})
		return
	}
	defer rows.Close()
	resp, pr := prevRealizedResp{}, prevRealized{}
	for rows.Next() {
		db.ScanRows(rows, &pr)
		resp.Prr = append(resp.Prr, pr)
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// cumulatedPayment is used to fetch the query result.
type cumulatedPayment struct {
	Year      int64   `json:"year"`
	Month     int64   `json:"month"`
	Cumulated float64 `json:"cumulated"`
}

// cumulatedPaymentResp embeddes an array of cumulatedPayments.
type cumulatedPaymentResp struct {
	Cpp []cumulatedPayment `json:"MonthCumulatedPayment"`
}

// GetCumulatedMonthPayment handles the request to calculate cumulated payment per month for all or for one beneficiary.
func GetCumulatedMonthPayment(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	var rows *sql.Rows

	bID, err := ctx.URLParamInt("beneficiaryId")
	if err != nil {
		rows, err = db.Raw(queries.MonthCumulatedAll).Rows()
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Paiement cumulés, erreur de requête : " + err.Error()})
			return
		}
	} else {
		b := models.Beneficiary{ID: bID}
		if err = db.First(&b).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				ctx.StatusCode(http.StatusBadRequest)
				ctx.JSON(jsonMessage{"Paiements cumulés : bénéficiaire introuvables"})
				return
			}
		}

		rows, err = db.Raw(queries.MonthCumulatedBeneficiary, b.Code).Rows()
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Paiement cumulés, erreur de requête : " + err.Error()})
			return
		}
	}

	defer rows.Close()
	resp, cp := cumulatedPaymentResp{}, cumulatedPayment{}
	for rows.Next() {
		db.ScanRows(rows, &cp)
		resp.Cpp = append(resp.Cpp, cp)
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
