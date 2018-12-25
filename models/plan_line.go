package models

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

// PlanLine model
type PlanLine struct {
	ID         int64      `json:"id"`
	PlanID     int64      `json:"plan_id"`
	Name       string     `json:"name"`
	Descript   NullString `json:"descript"`
	Value      int64      `json:"value"`
	TotalValue NullInt64  `json:"total_value"`
}

// PlanLines embeddes an array of PlanLine for json export.
type PlanLines struct {
	PlanLines []PlanLine `json:"PlanLine"`
}

// LinkFCs updates the financial commitments linked
// to a physical operation in database.
func (p *PlanLine) LinkFCs(fcIDs []int64, db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE financial_commitment SET plan_line_id = $1 
	WHERE id = ANY($2)`, p.ID, pq.Array(fcIDs))
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != int64(len(fcIDs)) {
		return errors.New("Ligne de plan ou engagements incorrects")
	}
	return nil
}

// Delete removes the plan lines from database including linked plan line ratios.
func (p *PlanLine) Delete(db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err = tx.Exec("DELETE FROM plan_line_ratios WHERE plan_line_id = $1",
		p.ID); err != nil {
		tx.Rollback()
		return err
	}
	res, err := tx.Exec("DELETE FROM plan_line WHERE id=$1", p.ID)
	if err != nil {
		tx.Rollback()
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if count != 1 {
		tx.Rollback()
		return errors.New("Ligne de plan introuvable")
	}
	tx.Commit()
	return err
}

// Create insert a new plan line and it's linked ratios into database.
func (p *PlanLine) Create(plr *PlanLineRatios, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	err = tx.QueryRow(`INSERT INTO plan_line (plan_id,name,descript,value,total_value) 
	VALUES($1,$2,$3,$4,$5) RETURNING id`, p.PlanID, p.Name, p.Descript,
		p.Value, p.TotalValue).Scan(&p.ID)
	if err != nil {
		tx.Rollback()
		return err
	}
	if err = plr.Save(p.ID, tx); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return err
}

// GetByID fetches the plan line whose ID is given from database.
func (p *PlanLine) GetByID(db *sql.DB) (err error) {
	err = db.QueryRow(`SELECT id, plan_id, name, descript, value, total_value 
	FROM plan_line WHERE id=$1`, p.ID).Scan(&p.ID, &p.PlanID, &p.Name, &p.Descript,
		&p.Value, &p.TotalValue)
	return err
}

// Update modifies a plan line and it's ratio into the database.
func (p *PlanLine) Update(plr *PlanLineRatios, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	res, err := tx.Exec(`UPDATE plan_line SET plan_id=$1,name=$2,descript=$3,
	value=$4,total_value=$5 WHERE id=$6`, p.PlanID, p.Name, p.Descript, p.Value,
		p.TotalValue, p.ID)
	if err != nil {
		tx.Rollback()
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if count != 1 {
		tx.Rollback()
		return errors.New("Ligne de plan introuvable")
	}
	if err = plr.Save(p.ID, tx); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return err
}
