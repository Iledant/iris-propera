package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// PaymentDemandLine is used to decode one line of payments demands batch
type PaymentDemandLine struct {
	IrisCode        string        `json:"iris_code"`
	IrisName        string        `json:"iris_name"`
	CommitmentDate  ExcelDate     `json:"commitment_date"`
	BeneficiaryCode int64         `json:"beneficiary_code"`
	DemandNumber    int64         `json:"demand_number"`
	DemandDate      ExcelDate     `json:"demand_date"`
	ReceiptDate     ExcelDate     `json:"receipt_date"`
	DemandValue     int64         `json:"demand_value"`
	CsfDate         NullExcelDate `json:"csf_date"`
	CsfComment      NullString    `json:"csf_comment"`
	DemandStatus    string        `json:"demand_status"`
	StatusComment   NullString    `json:"status_comment"`
}

// PaymentDemandBatch embeddes an array of PaymentDemandLine for dedicated query
type PaymentDemandBatch struct {
	Lines []PaymentDemandLine `json:"PaymentDemand"`
}

// PaymentDemand model
type PaymentDemand struct {
	ID              int64      `json:"id"`
	IrisCode        string     `json:"iris_code"`
	IrisName        string     `json:"iris_name"`
	BeneficiaryID   int64      `json:"beneficiary_code"`
	Beneficiary     string     `json:"beneficiary"`
	DemandNumber    int64      `json:"demand_number"`
	DemandDate      time.Time  `json:"demand_date"`
	ReceiptDate     time.Time  `json:"receipt_date"`
	DemandValue     int64      `json:"demand_value"`
	CsfDate         NullTime   `json:"csf_date"`
	CsfComment      NullString `json:"csf_comment"`
	DemandStatus    string     `json:"demand_status"`
	StatusComment   NullString `json:"status_comment"`
	Excluded        NullBool   `json:"excluded"`
	ExcludedComment NullString `json:"excluded_comment"`
	ProcessedDate   NullTime   `json:"processed_date"`
}

// PaymentDemands embeddes an array of PaymentDemand for json export and dedicated
// queries
type PaymentDemands struct {
	Lines []PaymentDemand `json:"PaymentDemand"`
}

// Update set excluded fields in the database
func (p *PaymentDemand) Update(db *sql.DB) error {
	res, err := db.Exec(`UPDATE payment_demands SET excluded=$1, excluded_comment=$2
	WHERE id=$3`, p.Excluded, p.ExcludedComment, p.ID)
	if err != nil {
		return fmt.Errorf("update %v", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected %v", err)
	}
	if count == 0 {
		return fmt.Errorf("demande de paiement introuvable")
	}
	return nil
}

// GetAll fetches all payment demand from database
func (p *PaymentDemands) GetAll(db *sql.DB) error {
	rows, err := db.Query(`SELECT p.id,p.iris_code,p.iris_name,
	p.beneficiary_id,b.name,p.demand_number,p.demand_date,p.receipt_date,
	p.demand_value,p.csf_date,p.csf_comment,p.demand_status,p.status_comment,
	p.excluded,p.excluded_comment,p.processed_date
	FROM payment_demands p
	JOIN beneficiary b on  b.id=p.beneficiary_id`)
	if err != nil {
		return fmt.Errorf("select %v", err)
	}
	var l PaymentDemand
	for rows.Next() {
		if err = rows.Scan(&l.ID, &l.IrisCode, &l.IrisName, &l.BeneficiaryID,
			&l.Beneficiary, &l.DemandNumber, &l.DemandDate, &l.ReceiptDate,
			&l.DemandValue, &l.CsfDate, &l.CsfComment, &l.DemandStatus,
			&l.StatusComment, &l.Excluded, &l.ExcludedComment, &l.ProcessedDate); err != nil {
			return fmt.Errorf("scan %v", err)
		}
		p.Lines = append(p.Lines, l)
	}
	err = rows.Err()
	if err != nil {
		return fmt.Errorf("rows err %v", err)
	}
	if len(p.Lines) == 0 {
		p.Lines = []PaymentDemand{}
	}
	return nil
}

// Validate checks if a payment batch has correct fields
func (p *PaymentDemandBatch) Validate() error {
	for i, l := range p.Lines {
		if l.IrisCode == "" {
			return fmt.Errorf("ligne %d iris_code vide", i+1)
		}
		if l.IrisName == "" {
			return fmt.Errorf("ligne %d iris_name vide", i+1)
		}
		if int64(l.CommitmentDate) == 0 {
			return fmt.Errorf("ligne %d commitment_date vide", i+1)
		}
		if l.BeneficiaryCode == 0 {
			return fmt.Errorf("ligne %d beneficiary_code vide", i+1)
		}
		if l.DemandNumber == 0 {
			return fmt.Errorf("ligne %d demand_number vide", i+1)
		}
		if int64(l.DemandDate) == 0 {
			return fmt.Errorf("ligne %d demand_date vide", i+1)
		}
		if int64(l.ReceiptDate) == 0 {
			return fmt.Errorf("ligne %d receipt_date vide", i+1)
		}
		if l.DemandValue == 0 {
			return fmt.Errorf("ligne %d demand_value vide", i+1)
		}
	}
	return nil
}

// Save import a batch of PaymentDemandLine and update the database accordingly
func (p *PaymentDemandBatch) Save(db *sql.DB) error {
	if err := p.Validate(); err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("transaction begin %v", err)
	}

	if _, err := tx.Exec("DELETE from temp_payment_demands"); err != nil {
		tx.Rollback()
		return err
	}

	stmt, err := tx.Prepare(pq.CopyIn("temp_payment_demands", "iris_code",
		"iris_name", "commitment_date", "beneficiary_code", "demand_number",
		"demand_date", "receipt_date", "demand_value", "csf_date", "csf_comment",
		"demand_status", "status_comment"))
	if err != nil {
		return fmt.Errorf("prepare stmt %v", err)
	}
	defer stmt.Close()
	for _, r := range p.Lines {
		if _, err = stmt.Exec(r.IrisCode, r.IrisName, r.CommitmentDate.ToDate(),
			r.BeneficiaryCode, r.DemandNumber, r.DemandDate.ToDate(),
			r.ReceiptDate.ToDate(), r.DemandValue, r.CsfDate.ToDate(), r.CsfComment,
			r.DemandStatus, r.StatusComment); err != nil {
			tx.Rollback()
			return fmt.Errorf("insertion de %+v  %v", r, err)
		}
	}
	if _, err = stmt.Exec(); err != nil {
		tx.Rollback()
		return fmt.Errorf("statement exec flush %v", err)
	}
	queries := []string{
		`INSERT INTO payment_demands (iris_code,iris_name,beneficiary_id,
			demand_number,demand_date,receipt_date,demand_value,csf_date,
			csf_comment,demand_status,status_comment,excluded,excluded_comment,
			processed_date)
		SELECT t.iris_code,t.iris_name,b.id,t.demand_number,t.demand_date,
			t.receipt_date,t.demand_value,t.csf_date,t.csf_comment,t.demand_status,
			t.status_comment,NULL::boolean,NULL::text,NULL::date
		FROM imported_payment_demands t
		JOIN beneficiary b ON b.code=t.beneficiary_code
		WHERE (t.iris_code,t.beneficiary_code,t.demand_number) NOT IN 
		(SELECT iris_code,beneficiary_code,demand_number FROM payment_demands)`,
		`UPDATE payment_demands SET csf_date=t.csf_date,csf_comment=t.csf_comment,
			demand_status=t.demand_status,status_comment=t.status_comment
		FROM (SELECT t.*,b.id AS beneficiary_id FROM imported_payment_demands t
				JOIN beneficiary b ON t.beneficiary_code=b.code) t
			WHERE (payment_demands.iris_code=t.iris_code AND
			payment_demands.beneficiary_id=t.beneficiary_id AND
			payment_demands.demand_number=t.demand_number)`,
		`UPDATE payment_demands SET processed_date=CURRENT_DATE
		WHERE (iris_code,beneficiary_id,demand_number) NOT IN 	
			(SELECT t.iris_code,b.id,t.demand_number FROM imported_payment_demands t
				JOIN beneficiary b ON t.beneficiary_code=b.code)
			AND processed_date IS NULL`,
	}
	for i, q := range queries {
		if _, err := tx.Exec(q); err != nil {
			tx.Rollback()
			return fmt.Errorf("requête %d %v", i+1, err)
		}
	}
	return tx.Commit()
}
