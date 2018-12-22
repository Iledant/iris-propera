package models

import (
	"database/sql"
	"errors"
)

// Beneficiary model
type Beneficiary struct {
	ID   int    `json:"id" gorm:"column:id"`
	Code int    `json:"code" gorm:"column:code"`
	Name string `json:"name" gorm:"column:name"`
}

// Beneficiaries an array of Beneficiary model with json schema
type Beneficiaries struct {
	Beneficiaries []Beneficiary `json:"Beneficiary"`
}

// Validate checks if beneficiary fields are correctly formed.
func (b *Beneficiary) Validate() error {
	if b.Name == "" {
		return errors.New("Champ name manquant")
	}
	return nil
}

// Update change the name of a beneficiary whose ID is given.
func (b *Beneficiary) Update(db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE beneficiary SET name=$1 WHERE id = $2`, b.Name, b.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Bénéficiaire introuvable")
	}
	return err
}

// GetAll fetch all beneficiaries in the database
func (b *Beneficiaries) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query("SELECT id, code, name FROM beneficiary")
	if err != nil {
		return err
	}
	defer rows.Close()
	var r Beneficiary
	for rows.Next() {
		err = rows.Scan(&r.ID, &r.Code, &r.Name)
		if err != nil {
			return err
		}
		b.Beneficiaries = append(b.Beneficiaries, r)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	return nil
}
