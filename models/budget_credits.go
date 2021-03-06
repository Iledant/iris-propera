package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// BudgetCredit model
type BudgetCredit struct {
	ID                 int64     `json:"id"`
	CommissionDate     NullTime  `json:"commission_date"`
	ChapterID          NullInt64 `json:"chapter_id"`
	PrimaryCommitment  int64     `json:"primary_commitment"`
	FrozenCommitment   int64     `json:"frozen_commitment"`
	ReservedCommitment int64     `json:"reserved_commitment"`
}

// BudgetCredits embeddes an array of BudgetCredit for json export.
type BudgetCredits struct {
	BudgetCredits []BudgetCredit `json:"BudgetCredits"`
}

// CompleteBudgetCredit is used to decode budget credits with full chapter name.
type CompleteBudgetCredit struct {
	ID                 int64     `json:"id"`
	CommissionDate     time.Time `json:"commission_date"`
	Chapter            int64     `json:"chapter"`
	PrimaryCommitment  int64     `json:"primary_commitment"`
	FrozenCommitment   int64     `json:"frozen_commitment"`
	ReservedCommitment int64     `json:"reserved_commitment"`
}

// CompleteBudgetCredits embeddes an array of CompleteBudgetCredit for batch import.
type CompleteBudgetCredits struct {
	CompleteBudgetCredits []CompleteBudgetCredit `json:"BudgetCredits"`
}

// BudgetCreditLine is used to decode budget credits batch.
type BudgetCreditLine struct {
	ID                 int64     `json:"id"`
	CommissionDate     ExcelDate `json:"commission_date"`
	Chapter            int64     `json:"chapter"`
	PrimaryCommitment  float64   `json:"primary_commitment"`
	FrozenCommitment   float64   `json:"frozen_commitment"`
	ReservedCommitment float64   `json:"reserved_commitment"`
}

// BudgetCreditBatch embeddes an array of BudgetCreditLine
// to decode budget credits batch.
type BudgetCreditBatch struct {
	Lines []BudgetCreditLine `json:"BudgetCredits"`
}

// Validate checks if fields are correctly formed.
func (c *CompleteBudgetCredit) Validate() error {
	if c.Chapter == 0 || c.CommissionDate.IsZero() {
		return errors.New("Erreur de chapitre ou de date de commission")
	}
	return nil
}

// GetAll fetches all budget credits from database.
func (b *BudgetCredits) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id, commission_date, chapter_id, primary_commitment, 
	frozen_commitment, reserved_commitment FroM budget_credits`)
	if err != nil {
		return err
	}
	var r BudgetCredit
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.CommissionDate, &r.ChapterID, &r.PrimaryCommitment,
			&r.FrozenCommitment, &r.ReservedCommitment); err != nil {
			return err
		}
		b.BudgetCredits = append(b.BudgetCredits, r)
	}
	err = rows.Err()
	return err
}

// GetLatest fetches the latest budget credits for all chapters.
func (b *BudgetCredits) GetLatest(year int64, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id, commission_date, chapter_id, primary_commitment, 
	frozen_commitment, reserved_commitment FROM budget_credits WHERE commission_date = 
	(SELECT max(commission_date) FROM budget_credits WHERE EXTRACT (year FROM commission_date) = $1)`, year)
	if err != nil {
		return err
	}
	var r BudgetCredit
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.CommissionDate, &r.ChapterID, &r.PrimaryCommitment,
			&r.FrozenCommitment, &r.ReservedCommitment); err != nil {
			return err
		}
		b.BudgetCredits = append(b.BudgetCredits, r)
	}
	err = rows.Err()
	if len(b.BudgetCredits) == 0 {
		b.BudgetCredits = []BudgetCredit{}
	}
	return err
}

// GetAll fetches all budget credits with complete chapter number from database.
func (c *CompleteBudgetCredits) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT bc.id, bc.commission_date, c.code AS chapter, 
	bc.primary_commitment, bc.frozen_commitment, bc.reserved_commitment
	FROM budget_credits bc, budget_chapter c
	WHERE bc.chapter_id = c.id`)
	if err != nil {
		return err
	}
	var r CompleteBudgetCredit
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.CommissionDate, &r.Chapter, &r.PrimaryCommitment,
			&r.FrozenCommitment, &r.ReservedCommitment); err != nil {
			return err
		}
		c.CompleteBudgetCredits = append(c.CompleteBudgetCredits, r)
	}
	err = rows.Err()
	if len(c.CompleteBudgetCredits) == 0 {
		c.CompleteBudgetCredits = []CompleteBudgetCredit{}
	}
	return err
}

// Create insert a new line of budget credits with datas stored in CompleteBudgetCredit.
func (c *CompleteBudgetCredit) Create(db *sql.DB) (err error) {
	err = db.QueryRow(`INSERT INTO budget_credits (commission_date, chapter_id,
		primary_commitment, frozen_commitment, reserved_commitment) 
		SELECT $1,id,$2,$3,$4 FROM budget_chapter WHERE code = $5 RETURNING id`,
		c.CommissionDate, c.PrimaryCommitment, c.FrozenCommitment, c.ReservedCommitment,
		c.Chapter).Scan(&c.ID)
	return err
}

// Update modifies a budget credits line using datas stores in a CompleteBudgetCredit.
func (c *CompleteBudgetCredit) Update(db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE budget_credits SET (commission_date, chapter_id,
		primary_commitment, frozen_commitment, reserved_commitment) = 
		(SELECT $1::date,id,$2::bigint,$3::bigint,$4::bigint 
			FROM budget_chapter WHERE code = $5) WHERE id = $6`,
		c.CommissionDate, c.PrimaryCommitment, c.FrozenCommitment,
		c.ReservedCommitment, c.Chapter, c.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("budget_credits introuvable")
	}
	return nil
}

// Delete remove the budget credits line whose ID is given from database.
func (b *BudgetCredit) Delete(db *sql.DB) (err error) {
	res, err := db.Exec("DELETE FROM budget_credits WHERE id = $1", b.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Crédits introuvables")
	}
	return nil
}

// Save update or insert a batch of budget credits lines into database.
func (b *BudgetCreditBatch) Save(db *sql.DB) error {
	for _, r := range b.Lines {
		if r.CommissionDate == 0 || r.Chapter == 0 {
			return errors.New("Date de commission ou chapitre incorrect")
		}

	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`DROP TABLE IF EXISTS temp_budget_credits`); err != nil {
		tx.Rollback()
		return err
	}
	q := `CREATE TABLE temp_budget_credits 
	(	commission_date date,
		chapter integer CHECK (chapter > 0), 
		primary_commitment bigint,
		frozen_commitment bigint, 
		reserved_commitment bigint)`
	if _, err = tx.Exec(q); err != nil {
		tx.Rollback()
		return fmt.Errorf("create temp table %v", err)
	}
	stmt, err := tx.Prepare(pq.CopyIn("temp_budget_credits", "commission_date",
		"chapter", "primary_commitment", "frozen_commitment", "reserved_commitment"))
	if err != nil {
		return fmt.Errorf("prepare stmt %v", err)
	}
	defer stmt.Close()
	for _, r := range b.Lines {
		if _, err = stmt.Exec(r.CommissionDate.ToDate(), r.Chapter,
			int64(r.PrimaryCommitment*100), int64(100*r.FrozenCommitment),
			int64(100*r.ReservedCommitment)); err != nil {
			tx.Rollback()
			return fmt.Errorf("insertion de %+v  %v", r, err)
		}
	}
	if _, err = stmt.Exec(); err != nil {
		tx.Rollback()
		return fmt.Errorf("statement exec flush %v", err)
	}
	if _, err = tx.Exec(`INSERT INTO budget_credits (commission_date,chapter_id,
		primary_commitment,frozen_commitment,reserved_commitment)
	SELECT t.commission_date,bc.id,t.primary_commitment,t.frozen_commitment,
		t.reserved_commitment
	FROM temp_budget_credits t
	LEFT JOIN budget_chapter bc ON t.chapter=bc.code
	WHERE (t.commission_date,t.chapter) NOT IN
	(SELECT b.commission_date,c.code
		FROM budget_credits b, budget_chapter c WHERE b.chapter_id = c.id)`); err != nil {
		tx.Rollback()
		return fmt.Errorf("insert query %v", err)
	}
	if _, err = tx.Exec(`DROP TABLE IF EXISTS temp_budget_credits`); err != nil {
		tx.Rollback()
		return fmt.Errorf("drop temp table %v", err)
	}
	err = tx.Commit()
	return err
}
