package models

import (
	"database/sql"
	"errors"
	"time"
)

// Plan model
type Plan struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Descript  NullString `json:"descript"`
	FirstYear NullInt64  `json:"first_year"`
	LastYear  NullInt64  `json:"last_year"`
}

// Plans embeddes an array of Plan for json exports.
type Plans struct {
	Plans []Plan `json:"Plan"`
}

// Validate checks if fields are correctly formed.
func (p *Plan) Validate() error {
	if p.Name == "" || len(p.Name) > 255 {
		return errors.New("Name incorrect")
	}
	return nil
}

// GetAll fetches all plans from database.
func (p *Plans) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id,name, descript,first_year,last_year FROM plan`)
	if err != nil {
		return err
	}
	var r Plan
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Name, &r.Descript, &r.FirstYear, &r.LastYear); err != nil {
			return err
		}
		p.Plans = append(p.Plans, r)
	}
	err = rows.Err()
	if len(p.Plans) == 0 {
		p.Plans = []Plan{}
	}
	return err
}

// Create inserts a new plan into database.
func (p *Plan) Create(db *sql.DB) (err error) {
	err = db.QueryRow(`INSERT INTO plan (name, descript,first_year,last_year) 
	VALUES($1,$2,$3,$4) RETURNING id`, p.Name, p.Descript, p.FirstYear, p.LastYear).Scan(&p.ID)
	return err
}

// Update modifies a plan into the database.
func (p *Plan) Update(db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE plan SET name=$1, descript=$2, first_year=$3, 
	last_year=$4 WHERE id = $5`, p.Name, p.Descript, p.FirstYear, p.LastYear, p.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Plan introuvable")
	}
	return err
}

// Delete removes a plan from database.
func (p *Plan) Delete(db *sql.DB) (err error) {
	tx, err := db.Begin()
	if _, err = tx.Exec(`DELETE from plan_line_ratios WHERE plan_line_id IN 
	(SELECT id FROM plan_line WHERE plan_id = $1)`, p.ID); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec("DELETE from plan_line WHERE plan_id = $1", p.ID); err != nil {
		tx.Rollback()
		return
	}
	res, err := db.Exec("DELETE FROM plan WHERE id = $1", p.ID)
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
		return errors.New("Plan introuvable")
	}
	tx.Commit()
	return nil
}

// GetByID fetch a plan from database using it's ID.
func (p *Plan) GetByID(db *sql.DB) (err error) {
	err = db.QueryRow(`SELECT id, name, descript, first_year, last_year FROM plan
	WHERE id = $1`, p.ID).Scan(&p.ID, &p.Name, &p.Descript, &p.FirstYear, &p.LastYear)
	return err
}

// GetFirstAndLastYear computes first and last year for previsions according either
// to plan's field or to actual year and previsions
func (p *Plan) GetFirstAndLastYear(db *sql.DB) (firstYear int64, lastYear int64, err error) {
	firstYear = int64(time.Now().Year() + 1)
	if p.FirstYear.Valid && p.FirstYear.Int64 > firstYear {
		firstYear = p.FirstYear.Int64
	}
	if p.LastYear.Valid {
		lastYear = p.LastYear.Int64
	} else {
		if err = db.QueryRow("SELECT max(year) FROM prev_commitment").
			Scan(&lastYear); err != nil {
			return 0, 0, err
		}
	}
	return firstYear, lastYear, nil
}
