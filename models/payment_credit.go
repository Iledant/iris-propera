package models

import (
	"database/sql"
	"fmt"
)

// PaymentCredit model
type PaymentCredit struct {
	Year            int64 `json:"Year"`
	ChapterID       int64 `json:"ChapterID"`
	ChapterCode     int64 `json:"ChapterCode"`
	SubFunctionCode int64 `json:"SubFunctionCode"`
	PrimitiveBudget int64 `json:"PrimitiveBudget"`
	Reported        int64 `json:"Reported"`
	AddedBudget     int64 `json:"AddedBudget"`
	ModifyDecision  int64 `json:"ModifyDecision"`
	Movement        int64 `json:"Movement"`
}

// PaymentCredits embeddes an array of PaymentCredit for json export
type PaymentCredits struct {
	Lines []PaymentCredit `json:"PaymentCredit"`
}

// PaymentCreditLine is used to decode one line of PaymentCreditBatch
type PaymentCreditLine struct {
	ChapterCode     int64 `json:"ChapterCode"`
	SubFunctionCode int64 `json:"SubFunctionCode"`
	PrimitiveBudget int64 `json:"PrimitiveBudget"`
	Reported        int64 `json:"Reported"`
	AddedBudget     int64 `json:"AddedBudget"`
	ModifyDecision  int64 `json:"ModifyDecision"`
	Movement        int64 `json:"Movement"`
}

// PaymentCreditBatch embeddes an array of PaumentCreditLine for batch import
type PaymentCreditBatch struct {
	Lines []PaymentCreditLine `json:"PaymentCredit"`
}

// GetAll fetches all PaymentCredits of a year from database
func (p *PaymentCredits) GetAll(year int, db *sql.DB) error {
	rows, err := db.Query(`SELECT pc.year,bc.id,bc.code,pc.sub_function_code,
	pc.primitive_budget,pc.reported,pc.added_budget,pc.modify_decision,pc.movement
	 FROM payment_credit pc
	JOIN budget_chapter bc ON bc.id=pc.chapter_id WHERE pc.year=$1`, year)
	if err != nil {
		return err
	}
	defer rows.Close()
	var row PaymentCredit
	for rows.Next() {
		if err = rows.Scan(&row.Year, &row.ChapterID, &row.ChapterCode,
			&row.SubFunctionCode, &row.PrimitiveBudget, &row.Reported,
			&row.AddedBudget, &row.ModifyDecision, &row.Movement); err != nil {
			return err
		}
		p.Lines = append(p.Lines, row)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	if len(p.Lines) == 0 {
		p.Lines = []PaymentCredit{}
	}
	return nil
}

// Save import a batch of payment credits into database
func (p *PaymentCreditBatch) Save(year int64, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for _, l := range p.Lines {
		if _, err = tx.Exec(`INSERT INTO temp_payment_credit (chapter_code,
			sub_function_code,primitive_budget,reported,added_budget,modify_decision,
			movement) VALUES($1,$2,$3,$4,$5,$6,$7)`, l.ChapterCode, l.SubFunctionCode,
			l.PrimitiveBudget, l.Reported, l.AddedBudget, l.ModifyDecision,
			l.Movement); err != nil {
			tx.Rollback()
			return fmt.Errorf("temp insert %v", err)
		}
	}
	if _, err = tx.Exec(`UPDATE payment_credit SET year=$1,
		chapter_id=t.chapter_id,sub_function_code=t.sub_function_code,
		primitive_budget=t.primitive_budget,reported=t.reported,
		added_budget=t.added_budget,modify_decision=t.modify_decision,
		movement=t.movement
		FROM (SELECT tpc.*,bc.id as chapter_id FROM temp_payment_credit tpc
			JOIN budget_chapter bc ON bc.code=tpc.chapter_code) t 
			WHERE (t.chapter_code,t.sub_function_code) IN (SELECT chapter_code,sub_function_code 
				FROM payment_credit)`, year); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec(`INSERT INTO payment_credit 
		(SELECT $1, bc.id, tpc.sub_function_code,tpc.primitive_budget,tpc.reported,
			tpc.added_budget,tpc.modify_decision,tpc.movement
			FROM temp_payment_credit tpc
			JOIN budget_chapter bc ON bc.code=tpc.chapter_code
			WHERE (tpc.chapter_code,tpc.sub_function_code) NOT IN
			(SELECT chapter_code,sub_function_code FROM payment_credit))`, year); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec(`DELETE FROM temp_payment_credit`); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
