package models

import (
	"database/sql"
	"errors"
	"time"
)

// Commission model
type Commission struct {
	ID   int64     `json:"id"`
	Date time.Time `json:"date"`
	Name string    `json:"name"`
}

// Commissions embeddes an array of commissions for json export.
type Commissions struct {
	Commissions []Commission `json:"Commissions"`
}

// Validate checks if fields are correctly formed.
func (c *Commission) Validate() error {
	if c.Name == "" || c.Date.IsZero() {
		return errors.New("Name ou date incorrect")
	}
	return nil
}

// GetAll fetches all commissions from database.
func (c *Commissions) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query("SELECT id, date, name FROM commissions")
	if err != nil {
		return err
	}
	var r Commission
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Date, &r.Name); err != nil {
			return err
		}
		c.Commissions = append(c.Commissions, r)
	}
	err = rows.Err()
	if len(c.Commissions) == 0 {
		c.Commissions = []Commission{}
	}
	return err
}

// Create insert a new commission into database.
func (c *Commission) Create(db *sql.DB) (err error) {
	err = db.QueryRow("INSERT INTO commissions (date,name) VALUES($1,$2) RETURNING id",
		c.Date, c.Name).Scan(&c.ID)
	return err
}

// Update modifies a commission in database.
func (c *Commission) Update(db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE commissions SET date=$1, name=$2 WHERE id=$3`,
		c.Date, c.Name, c.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Commission introuvable")
	}
	return err
}

// Delete removes a commission from database.
func (c *Commission) Delete(db *sql.DB) (err error) {
	res, err := db.Exec("DELETE FROM commissions WHERE id = $1", c.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Commission introuvable")
	}
	return nil
}
