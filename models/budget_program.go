package models

import (
	"database/sql"
	"errors"
)

// BudgetProgram model
type BudgetProgram struct {
	ID              int64      `json:"id"`
	CodeContract    string     `json:"code_contract"`
	CodeFunction    string     `json:"code_function"`
	CodeNumber      string     `json:"code_number"`
	CodeSubfunction NullString `json:"code_subfunction"`
	Name            string     `json:"name"`
	ChapterID       int64      `json:"chapter_id"`
}

// BudgetPrograms embeddes an array of BudgetPrograms for json export.
type BudgetPrograms struct {
	BudgetPrograms []BudgetProgram `json:"BudgetProgram"`
}

// Validate checks if fields are well formed.
func (b *BudgetProgram) Validate() error {
	if len(b.CodeContract) != 1 || b.CodeFunction == "" || len(b.CodeFunction) > 2 ||
		b.CodeNumber == "" || len(b.CodeNumber) > 3 ||
		(b.CodeSubfunction.Valid && len(b.CodeSubfunction.String) != 1) || b.Name == "" {
		return errors.New("Champ manquant ou incorrect")
	}
	return nil
}

// GetAll fetches all budget programs from database.
func (b *BudgetPrograms) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id, code_contract, code_function, code_number,
	code_subfunction, name, chapter_id FROM budget_program`)
	if err != nil {
		return err
	}
	var r BudgetProgram
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.CodeContract, &r.CodeFunction, &r.CodeNumber,
			&r.CodeSubfunction, &r.Name, &r.ChapterID); err != nil {
			return err
		}
		b.BudgetPrograms = append(b.BudgetPrograms, r)
	}
	err = rows.Err()
	return err
}

// GetAllChapterLinked fetches all budget programs linked to a chapter for json export.
func (b *BudgetPrograms) GetAllChapterLinked(chapterID int64, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id, code_contract, code_function, code_number, 
	code_subfunction, name, chapter_id FROM budget_program WHERE chapter_id=$1`, chapterID)
	if err != nil {
		return err
	}
	var r BudgetProgram
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.CodeContract, &r.CodeFunction, &r.CodeNumber,
			&r.CodeSubfunction, &r.Name, &r.ChapterID); err != nil {
			return err
		}
		b.BudgetPrograms = append(b.BudgetPrograms, r)
	}
	err = rows.Err()
	return err
}

// Create insert an budget program into database returning ID if succeed.
func (b *BudgetProgram) Create(db *sql.DB) (err error) {
	err = db.QueryRow(`INSERT INTO budget_program (code_contract, code_function, code_number, 
		code_subfunction,name, chapter_id) VALUES($1,$2,$3,$4,$5,$6) RETURNING id`, b.CodeContract,
		b.CodeFunction, b.CodeNumber, b.CodeSubfunction, b.Name, b.ChapterID).Scan(&b.ID)
	return err
}

// Update a budget program in the database. All fields are updated.
func (b *BudgetProgram) Update(db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE budget_program SET code_contract = $1, code_function = $2,
	code_number = $3, code_subfunction = $4, name = $5, chapter_id = $6
	WHERE id = $7`, b.CodeContract, b.CodeFunction, b.CodeNumber, b.CodeSubfunction, b.Name,
		b.ChapterID, b.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Programme introuvable")
	}
	return err
}

// Delete a program from database given its ID.
func (b *BudgetProgram) Delete(db *sql.DB) (err error) {
	res, err := db.Exec("DELETE FROM budget_program WHERE id = $1", b.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Programme introuvable")
	}
	return nil
}
