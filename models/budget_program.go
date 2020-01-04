package models

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
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

// BudgetProgramLine is used to decode one line par BudgetProgram batch.
type BudgetProgramLine struct {
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Chapter     int64      `json:"chapter"`
	Subfunction NullString `json:"subfunction"`
}

// BudgetProgramBatch embeddes an array of BudgetProgramLine for batch import.
type BudgetProgramBatch struct {
	Lines []BudgetProgramLine `json:"BudgetProgram"`
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
	if len(b.BudgetPrograms) == 0 {
		b.BudgetPrograms = []BudgetProgram{}
	}
	return err
}

// GetAllChapterLinked fetches all budget programs linked to a chapter for json export.
func (b *BudgetPrograms) GetAllChapterLinked(chapID int64, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id,code_contract,code_function,code_number, 
	code_subfunction,name,chapter_id FROM budget_program WHERE chapter_id=$1`, chapID)
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
	if len(b.BudgetPrograms) == 0 {
		b.BudgetPrograms = []BudgetProgram{}
	}
	return err
}

// Create insert an budget program into database returning ID if succeed.
func (b *BudgetProgram) Create(db *sql.DB) (err error) {
	err = db.QueryRow(`INSERT INTO budget_program (code_contract,code_function,
		code_number,code_subfunction,name,chapter_id) VALUES($1,$2,$3,$4,$5,$6)
		RETURNING id`, b.CodeContract, b.CodeFunction, b.CodeNumber,
		b.CodeSubfunction, b.Name, b.ChapterID).Scan(&b.ID)
	return err
}

// Update a budget program in the database. All fields are updated.
func (b *BudgetProgram) Update(db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE budget_program SET code_contract=$1,code_function=$2,
	code_number=$3,code_subfunction=$4,name=$5,chapter_id=$6 WHERE id = $7`,
		b.CodeContract, b.CodeFunction, b.CodeNumber, b.CodeSubfunction, b.Name,
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

// Save decodes, checks and insert into database a batch of budget programs.
func (b *BudgetProgramBatch) Save(db *sql.DB) (err error) {
	if len(b.Lines) == 0 {
		return nil
	}

	for _, r := range b.Lines {
		if len(r.Code) < 7 {
			return errors.New("Code " + r.Code + " trop court")
		}
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`DROP TABLE IF EXISTS temp_programs`); err != nil {
		tx.Rollback()
		return err
	}
	q := `CREATE TABLE temp_programs (
		code_contract varchar(1),
		code_function varchar(2),
		code_number varchar(3),
		code_subfunction varchar(1),
		name varchar(100),
		chapter integer)`
	if _, err = tx.Exec(q); err != nil {
		tx.Rollback()
		return err
	}

	stmt, err := tx.Prepare(pq.CopyIn("temp_programs", "code_contract",
		"code_function", "code_number", "code_subfunction", "name", "chapter"))
	if err != nil {
		return fmt.Errorf("prepare stmt %v", err)
	}
	defer stmt.Close()
	var subFunction string
	for _, r := range b.Lines {
		if r.Subfunction.Valid && len(r.Subfunction.String) > 2 {
			r.Subfunction.String = r.Subfunction.String[2:3]
		} else {
			r.Subfunction.Valid = false
		}
		if _, err = stmt.Exec(r.Code[0:1], r.Code[1:3], r.Code[3:6], subFunction,
			r.Name, r.Chapter); err != nil {
			tx.Rollback()
			return fmt.Errorf("insertion de %+v  %v", r, err)
		}
	}
	if _, err = stmt.Exec(); err != nil {
		tx.Rollback()
		return fmt.Errorf("statement exec flush %v", err)
	}

	queries := []string{
		`WITH new AS (SELECT p.id,t.name FROM temp_programs t, budget_program p
			WHERE p.code_contract=t.code_contract AND p.code_function=t.code_function
				AND p.code_number=t.code_number)
		UPDATE budget_program SET name=new.name FROM new WHERE budget_program.id=new.id`,
		`INSERT INTO budget_program (chapter_id,code_contract,code_function, 
		code_number,code_subfunction,name)
		SELECT c.id AS chapter_id,t.code_contract,t.code_function,t.code_number, 
			t.code_subfunction,t.name FROM temp_programs t, budget_chapter c
		WHERE c.code=t.chapter AND (t.code_contract,t.code_function,t.code_number)
		NOT IN (SELECT code_contract,code_function,code_number FROM budget_program)`,
		`DROP TABLE IF EXISTS temp_programs`}
	for _, qry := range queries {
		if _, err := tx.Exec(qry); err != nil {
			tx.Rollback()
			return err
		}
	}
	err = tx.Commit()
	return err
}
