package models

import (
	"database/sql"
	"errors"
)

// BudgetSector model
type BudgetSector struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// BudgetSectors embeddes an array of BudgetSectors for json export.
type BudgetSectors struct {
	BudgetSectors []BudgetSector `json:"BudgetSector"`
}

// Validate check if field are correctly formed.
func (b *BudgetSector) Validate() error {
	if b.Code == "" || len(b.Code) > 10 || b.Name == "" || len(b.Name) > 100 {
		return errors.New("Code ou nom incorrect")
	}
	return nil
}

// Create save a new budget sector in the database.
func (b *BudgetSector) Create(db *sql.DB) (err error) {
	err = db.QueryRow("INSERT INTO budget_sector (code,name) VALUES($1,$2) RETURNING id",
		b.Code, b.Name).Scan(&b.ID)
	return err
}

// Update modifies a budget sector in the database.
func (b *BudgetSector) Update(db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE budget_sector SET code=$1, name=$2 WHERE id=$3`,
		b.Code, b.Name, b.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Secteur budgétaire introuvable")
	}
	return err
}

// Delete removes a budget sector from database.
func (b *BudgetSector) Delete(db *sql.DB) (err error) {
	res, err := db.Exec("DELETE FROM budget_sector WHERE id = $1", b.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Secteur budgétaire introuvable")
	}
	return nil
}

// GetAll fetches all budget sectors in the datbase.
func (b *BudgetSectors) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id,code,name FROM budget_sector`)
	if err != nil {
		return err
	}
	var r BudgetSector
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Code, &r.Name); err != nil {
			return err
		}
		b.BudgetSectors = append(b.BudgetSectors, r)
	}
	err = rows.Err()
	if len(b.BudgetSectors) == 0 {
		b.BudgetSectors = []BudgetSector{}
	}
	return err
}
