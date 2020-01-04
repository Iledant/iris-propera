package models

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
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
	stmt, err := tx.Prepare(pq.CopyIn("temp_attachment", "op_number", "coriolis_year",
		"coriolis_egt_code", "coriolis_egt_num", "coriolis_egt_line"))
	if err != nil {
		return fmt.Errorf("prepare stmt %v", err)
	}
	defer stmt.Close()
	for _, r := range o.OpFCs {
		if _, err = stmt.Exec(r.OpNumber, r.CoriolisYear, r.CoriolisEgtCode,
			r.CoriolisEgtNum, r.CoriolisEgtLine); err != nil {
			tx.Rollback()
			return fmt.Errorf("insertion de %+v  %v", r, err)
		}
	}
	if _, err = stmt.Exec(); err != nil {
		tx.Rollback()
		return fmt.Errorf("statement exec flush %v", err)
	}
	if _, err = tx.Exec(`UPDATE financial_commitment f SET physical_op_id = op.id
	FROM physical_op op, temp_attachment t WHERE op.number=t.op_number AND 
	f.coriolis_year=t.coriolis_year AND f.coriolis_egt_code=t.coriolis_egt_code AND
	f.coriolis_egt_num=t.coriolis_egt_num AND
	f.coriolis_egt_line=t.coriolis_egt_line`); err != nil {
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
