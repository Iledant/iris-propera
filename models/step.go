package models

import (
	"database/sql"
	"errors"
)

// Step model
type Step struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Steps embeddes an array of step for json export.
type Steps struct {
	Steps []Step `json:"Step"`
}

// Validate checks if fields are correctly formed.
func (s Step) Validate() error {
	if s.Name == "" || len(s.Name) > 50 {
		return errors.New("Name incorrect")
	}
	return nil
}

// Create insert a new step into database.
func (s *Step) Create(db *sql.DB) (err error) {
	err = db.QueryRow("INSERT INTO step (name) VALUES($1) RETURNING id",
		s.Name).Scan(&s.ID)
	return err
}

// Update modifies a step into database.
func (s *Step) Update(db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE step SET name=$1 WHERE id = $2`, s.Name, s.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Etape introuvable")
	}
	return err
}

// Delete remote a step from database.
func (s *Step) Delete(db *sql.DB) (err error) {
	res, err := db.Exec("DELETE FROM step WHERE id = $1", s.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Etape introuvable")
	}
	return nil
}

// GetAll fetches all steps from database.
func (s *Steps) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id,name FROM step`)
	if err != nil {
		return err
	}
	var r Step
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Name); err != nil {
			return err
		}
		s.Steps = append(s.Steps, r)
	}
	err = rows.Err()
	return err
}
