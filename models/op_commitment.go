package models

import "database/sql"

// OpAndCommitment is used to decade the query that fetches link between physical operation
// and financial commitment.
type OpAndCommitment struct {
	Number   NullString `json:"number"`
	Name     NullString `json:"op_name"`
	IrisCode NullString `json:"iris_code"`
	IrisName NullString `json:"iris_name"`
}

// OpAndCommitments embeddes an array of OpAndCommitment for json export.
type OpAndCommitments struct {
	OpAndCommitments []OpAndCommitment `json:"PhysicalOpFinancialCommitments"`
}

// GetAll fetches the list of physical operations and financial commitments linked.
func (o *OpAndCommitments) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT op.number, op.name AS op_name, f.iris_code, 
	f.name as iris_name FROM financial_commitment f
	FULL OUTER JOIN physical_op op ON f.physical_op_id = op.id
	ORDER BY 1,3`)
	if err != nil {
		return err
	}
	var r OpAndCommitment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Number, &r.Name, &r.IrisCode, &r.IrisName); err != nil {
			return err
		}
		o.OpAndCommitments = append(o.OpAndCommitments, r)
	}
	err = rows.Err()
	return err
}
