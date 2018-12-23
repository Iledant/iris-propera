package models

import (
	"database/sql"
	"strings"
)

// PrevCommitment model
type PrevCommitment struct {
	ID           int         `json:"id" gorm:"column:id"`
	PhysicalOpID int         `json:"physical_op_id" gorm:"column:physical_op_id"`
	Year         int         `json:"year" gorm:"column:year"`
	Value        int64       `json:"value" gorm:"column:value"`
	Descript     NullString  `json:"descript" gorm:"column:descript"`
	StateRatio   NullFloat64 `json:"state_ratio" gorm:"column:state_ratio"`
	TotalValue   NullInt64   `json:"total_value" gorm:"column:total_value"`
}

// PrevCommitments embeddes an array of PrevCommitment.
type PrevCommitments struct {
	PrevCommitments []PrevCommitment `json:"PrevCommitment"`
}

// PrevCommitmentLine is used to decode a line of prevision commiment batch.
type PrevCommitmentLine struct {
	Number     string      `json:"number"`
	Year       int64       `json:"year"`
	Value      int64       `json:"value"`
	TotalValue NullInt64   `json:"total_value"`
	StateRatio NullFloat64 `json:"state_ratio"`
}

// PrevCommitmentBatch embeddes an array of PrevCommitmentLine.
type PrevCommitmentBatch struct {
	PrevCommitments []PrevCommitmentLine `json:"PrevCommitment"`
}

// Save inserts and updates a batch of prevision commitments into database.
func (p *PrevCommitmentBatch) Save(db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err = tx.Exec("DROP TABLE IF EXISTS temp_prev_commitment"); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec(`CREATE TABLE temp_prev_commitment (number varchar(10), 
	year integer, value bigint, total_value bigint, state_ratio double precision)`); err != nil {
		tx.Rollback()
		return err
	}
	var value string
	var values []string
	for _, pc := range p.PrevCommitments {
		value = "(" + toSQL(pc.Number) + "," + toSQL(pc.Year) + "," + toSQL(pc.Value) +
			"," + toSQL(pc.TotalValue) + "," + toSQL(pc.StateRatio) + ")"
		values = append(values, value)
	}
	if _, err = tx.Exec(`INSERT INTO temp_prev_commitment (number,year,value,
		total_value,state_ratio) VALUES ` + strings.Join(values, ",")); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec(`UPDATE prev_commitment SET value = t.value, total_value = t.total_value, 
	state_ratio = t.state_ratio FROM temp_prev_commitment t, physical_op op
	WHERE t.number=op.number AND prev_commitment.physical_op_id = op.id AND t.year = prev_commitment.year`); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec(`INSERT INTO prev_commitment (physical_op_id, year, value, descript, total_value, state_ratio)
	SELECT op.id, t.year, t.value, NULL, t.total_value, t.state_ratio FROM physical_op op, temp_prev_commitment t
	WHERE op.number = t.number AND ((op.id, t.year) NOT IN (SELECT physical_op_id, year FROM prev_commitment))`); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec("DROP TABLE IF EXISTS temp_prev_commitment"); err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}
