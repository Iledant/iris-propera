package models

import (
	"database/sql"
	"errors"
)

// BudgetAction model
type BudgetAction struct {
	ID        int64  `json:"id" gorm:"column:id"`
	Code      string `json:"code" gorm:"column:code"`
	Name      string `json:"name" gorm:"column:name"`
	ProgramID int64  `json:"program_id" gorm:"column:program_id"`
	SectorID  int64  `json:"sector_id" gorm:"column:sector_id"`
}

// BudgetActionLine embeddes one line of batch of budget actions.
type BudgetActionLine struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Sector string `json:"sector"`
}

// BudgetActionsBatch embeddes an array of budget action.
type BudgetActionsBatch struct {
	BudgetActions []BudgetActionLine `json:"BudgetAction"`
}

// BudgetActions embeddes an array of BudgetActions for json export.
type BudgetActions struct {
	BudgetActions []BudgetAction `json:"BudgetAction"`
}

// Validate checks if fields are correctly formed.
func (b *BudgetAction) Validate() error {
	if b.Code == "" || b.Name == "" || b.SectorID == 0 {
		return errors.New("Code, nom ou ID secteur incorrect")
	}
	return nil
}

// GetAll fetches all budget actions of database.
func (b *BudgetActions) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query("SELECT id, code, name, program_id, sector_id FROM budget_action")
	if err != nil {
		return err
	}
	var r BudgetAction
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Code, &r.Name, &r.ProgramID, &r.SectorID); err != nil {
			return err
		}
		b.BudgetActions = append(b.BudgetActions, r)
	}
	err = rows.Err()
	return err
}

// GetAllPrgID fetches all budget actions of database linked to a program ID.
func (b *BudgetActions) GetAllPrgID(pID int, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id, code, name, program_id, sector_id 
	FROM budget_action WHERE program_id = $1`, pID)
	if err != nil {
		return err
	}
	var r BudgetAction
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Code, &r.Name, &r.ProgramID, &r.SectorID); err != nil {
			return err
		}
		b.BudgetActions = append(b.BudgetActions, r)
	}
	err = rows.Err()
	return err
}

// Create insert the budget action into the database
func (b *BudgetAction) Create(db *sql.DB) (err error) {
	err = db.QueryRow(`INSERT INTO budget_action (code, name, program_id, sector_id) 
	VALUES($1,$2,$3,$4) RETURNING id`, b.Code, b.Name, b.ProgramID, b.SectorID).Scan(&b.ID)
	return err
}

// Get fetch a budget action from database by ID.
func (b *BudgetAction) Get(ID int, db *sql.DB) (err error) {
	err = db.QueryRow(`SELECT id, code, name, program_id, sector_id 
	FROM budget_action WHERE id = $1`, ID).Scan(&b.ID, &b.Code, &b.Name, &b.ProgramID, &b.SectorID)
	return err
}

// Update a budget action in database.
func (b *BudgetAction) Update(db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE budget_action SET code = $1, name = $2, program_id = $3, sector_id = $4
	 WHERE id = $5`, b.Code, b.Name, b.ProgramID, b.SectorID, b.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Action introuvable")
	}
	return err
}

// Delete remove budget action whose ID is given from database.
func (b *BudgetAction) Delete(db *sql.DB) (err error) {
	res, err := db.Exec("DELETE FROM budget_action WHERE id = $1", b.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Action budgétaire introuvable")
	}
	return nil
}

// Save a batch of budget actions to database.
func (b *BudgetActionsBatch) Save(db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`DROP TABLE IF EXISTS temp_actions`); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec(`CREATE TABLE temp_actions (code_contract VARCHAR(1), 
	code_function VARCHAR(2), code_number VARCHAR(3), action_code VARCHAR(4), 
	name VARCHAR(255),sector VARCHAR(10))`); err != nil {
		tx.Rollback()
		return err
	}
	for _, ba := range b.BudgetActions {
		if len(ba.Code) < 7 {
			tx.Rollback()
			return errors.New("Erreur lors de l'import, code trop court :" + ba.Code)
		}
		cc, cf, cn, ac := ba.Code[0:1], ba.Code[1:3], ba.Code[3:6], ba.Code[6:]
		if _, err = tx.Exec(`INSERT INTO temp_actions (code_contract, code_function, 
			code_number, action_code, name, sector) VALUES ($1, $2, $3, $4, $5, $6)`,
			cc, cf, cn, ac, ba.Name, ba.Sector); err != nil {
			tx.Rollback()
			return err
		}
	}
	if _, err = tx.Exec(`WITH new AS (
		SELECT a.id, t.name FROM temp_actions t, budget_program p, budget_action a
		WHERE t.action_code=a.code AND t.code_contract=p.code_contract AND
					t.code_function=p.code_function AND t.code_number=p.code_number AND a.program_id=p.id)
	UPDATE budget_action SET name = new.name
	FROM new WHERE budget_action.id = new.id`); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec(`INSERT INTO budget_action (program_id, sector_id, code, name) 
	SELECT p.id AS program_id, s.id AS sector_id, t.action_code, t.name FROM temp_actions t
		LEFT JOIN budget_sector s ON s.code = t.sector
		LEFT JOIN budget_program p ON ( p.code_contract = t.code_contract AND
																		p.code_function = t.code_function AND
																		p.code_number = t.code_number)
	WHERE (s.id, p.id, t.action_code) NOT IN (SELECT sector_id, program_id, code FROM budget_action) 
		AND p.id NOTNULL`); err != nil {
		tx.Rollback()
		return err
	}
	tx.Exec(`DROP TABLE IF EXISTS temp_actions`)
	tx.Commit()
	return nil
}
