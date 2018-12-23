package models

import "database/sql"

// OpBeneficiaryValue model.
type OpBeneficiaryValue struct {
	Beneficiary string `json:"beneficiary"`
	Value       int64  `json:"value"`
}

// PaymentsPerBeneficiary embeddes a array of OpBeneficiaryValue for json export.
type PaymentsPerBeneficiary struct {
	Payments []OpBeneficiaryValue `json:"PaymentPerBeneficiary"`
}

// FinancialCommitmentsPerBeneficiary embeddes an array of OpBeneficiaryValue for json export.
type FinancialCommitmentsPerBeneficiary struct {
	Commitments []OpBeneficiaryValue `json:"FinancialCommitmentPerBeneficiary"`
}

// GetOpAll fetches payments per beneficiaries linked to a physical
// operation from database.
func (p *PaymentsPerBeneficiary) GetOpAll(opID int64, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT b.name, SUM(p.value - p.cancelled_value)
	FROM payment p, financial_commitment f, beneficiary b
	WHERE p.financial_commitment_id = f.id AND b.code = f.beneficiary_code AND
	p.financial_commitment_id IN (SELECT f.id FROM financial_commitment f WHERE f.physical_op_id = $1)
	GROUP BY b.name`, opID)
	if err != nil {
		return err
	}
	var r OpBeneficiaryValue
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Beneficiary, &r.Value); err != nil {
			return err
		}
		p.Payments = append(p.Payments, r)
	}
	err = rows.Err()
	return err
}

// GetOpAll fetches financial commitments per beneficiaries linked to a physical
// operation from database.
func (p *FinancialCommitmentsPerBeneficiary) GetOpAll(opID int64, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT b.name, SUM(f.value) FROM financial_commitment f  
	JOIN beneficiary b ON b.code=f.beneficiary_code 
	WHERE f.physical_op_id = $1 GROUP BY b.name`, opID)
	if err != nil {
		return err
	}
	var r OpBeneficiaryValue
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Beneficiary, &r.Value); err != nil {
			return err
		}
		p.Commitments = append(p.Commitments, r)
	}
	err = rows.Err()
	return err
}
