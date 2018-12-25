package models

import (
	"database/sql"
	"strings"
)

// OpFCLine embeddes a line of operation / commitment link batch request.
type OpFCLine struct {
	OpNumber        string `json:"op_number"`
	CoriolisYear    string `json:"coriolis_year"`
	CoriolisEgtCode string `json:"coriolis_egt_code"`
	CoriolisEgtNum  string `json:"coriolis_egt_num"`
	CoriolisEgtLine string `json:"coriolis_egt_line"`
}

// OpFCsBatch embeddes datas sent by a operation / commitments link request.
type OpFCsBatch struct {
	OpFCs []OpFCLine `json:"Attachment"`
}

// Save a batch of link between operations and  commitments to the database.
func (o *OpFCsBatch) Save(db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err = tx.Exec("DELETE from temp_attachment"); err != nil {
		tx.Rollback()
		return err
	}
	var values []string
	var value string
	for _, o := range o.OpFCs {
		value = "(" + toSQL(o.OpNumber) + ", " + toSQL(o.CoriolisYear) + ", " +
			toSQL(o.CoriolisEgtCode) + ", " + toSQL(o.CoriolisEgtNum) + ", " +
			toSQL(o.CoriolisEgtLine) + ")"
		values = append(values, value)
	}
	if _, err = tx.Exec(`INSERT INTO temp_attachment (op_number, coriolis_year,
		coriolis_egt_code, coriolis_egt_num, coriolis_egt_line) VALUES ` +
		strings.Join(values, ",")); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec(`UPDATE financial_commitment SET physical_op_id = op.id
	FROM physical_op op, temp_attachment WHERE op.number = temp_attachment.op_number AND 
	financial_commitment.coriolis_year=temp_attachment.coriolis_year AND 
	financial_commitment.coriolis_egt_code =temp_attachment.coriolis_egt_code AND
	financial_commitment.coriolis_egt_num=temp_attachment.coriolis_egt_num AND
	financial_commitment.coriolis_egt_line=temp_attachment.coriolis_egt_line`); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec("DELETE from temp_attachment"); err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}
