package models

import (
	"database/sql"
	"errors"
)

// PaymentType model
type PaymentType struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// PaymentTypes embeddes an array of PaymentType for json export.
type PaymentTypes struct {
	PaymentTypes []PaymentType `json:"PaymentType"`
}

// Validate checks if fields are well formed
func (p *PaymentType) Validate() error {
	if p.Name == "" || len(p.Name) > 255 {
		return errors.New("Name incorrect")
	}
	return nil
}

// GetAll fetches all payment types from database.
func (p *PaymentTypes) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id,name FROM payment_types`)
	if err != nil {
		return err
	}
	var r PaymentType
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Name); err != nil {
			return err
		}
		p.PaymentTypes = append(p.PaymentTypes, r)
	}
	err = rows.Err()
	if len(p.PaymentTypes) == 0 {
		p.PaymentTypes = []PaymentType{}
	}
	return err
}

// Create inserts a new payent type into database.
func (p *PaymentType) Create(db *sql.DB) (err error) {
	err = db.QueryRow("INSERT INTO payment_types (name) VALUES($1) RETURNING id",
		p.Name).Scan(&p.ID)
	return err
}

// Update modifies the payment type's name in database.
func (p *PaymentType) Update(db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE payment_types SET name = $1 WHERE id = $2`,
		p.Name, p.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Chronique de paiement introuvable")
	}
	return err
}

// Delete removes thepayment type from database.
func (p *PaymentType) Delete(db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM payment_ratios WHERE payment_types_id = $1", p.ID)
	if err != nil {
		tx.Rollback()
		return err
	}
	res, err := tx.Exec("DELETE FROM payment_types WHERE id = $1", p.ID)
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
		return errors.New("Chronique de paiement introuvable")
	}
	tx.Commit()
	return nil
}
