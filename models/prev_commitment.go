package models

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

// PrevCommitment model
type PrevCommitment struct {
	ID           int         `json:"id"`
	PhysicalOpID int         `json:"physical_op_id"`
	Year         int         `json:"year"`
	Value        int64       `json:"value"`
	Descript     NullString  `json:"descript"`
	StateRatio   NullFloat64 `json:"state_ratio"`
	TotalValue   NullInt64   `json:"total_value"`
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

// PrevCommitmentTotal is used to calculate the value of prevision commitment
// of a given year for others queries with duplicated prevision commitment lines
type PrevCommitmentTotal struct {
	PrevCommitmentTotal int64 `json:"PrevCommitmentTotal"`
}

// Save inserts and updates a batch of prevision commitments into database.
func (p *PrevCommitmentBatch) Save(db *sql.DB) (err error) {
	for _, pc := range p.PrevCommitments {
		if pc.Number == "" {
			return errors.New("Numéro d'opération vide")
		}
		if pc.Year == 0 {
			return errors.New("Année de prévision non renseignée")
		}
		if pc.Value == 0 {
			return errors.New("Prévision nulle")
		}
	}
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

	stmt, err := tx.Prepare(pq.CopyIn("temp_prev_commitment", "number", "year", "value",
		"total_value", "state_ratio"))
	if err != nil {
		return fmt.Errorf("prepare stmt %v", err)
	}
	defer stmt.Close()
	for _, r := range p.PrevCommitments {
		if _, err = stmt.Exec(r.Number, r.Year, r.Value, r.TotalValue, r.StateRatio); err != nil {
			tx.Rollback()
			return fmt.Errorf("insertion de %+v  %v", r, err)
		}
	}
	if _, err = stmt.Exec(); err != nil {
		tx.Rollback()
		return fmt.Errorf("statement exec flush %v", err)
	}

	if _, err = tx.Exec(`UPDATE prev_commitment SET value=t.value, total_value=t.total_value, 
	state_ratio=t.state_ratio FROM temp_prev_commitment t, physical_op op
	WHERE t.number=op.number AND prev_commitment.physical_op_id = op.id AND
	t.year = prev_commitment.year`); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec(`INSERT INTO prev_commitment (physical_op_id, year, value,
		descript, total_value, state_ratio)
	SELECT op.id, t.year, t.value, NULL, t.total_value, t.state_ratio 
		FROM physical_op op, temp_prev_commitment t
	WHERE op.number = t.number AND 
		((op.id, t.year) NOT IN (SELECT physical_op_id, year FROM prev_commitment))`); err != nil {
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

// Get calculted the prevision commitment total of the given year
func (p *PrevCommitmentTotal) Get(year int64, db *sql.DB) error {
	return db.QueryRow(`SELECT coalesce(sum(value),0) FROM prev_commitment 
		WHERE year=$1`, year).Scan(&p.PrevCommitmentTotal)
}
