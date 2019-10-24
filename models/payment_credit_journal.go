package models

import (
	"database/sql"
	"fmt"
	"time"
)

// PaymentCreditJournal model
type PaymentCreditJournal struct {
	Chapter           int64     `json:"Chapter"`
	ID                int64     `json:"ID"`
	Function          int64     `json:"Function"`
	CreationDate      time.Time `json:"CreationDate"`
	ModificationFDate time.Time `json:"ModificationDate"`
	Name              string    `json:"Name"`
	Value             int64     `json:"Value"`
}

// PaymentCreditJournals embeddes an array of PaymentCreditJournal for json export
type PaymentCreditJournals struct {
	Lines []PaymentCreditJournal `json:"PaymentCreditJournal"`
}

// PaymentCreditJournalLine is used to decode one line of a batch
type PaymentCreditJournalLine struct {
	Chapter          int64  `json:"Chapter"`
	Function         int64  `json:"Function"`
	CreationDate     int64  `json:"CreationDate"`
	ModificationDate int64  `json:"ModificationDate"`
	Name             string `json:"Name"`
	Value            int64  `json:"Value"`
}

// PaymentCreditJournalBatch embeddes an array of PaymentCreditJournalLine
// for import into database
type PaymentCreditJournalBatch struct {
	Lines []PaymentCreditJournalLine `json:"PaymentCreditJournal"`
}

// GetAll fetches all payment credits journal entries of a given year
func (p *PaymentCreditJournals) GetAll(year int, db *sql.DB) error {
	rows, err := db.Query(`SELECT pcj.id,bc.code,pcj.function,pcj.creation_date,
	pcj.modification_date,pcj.name,pcj.value FROM payment_credit_journal pcj
	JOIN budget_chapter bc ON bc.id=pcj.chapter_id
	WHERE extract(year FROM creation_date)=$1`, year)
	if err != nil {
		return fmt.Errorf("select %v", err)
	}
	defer rows.Close()
	var row PaymentCreditJournal
	for rows.Next() {
		if err = rows.Scan(&row.ID, &row.Chapter, &row.Function, &row.CreationDate,
			&row.ModificationFDate, &row.Name, &row.Value); err != nil {
			return fmt.Errorf("scan %v", err)
		}
		p.Lines = append(p.Lines, row)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("rows err %v", err)
	}
	if len(p.Lines) == 0 {
		p.Lines = []PaymentCreditJournal{}
	}
	return nil
}

// Save import a batch of payment credit journal entries into database
func (p *PaymentCreditJournalBatch) Save(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM payment_credit_journal 
	WHERE extract(year FROM creation_date)=extract(year FROM CURRENT_DATE)`); err != nil {
		tx.Rollback()
		return fmt.Errorf("initial delete %v", err)
	}
	var c, m time.Time
	for _, l := range p.Lines {
		c = time.Date(int(l.CreationDate/10000), time.Month(l.CreationDate/100%100),
			int(l.CreationDate%100), 0, 0, 0, 0, time.UTC)
		m = time.Date(int(l.ModificationDate/10000),
			time.Month(l.ModificationDate/100%100), int(l.ModificationDate%100), 0, 0,
			0, 0, time.UTC)
		if _, err = tx.Exec(`INSERT INTO temp_payment_credit_journal (chapter,
			function,creation_date,modification_date,name,value)
			VALUES($1,$2,$3,$4,$5,$6)`, l.Chapter, l.Function, c, m, l.Name, l.Value); err != nil {
			tx.Rollback()
			return fmt.Errorf("temp insert %v", err)
		}
	}
	if _, err = tx.Exec(`INSERT INTO payment_credit_journal (chapter_id,function,
		creation_date,modification_date,name,value)
		(SELECT bc.id, tpc.function,tpc.creation_date,tpc.modification_date,
			tpc.name,tpc.value
			FROM temp_payment_credit_journal tpc
			JOIN budget_chapter bc ON bc.code=tpc.chapter)`); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec(`DELETE FROM temp_payment_credit_journal`); err != nil {
		tx.Rollback()
		return fmt.Errorf("final delete %v", err)
	}
	tx.Commit()
	return nil
}
