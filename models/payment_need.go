package models

import (
	"database/sql"
	"fmt"
	"time"
)

// PaymentNeed model
type PaymentNeed struct {
	ID              int64     `json:"ID"`
	BeneficiaryID   int64     `json:"BeneficiaryID"`
	BeneficiaryName string    `json:"BeneficiaryName"`
	Date            time.Time `json:"Date"`
	Value           int64     `json:"Value"`
	Comment         string    `json:"Comment"`
}

// PaymentNeeds embeddes an array of PAymentNeed for json export and dedicated
// queries
type PaymentNeeds struct {
	Lines []PaymentNeed `json:"PaymentNeed"`
}

// LastPaymentNeed is used to decode one line of the query that fetches the most
// recent payment need for each beneficiary
type LastPaymentNeed struct {
	BeneficiaryName string     `json:"BeneficiaryName"`
	Date            NullTime   `json:"Date"`
	Need            NullInt64  `json:"Need"`
	Comment         NullString `json:"Comment"`
	Payment         NullInt64  `json:"Payment"`
	Forecast        NullInt64  `json:"Forecast"`
}

// LastPaymentNeeds embeddes an array of LastPaymentNeed forjson export
type LastPaymentNeeds struct {
	Lines            []LastPaymentNeed `json:"PaymentNeed"`
	RemainingPayment int64             `json:"RemainingPayment"`
}

// validate checks if fields matches database constraints
func (p *PaymentNeed) validate() error {
	if p.BeneficiaryID == 0 {
		return fmt.Errorf("beneficiary ID nul")
	}
	if p.Value == 0 {
		return fmt.Errorf("value nul")
	}
	return nil
}

// Create insert a new PaymentNeed into database
func (p *PaymentNeed) Create(db *sql.DB) error {
	if err := p.validate(); err != nil {
		return err
	}
	if err := db.QueryRow(`INSERT INTO payment_need (beneficiary_id,date,value,comment) 
	VALUES ($1,$2,$3,$4) RETURNING id`, p.BeneficiaryID, p.Date, p.Value,
		p.Comment).Scan(&p.ID); err != nil {
		return fmt.Errorf("insert %v", err)
	}
	if err := db.QueryRow(`SELECT name from beneficiary WHERE id=$1`,
		p.BeneficiaryID).Scan(&p.BeneficiaryName); err != nil {
		return fmt.Errorf("select %v", err)
	}
	return nil
}

// Update modifies a PaymentNeed into database
func (p *PaymentNeed) Update(db *sql.DB) error {
	if err := p.validate(); err != nil {
		return err
	}

	res, err := db.Exec(`UPDATE payment_need SET beneficiary_id=$1,date=$2,
		value=$3,comment=$4 WHERE id=$5`, p.BeneficiaryID, p.Date, p.Value,
		p.Comment, p.ID)
	if err != nil {
		return fmt.Errorf("update %v", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected %v", err)
	}
	if count != 1 {
		return fmt.Errorf("besoin de paiement introuvable")
	}
	if err := db.QueryRow(`SELECT name from beneficiary WHERE id=$1`,
		p.BeneficiaryID).Scan(&p.BeneficiaryName); err != nil {
		return fmt.Errorf("select %v", err)
	}
	return nil
}

// Delete remove a PaymentNeed from database
func (p *PaymentNeed) Delete(db *sql.DB) error {
	res, err := db.Exec(`DELETE FROM payment_need WHERE id=$1`, p.ID)
	if err != nil {
		return fmt.Errorf("delete %v", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected %v", err)
	}
	if count != 1 {
		return fmt.Errorf("payment need introuvable")
	}
	return err
}

// GetAll fetches all PaymentNeed of the given year from database
func (p *PaymentNeeds) GetAll(year int64, db *sql.DB) error {
	rows, err := db.Query(`SELECT p.id,p.beneficiary_id,b.name,p.date,p.value,
	p.comment FROM payment_need p JOIN beneficiary b ON p.beneficiary_id=b.id
	WHERE extract(year FROM p.date)=$1`, year)
	if err != nil {
		return fmt.Errorf("select %v", err)
	}
	defer rows.Close()
	var l PaymentNeed
	for rows.Next() {
		if err = rows.Scan(&l.ID, &l.BeneficiaryID, &l.BeneficiaryName, &l.Date,
			&l.Value, &l.Comment); err != nil {
			return fmt.Errorf("scan %v", err)
		}
		p.Lines = append(p.Lines, l)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("rows err %v", err)
	}
	if len(p.Lines) == 0 {
		p.Lines = []PaymentNeed{}
	}
	return nil
}

// GetAll fetches the last PaymentNeed for every beneficiary of a given year,
// the actual payment sum and the statistical forecast
func (p *LastPaymentNeeds) GetAll(year int64, pmtType int64, db *sql.DB) error {
	q := `
	WITH pr AS (SELECT * FROM payment_ratios WHERE payment_types_id=$1),
	fc_sum AS (SELECT beneficiary_code, EXTRACT(year FROM date)::integer AS year, 
							SUM(value) AS value 
							FROM financial_commitment WHERE EXTRACT(year FROM date)<=$2
							GROUP BY 1,2),
	fc AS (SELECT fc_sum.beneficiary_code, SUM(fc_sum.value * pr.ratio) AS value 
						FROM fc_sum, pr WHERE fc_sum.year + pr.index=$2 GROUP BY 1),
	p AS (SELECT beneficiary_code, SUM(value) AS value 
						FROM payment WHERE EXTRACT(YEAR from date)=$2 GROUP BY 1)
	SELECT 'Autre'::varchar(255),null::bigint,null::text,null::date,
		sum(fc.value)::bigint,sum(p.value)::bigint
	FROM fc
	JOIN beneficiary b ON fc.beneficiary_code=b.code
	JOIN p ON p.beneficiary_code=b.code
	WHERE b.id NOT IN
		(SELECT beneficiary_id FROM payment_need WHERE extract(year FROM date)=$2)
	UNION ALL
	SELECT b.name, pn.value,pn.comment, max(pn.date)::date,fc.value::bigint,
		COALESCE(p.value,0)::bigint
	FROM payment_need pn
	LEFT JOIN beneficiary b ON pn.beneficiary_id=b.id
	LEFT OUTER JOIN fc ON fc.beneficiary_code=b.code
	LEFT OUTER JOIN p ON p.beneficiary_code = b.code
	WHERE extract(year FROM pn.date)=$2
	GROUP BY 1,2,3,5,6;`

	rows, err := db.Query(q, pmtType, year)
	if err != nil {
		return fmt.Errorf("select %v", err)
	}
	defer rows.Close()
	var l LastPaymentNeed
	for rows.Next() {
		if err = rows.Scan(&l.BeneficiaryName, &l.Need, &l.Comment, &l.Date,
			&l.Forecast, &l.Payment); err != nil {
			return fmt.Errorf("scan %v", err)
		}
		p.Lines = append(p.Lines, l)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("rows err %v", err)
	}
	if len(p.Lines) == 0 {
		p.Lines = []LastPaymentNeed{}
	}
	return nil
}
