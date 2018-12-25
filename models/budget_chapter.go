package models

import (
	"database/sql"
	"errors"
)

// BudgetChapter model
type BudgetChapter struct {
	ID   int64  `json:"id"`
	Code int    `json:"code"`
	Name string `json:"name"`
}

// BudgetChapters embeddes an array of budget chapters to json export
type BudgetChapters struct {
	BudgetChapters []BudgetChapter `json:"BudgetChapter"`
}

// Validate checks if fields are correctly formed.
func (b *BudgetChapter) Validate() error {
	if b.Name == "" || len(b.Name) > 100 || b.Code == 0 {
		return errors.New("Name manquant ou trop long ou code absent")
	}
	return nil
}

// GetAll fetches all budget chapters in database.
func (b *BudgetChapters) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query("SELECT id, code, name FROM budget_chapter")
	if err != nil {
		return err
	}
	var r BudgetChapter
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Code, &r.Name); err != nil {
			return err
		}
		b.BudgetChapters = append(b.BudgetChapters, r)
	}
	err = rows.Err()
	return err
}

// Create add data sent to database.
func (b *BudgetChapter) Create(db *sql.DB) (err error) {
	err = db.QueryRow("INSERT INTO budget_chapter (code, name) VALUES($1,$2) RETURNING id",
		b.Code, b.Name).Scan(&b.ID)
	return err
}

// Update a budget chapter in database.
func (b *BudgetChapter) Update(db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE budget_chapter SET code = $1, name = $2 WHERE id = $3`, b.Code, b.Name, b.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Chapitre budgétaire introuvable")
	}
	return err
}

// Delete remove budget chapter whose ID is given from database.
func (b *BudgetChapter) Delete(db *sql.DB) (err error) {
	res, err := db.Exec("DELETE FROM budget_chapter WHERE id = $1", b.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Chapitre budgétaire introuvable")
	}
	return nil
}
