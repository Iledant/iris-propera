package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// Payment model
type Payment struct {
	ID                    int64     `json:"id"`
	FinancialCommitmentID NullInt64 `json:"financial_commitment_id"`
	CoriolisYear          string    `json:"coriolis_year"`
	CoriolisEgtCode       string    `json:"coriolis_egt_code"`
	CoriolisEgtNum        string    `json:"coriolis_egt_num"`
	CoriolisEgtLine       string    `json:"coriolis_egt_line"`
	Date                  time.Time `json:"date"`
	Number                string    `json:"number"`
	Value                 int64     `json:"value"`
	CancelledValue        int64     `json:"cancelled_value"`
	BeneficiaryCode       int64     `json:"beneficiary_code"`
	ReceiptDate           NullTime  `json:"receipt_date"`
}

// Payments embeddes an array of Payment for json export.
type Payments struct {
	Payments []Payment `json:"Payment"`
}

// PaymentPerMonth is used to fetch results for the query calculating it.
type PaymentPerMonth struct {
	Year  int64 `json:"year"`
	Month int64 `json:"month"`
	Value int64 `json:"value"`
}

// PaymentPerMonths embeddes an array of PaymentPerMonth for json export.
type PaymentPerMonths struct {
	PaymentPerMonths []PaymentPerMonth `json:"PaymentsPerMonth"`
}

// PaymentLine is used to decode a line of payment batch payload.
type PaymentLine struct {
	CoriolisYear    string        `json:"coriolis_year"`
	CoriolisEgtCode string        `json:"coriolis_egt_code"`
	CoriolisEgtNum  string        `json:"coriolis_egt_num"`
	CoriolisEgtLine string        `json:"coriolis_egt_line"`
	Date            ExcelDate     `json:"date"`
	Number          string        `json:"number"`
	Value           float64       `json:"value"`
	CancelledValue  float64       `json:"cancelled_value"`
	BeneficiaryCode int64         `json:"beneficiary_code"`
	ReceiptDate     NullExcelDate `json:"receipt_date"`
}

// PaymentBatch embeddes an array of PaymentLine for batch request.
type PaymentBatch struct {
	PaymentBatch []PaymentLine `json:"Payment"`
}

// PrevisionRealized is used to decode a line of the dedicated query.
type PrevisionRealized struct {
	Name        string `json:"name"`
	PrevPayment int64  `json:"prev_payment"`
	Payment     int64  `json:"payment"`
}

// PrevisionsRealized embeddes an array of PrevisionRealized for json export.
type PrevisionsRealized struct {
	PrevisionsRealized []PrevisionRealized `json:"PaymentPrevisionAndRealized"`
}

// MonthCumulatedPayment is used to decode a line of the dedicated query.
type MonthCumulatedPayment struct {
	Year      int64   `json:"year"`
	Month     int64   `json:"month"`
	Cumulated float64 `json:"cumulated"`
}

// MonthCumulatedPayments embeddes an array of MonthCumulatedPayment.
type MonthCumulatedPayments struct {
	MonthCumulatedPayments []MonthCumulatedPayment `json:"MonthCumulatedPayment"`
}

// GetFcAll fetches all payments linked to a financial commitment.
func (p *Payments) GetFcAll(fcID int64, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT  id, financial_commitment_id, coriolis_year, 
	coriolis_egt_code, coriolis_egt_num, coriolis_egt_line, date, number, value, 
	cancelled_value, beneficiary_code, receipt_date FROM payment 
	WHERE financial_commitment_id = $1`, fcID)
	if err != nil {
		return err
	}
	var r Payment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.FinancialCommitmentID, &r.CoriolisYear,
			&r.CoriolisEgtCode, &r.CoriolisEgtNum, &r.CoriolisEgtLine, &r.Date,
			&r.Number, &r.Value, &r.CancelledValue, &r.BeneficiaryCode, &r.ReceiptDate); err != nil {
			return err
		}
		p.Payments = append(p.Payments, r)
	}
	err = rows.Err()
	if len(p.Payments) == 0 {
		p.Payments = []Payment{}
	}
	return err
}

// GetAll calculates payments per month of a given year, fetching datas from database.
func (p *PaymentPerMonths) GetAll(year int, db *sql.DB) (err error) {
	d0 := time.Date(year-1, 1, 1, 0, 0, 0, 0, time.UTC)
	rows, err := db.Query(`SELECT EXTRACT(YEAR FROM date) AS year, 
  EXTRACT(MONTH FROM date) AS month, SUM(value - cancelled_value) AS value
FROM payment WHERE date >= $1 GROUP BY 1,2 ORDER BY 1,2`, d0)
	if err != nil {
		return err
	}
	var r PaymentPerMonth
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Year, &r.Month, &r.Value); err != nil {
			return err
		}
		p.PaymentPerMonths = append(p.PaymentPerMonths, r)
	}
	err = rows.Err()
	if len(p.PaymentPerMonths) == 0 {
		p.PaymentPerMonths = []PaymentPerMonth{}
	}
	return err
}

// Save a batch of payments to the database.
func (p *PaymentBatch) Save(db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err = tx.Exec("DELETE from temp_payment"); err != nil {
		tx.Rollback()
		return err
	}

	stmt, err := tx.Prepare(pq.CopyIn("temp_payment", "coriolis_year",
		"coriolis_egt_code", "coriolis_egt_num", "coriolis_egt_line",
		"beneficiary_code", "date", "value", "cancelled_value", "number",
		"receipt_date"))
	if err != nil {
		return fmt.Errorf("prepare stmt %v", err)
	}
	defer stmt.Close()
	for _, r := range p.PaymentBatch {
		if _, err = stmt.Exec(r.CoriolisYear, r.CoriolisEgtCode, r.CoriolisEgtNum,
			r.CoriolisEgtLine, r.BeneficiaryCode, r.Date.ToDate(), int64(100*r.Value),
			int64(100*r.CancelledValue), r.Number, r.ReceiptDate.ToDate()); err != nil {
			tx.Rollback()
			return fmt.Errorf("insertion de %+v  %v", r, err)
		}
	}
	if _, err = stmt.Exec(); err != nil {
		tx.Rollback()
		return fmt.Errorf("statement exec flush %v", err)
	}

	queries := []string{`WITH new AS (
		SELECT p.id, t.number, t.date, t.value, t.cancelled_value,t.receipt_date
			FROM temp_payment t
			LEFT JOIN payment p ON t.number = p.number AND t.date = p.date
		WHERE p.value <> t.value OR p.cancelled_value <> t.cancelled_value
			OR p.receipt_date <> t.receipt_date)
	UPDATE payment SET value = new.value, cancelled_value = new.cancelled_value, 
		receipt_date=new.receipt_date
	FROM new WHERE payment.id = new.id`,
		`INSERT INTO PAYMENT (financial_commitment_id, coriolis_year, coriolis_egt_code,
		coriolis_egt_num, coriolis_egt_line, date, number, value, cancelled_value, 
		beneficiary_code, receipt_date)
		SELECT NULL, coriolis_year, coriolis_egt_code, coriolis_egt_num, 
	 coriolis_egt_line, date, number, value, cancelled_value, beneficiary_code,
	 receipt_date FROM temp_payment t
		WHERE (t.number, t.date) NOT IN (SELECT number, date FROM payment)`,
		`WITH ref AS (
			SELECT DISTINCT ON (coriolis_year, coriolis_egt_code, coriolis_egt_num, coriolis_egt_line) 
			id, coriolis_year, coriolis_egt_code, coriolis_egt_num, coriolis_egt_line 
			FROM financial_commitment ORDER BY 2,3,4,5) 
			 UPDATE payment SET 
				 financial_commitment_id = ref.id 
			 FROM ref WHERE (payment.coriolis_year = ref.coriolis_year AND 
			payment.coriolis_egt_code = ref.coriolis_egt_code AND 
			payment.coriolis_egt_num = ref.coriolis_egt_num AND 
			payment.coriolis_egt_line = ref.coriolis_egt_line)`,
		"DELETE from temp_attachment"}
	for _, q := range queries {
		if _, err = tx.Exec(q); err != nil {
			tx.Rollback()
			return err
		}
	}
	if _, err = tx.Exec(`INSERT INTO import_logs (category,last_date)
		VALUES ('Payments', $1)
		ON CONFLICT (category) DO UPDATE SET last_date = EXCLUDED.last_date;`,
		time.Now()); err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}

// GetAll calculates payement previsions and realized from database.
func (p *PrevisionsRealized) GetAll(year int64, ptID int64, db *sql.DB) (err error) {
	rows, err := db.Query(`WITH pr AS (SELECT * FROM payment_ratios WHERE payment_types_id=$1),
	fc_sum AS (SELECT beneficiary_code, EXTRACT(year FROM date)::integer AS year, SUM(value) AS value 
								FROM financial_commitment WHERE EXTRACT(year FROM date)<$2 GROUP BY 1,2),
	fc AS (SELECT fc_sum.beneficiary_code, SUM(fc_sum.value * pr.ratio) AS value 
						FROM fc_sum, pr WHERE fc_sum.year + pr.index=$2 GROUP BY 1),
	p AS (SELECT beneficiary_code, SUM(value) AS value 
						FROM payment WHERE EXTRACT(YEAR from date)=$2 GROUP BY 1)
	SELECT b.name, fc.value::bigint AS prev_payment, COALESCE(p.value,0) AS payment FROM fc
	LEFT JOIN beneficiary b ON fc.beneficiary_code = b.code
	LEFT OUTER JOIN p ON p.beneficiary_code = b.code
	ORDER BY 2 DESC`, ptID, year)
	if err != nil {
		return err
	}
	var r PrevisionRealized
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Name, &r.PrevPayment, &r.Payment); err != nil {
			return err
		}
		p.PrevisionsRealized = append(p.PrevisionsRealized, r)
	}
	err = rows.Err()
	if len(p.PrevisionsRealized) == 0 {
		p.PrevisionsRealized = []PrevisionRealized{}
	}
	return err
}

// GetAll calculates month cumulated payments for a beneficiary or for all ones if ID is 0.
func (m *MonthCumulatedPayments) GetAll(bID int64, db *sql.DB) (err error) {
	var rows *sql.Rows
	if bID != 0 {
		rows, err = db.Query(`SELECT tot.year, tot.month, sum(tot.value) 
		OVER (PARTITION BY tot.year ORDER BY tot.month) as cumulated FROM
		(SELECT extract(month from p.date) as month, EXTRACT(year FROM p.date) AS year, 
				0.01*sum(p.value) as value 
			FROM payment p, beneficiary b 
			WHERE p.beneficiary_code=b.code AND b.id=$1 GROUP BY 1,2 ORDER BY 2,1) tot
		ORDER BY 1,2`, bID)
	} else {
		rows, err = db.Query(`SELECT tot.year, tot.month, sum(tot.value) OVER
		 (PARTITION BY tot.year ORDER BY tot.month) as cumulated FROM
		(SELECT extract(month from DATE) as month, EXTRACT(year FROM date) AS year,
		 0.01*sum(value) as value 
			 FROM payment GROUP BY 1,2 ORDER BY 2,1) tot ORDER BY 1,2`)
	}
	if err != nil {
		return err
	}
	var r MonthCumulatedPayment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Year, &r.Month, &r.Cumulated); err != nil {
			return err
		}
		m.MonthCumulatedPayments = append(m.MonthCumulatedPayments, r)
	}
	err = rows.Err()
	if len(m.MonthCumulatedPayments) == 0 {
		m.MonthCumulatedPayments = []MonthCumulatedPayment{}
	}
	return err
}

// LinkCmt is used to add to a payment a link to a commitment
func (p *Payment) LinkCmt(cmtID int64, db *sql.DB) error {
	res, err := db.Exec(`UPDATE payment SET financial_commitment_id=$1 WHERE id=$2`,
		cmtID, p.ID)
	if err != nil {
		return fmt.Errorf("update %v", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("count %v", err)
	}
	if count != 1 {
		return fmt.Errorf("payment not found")
	}
	return nil
}
