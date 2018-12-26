package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"
)

// PendingCommitment model
type PendingCommitment struct {
	ID             int       `json:"id"`
	PhysicalOpID   NullInt64 `json:"physical_op_id"`
	IrisCode       string    `json:"iris_code"`
	Name           string    `json:"name"`
	Chapter        string    `json:"chapter"`
	ProposedValue  int64     `json:"proposed_value"`
	Action         string    `json:"action"`
	CommissionDate time.Time `json:"commission_date"`
	Beneficiary    string    `json:"beneficiary"`
}

// PendingCommitments embeddes an array of PendingCommitment.
type PendingCommitments struct {
	PendingCommitments []PendingCommitment `json:"PendingCommitments"`
}

// PendingLine is used to decode a row of array of a batch of pending commitments.
type PendingLine struct {
	Chapter        string    `json:"chapter"`
	Action         string    `json:"action"`
	IrisCode       string    `json:"iris_code"`
	Name           string    `json:"name"`
	Beneficiary    string    `json:"beneficiary"`
	CommissionDate time.Time `json:"commission_date"`
	ProposedValue  int64     `json:"proposed_value"`
}

// PendingsBatch embeddes an array of PendingLine for batch import.
type PendingsBatch struct {
	PendingsBatch []PendingLine `json:"PendingCommitment"`
}

// CompletePendingCommitment is used to decode explicit pending commitment
//linked to a physical operation for settings frontend page.
type CompletePendingCommitment struct {
	ID            int64     `json:"id"`
	PeName        string    `json:"peName"`
	PeIrisCode    string    `json:"peIrisCode"`
	PeDate        time.Time `json:"peDate"`
	PeBeneficiary string    `json:"peBeneficiary"`
	PeValue       int64     `json:"peValue"`
	OpName        string    `json:"opName"`
}

// CompletePendingCommitments embeddes an array of CompletePendingCommitment for json export.
type CompletePendingCommitments struct {
	CompletePendingCommitments []CompletePendingCommitment `json:"LinkedPendingCommitments"`
}

// PendingIDs embeddes an array of ID of pending commitments for linking or unlinking.
type PendingIDs struct {
	IDs []int64 `json:"peIdList"`
}

// GetAll fetches all pending commitments from database.
func (p *PendingCommitments) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id, physical_op_id, iris_code, name, chapter,
	 proposed_value, action, commission_date, beneficiary FROM pending_commitments`)
	if err != nil {
		return err
	}
	var r PendingCommitment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.PhysicalOpID, &r.IrisCode, &r.Name, &r.Chapter,
			&r.ProposedValue, &r.Action, &r.CommissionDate, &r.Beneficiary); err != nil {
			return err
		}
		p.PendingCommitments = append(p.PendingCommitments, r)
	}
	err = rows.Err()
	if len(p.PendingCommitments) == 0 {
		p.PendingCommitments = []PendingCommitment{}
	}
	return err
}

// GetAllUnlinked fetches all pending commitments not linked to a physical operation from database.
func (p *PendingCommitments) GetAllUnlinked(db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id, physical_op_id, iris_code, name, chapter,
	 proposed_value, action, commission_date, beneficiary FROM pending_commitments
	 WHERE physical_op_id ISNULL`)
	if err != nil {
		return err
	}
	var r PendingCommitment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.PhysicalOpID, &r.IrisCode, &r.Name, &r.Chapter,
			&r.ProposedValue, &r.Action, &r.CommissionDate, &r.Beneficiary); err != nil {
			return err
		}
		p.PendingCommitments = append(p.PendingCommitments, r)
	}
	err = rows.Err()
	if len(p.PendingCommitments) == 0 {
		p.PendingCommitments = []PendingCommitment{}
	}
	return err
}

// GetAllLinked fetches all pending commitments not linked to a physical operation from database.
func (p *PendingCommitments) GetAllLinked(db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id, physical_op_id, iris_code, name, chapter,
	 proposed_value, action, commission_date, beneficiary FROM pending_commitments
	 WHERE physical_op_id NOTNULL`)
	if err != nil {
		return err
	}
	var r PendingCommitment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.PhysicalOpID, &r.IrisCode, &r.Name, &r.Chapter,
			&r.ProposedValue, &r.Action, &r.CommissionDate, &r.Beneficiary); err != nil {
			return err
		}
		p.PendingCommitments = append(p.PendingCommitments, r)
	}
	err = rows.Err()
	if len(p.PendingCommitments) == 0 {
		p.PendingCommitments = []PendingCommitment{}
	}
	return err
}

// LinkPendings link pendings who IDs are sent to the physical operations into the database.
func (p *PhysicalOp) LinkPendings(i *PendingIDs, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	res, err := tx.Exec(`UPDATE pending_commitments SET physical_op_id = $1 WHERE id = ANY($2)`,
		p.ID, pq.Array(i.IDs))
	if err != nil {
		tx.Rollback()
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if int(count) != len(i.IDs) {
		tx.Rollback()
		return errors.New("Opération ou engagements en cours introuvables")
	}
	err = tx.Commit()
	return err
}

// Unlink remove link between pending commitments whose IDs are given and physical operation into database.
func (p *PendingCommitments) Unlink(i *PendingIDs, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	res, err := tx.Exec(`UPDATE pending_commitments SET physical_op_id = NULL WHERE id = ANY($1)`,
		pq.Array(i.IDs))
	if err != nil {
		tx.Rollback()
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if int(count) != len(i.IDs) {
		tx.Rollback()
		return errors.New("Opération ou engagements en cours introuvables")
	}
	err = tx.Commit()
	return err
}

// Save a batch of pendings commitment to the database.
func (p *PendingsBatch) Save(db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DROP TABLE IF EXISTS temp_pending`)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec(`CREATE TABLE temp_pending (
		chapter VARCHAR(5), action VARCHAR(154), iris_code VARCHAR(32),
		name VARCHAR(200), beneficiary VARCHAR(200), commission_date DATE,
		proposed_value BIGINT)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	var value string
	var values []string
	for _, pc := range p.PendingsBatch {
		value = "(" + toSQL(pc.Chapter) + ", " + toSQL(pc.Action) + ", " +
			toSQL(pc.IrisCode) + ", " + toSQL(pc.Name) + ", " + toSQL(pc.Beneficiary) +
			", " + toSQL(pc.CommissionDate) + ", " + toSQL(pc.ProposedValue) + ")"
		values = append(values, value)
	}
	_, err = tx.Exec(`INSERT INTO temp_pending (chapter, action, iris_code, name, 
		beneficiary, commission_date,proposed_value) VALUES` + strings.Join(values, ","))
	if err != nil {
		tx.Rollback()
		return
	}
	queries := []string{
		`UPDATE pending_commitments 
		SET chapter = tp.chapter, action = tp.action, name = tp.name,
				beneficiary = tp.beneficiary, commission_date = tp.commission_date,
				proposed_value = tp.proposed_value
		FROM (SELECT * FROM temp_pending) tp WHERE tp.iris_code = pending_commitments.iris_code`,
		`INSERT INTO pending_commitments 
			(physical_op_id, chapter, action, iris_code, name,  beneficiary, 
				commission_date, proposed_value)
			SELECT NULL,* FROM temp_pending 
			  WHERE iris_code NOT IN (SELECT iris_code FROM pending_commitments)`,
		`DELETE FROM pending_commitments 
			WHERE iris_code NOT IN (SELECT iris_code FROM temp_pending)`,
		`DROP TABLE IF EXISTS temp_pending`}
	for _, qry := range queries {
		_, err = tx.Exec(qry)
		if err != nil {
			tx.Rollback()
			return
		}
	}
	err = tx.Commit()
	return err
}

// GetAll fetches explicit pending commitments linked to a physical operation from database.
func (c *CompletePendingCommitments) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT pe.id, pe.name, pe.iris_code, 
	pe.commission_date, pe.beneficiary, pe.proposed_value, 
	op.number || ' - ' || op.name
FROM pending_commitments pe, physical_op op WHERE pe.physical_op_id = op.id`)
	if err != nil {
		return err
	}
	var r CompletePendingCommitment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.PeName, &r.PeIrisCode, &r.PeDate,
			&r.PeBeneficiary, &r.PeName, &r.OpName); err != nil {
			return err
		}
		c.CompletePendingCommitments = append(c.CompletePendingCommitments, r)
	}
	err = rows.Err()
	if len(c.CompletePendingCommitments) == 0 {
		c.CompletePendingCommitments = []CompletePendingCommitment{}
	}
	return err
}
