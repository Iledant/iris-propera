package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

// PhysicalOp is the model for physical operations. Number is unique.
type PhysicalOp struct {
	ID             int64      `json:"id"`
	Number         string     `json:"number"`
	Name           string     `json:"name"`
	Descript       NullString `json:"descript"`
	Isr            bool       `json:"isr"`
	Value          NullInt64  `json:"value"`
	ValueDate      NullTime   `json:"valuedate"`
	Length         NullInt64  `json:"length"`
	TRI            NullInt64  `json:"tri"`
	VAN            NullInt64  `json:"van"`
	BudgetActionID NullInt64  `json:"budget_action_id"`
	PaymentTypeID  NullInt64  `json:"payment_types_id"`
	PlanLineID     NullInt64  `json:"plan_line_id"`
	StepID         NullInt64  `json:"step_id"`
	CategoryID     NullInt64  `json:"category_id"`
}

// PhysicalOps embeddes an array of PhysicalOp.
type PhysicalOps struct {
	PhysicalOps []PhysicalOp `json:"PhysicalOp"`
}

// OpPending embeddes the pending value attached to a physical operation.
type OpPending struct {
	Value NullInt64 `json:"value"`
}

// OpPendings embeddes an array of OpPendings for json export.
type OpPendings struct {
	OpPendings []OpPending `json:"PendingCommitment"`
}

// PhysicalOpLine is used to decode request for an upload of a batch of physical operations.
// The struct uses pointer for optional fields.
type PhysicalOpLine struct {
	Number        string        `json:"number"`
	Name          string        `json:"name"`
	Descript      NullString    `json:"descript"`
	Isr           bool          `json:"isr"`
	Value         NullInt64     `json:"value"`
	Valuedate     NullExcelDate `json:"valuedate"`
	Length        NullInt64     `json:"length"`
	Step          NullString    `json:"step"`
	Category      NullString    `json:"category"`
	TRI           NullInt64     `json:"tri"`
	VAN           NullInt64     `json:"van"`
	Action        NullString    `json:"action"`
	PaymentTypeID NullInt64     `json:"payment_types_id"`
	PlanLineID    NullInt64     `json:"plan_line_id"`
}

// PhysicalOpsBatch embeddes an array of PhysicalOpLine for upload.
type PhysicalOpsBatch struct {
	PhysicalOps []PhysicalOpLine `json:"PhysicalOp"`
}

// OpCommitment is used to decode the query of commitments prevision of a physical operation.
type OpCommitment struct {
	ID          int64     `json:"id"`
	Date        time.Time `json:"date"`
	IrisCode    string    `json:"iris_code"`
	Name        string    `json:"name"`
	Beneficiary string    `json:"beneficiary"`
	Value       int64     `json:"value"`
	LapseDate   NullTime  `json:"lapse_date"`
	Available   int64     `json:"available"`
}

// OpCommitments embeddes an array of OpCommitment for json export.
type OpCommitments struct {
	OpCommitments []OpCommitment `json:"FinancialCommitment"`
}

// OpPrevCommitment is used to decode prevision commitments attached to a physical operation.
type OpPrevCommitment struct {
	Year       int64       `json:"year"`
	Value      int64       `json:"value"`
	Descript   NullString  `json:"descript"`
	TotalValue NullInt64   `json:"total_value"`
	StateRatio NullFloat64 `json:"state_ratio"`
}

// OpPrevPayment is used to decode prevision payments attached to a physical operation.
type OpPrevPayment struct {
	Year     int64      `json:"year"`
	Value    int64      `json:"value"`
	Descript NullString `json:"descript"`
}

// OpPrevisions embedded two arrays of prevision commitments and prevision payments attached
// a physical operation
type OpPrevisions struct {
	Commitments []OpPrevCommitment `json:"PrevCommitment"`
	Payments    []OpPrevPayment    `json:"PrevPayment"`
}

// Validate checks if fields are correctly formed.
func (op *PhysicalOp) Validate() error {
	if len(op.Number) != 7 || op.Name == "" {
		return errors.New("Number ou Name incorrect")
	}
	return nil
}

// Get fetches a physical operation from database using ID.
func (op *PhysicalOp) Get(db *sql.DB) (err error) {
	err = db.QueryRow(`SELECT id, number, name, descript, isr, value, valuedate, length,
	tri, van, budget_action_id, payment_types_id, plan_line_id, step_id, category_id	
	FROM physical_op WHERE id = $1`, op.ID).
		Scan(&op.ID, &op.Number, &op.Name, &op.Descript, &op.Isr, &op.Value,
			&op.ValueDate, &op.Length, &op.TRI, &op.VAN, &op.BudgetActionID,
			&op.PaymentTypeID, &op.PlanLineID, &op.StepID, &op.CategoryID)
	return err
}

// Exists check of sent physical operation ID exists in the database.
func (op *PhysicalOp) Exists(db *sql.DB) (err error) {
	var count int64
	err = db.QueryRow(`SELECT count(1) from physical_op WHERE id = $1`, op.ID).Scan(&count)
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Opération introuvable")
	}
	return nil
}

// Create insert a new physical operation into database checking number.
func (op *PhysicalOp) Create(db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	var count int64
	if err = tx.QueryRow("SELECT count(1) FROM physical_op WHERE number = $1",
		op.Number).Scan(&count); err != nil {
		tx.Rollback()
		return err
	}
	if count > 0 {
		opNumPattern := op.Number[0:4] + "%"
		var lastOpNum string
		if err := tx.QueryRow(`SELECT number FROM physical_op WHERE number ILIKE $1 
		ORDER BY number DESC LIMIT 1`, opNumPattern).Scan(&lastOpNum); err != nil {
			tx.Rollback()
			return err
		}
		newOpNum, err := strconv.Atoi(lastOpNum[4:])
		if err != nil {
			tx.Rollback()
			return err
		}
		op.Number = fmt.Sprintf("%s%03d", op.Number[0:4], newOpNum+1)
	}
	err = db.QueryRow(`INSERT INTO physical_op (number, name, descript, isr, value, 
		valuedate, length, tri, van, budget_action_id, payment_types_id, plan_line_id, 
		step_id, category_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING id`,
		op.Number, op.Name, op.Descript, op.Isr, op.Value, op.ValueDate, op.Length,
		op.TRI, op.VAN, op.BudgetActionID, op.PaymentTypeID, op.PlanLineID, op.StepID,
		op.CategoryID).Scan(&op.ID)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return err
}

// Update modifies a physical operation in the database.
func (op *PhysicalOp) Update(uID int64, db *sql.DB) (err error) {
	if uID != 0 {
		var count int64
		if err = db.QueryRow(`SELECT count(1) FROM rights WHERE users_id=$1 AND physical_op_id=$2`,
			uID, op.ID).Scan(&count); err != nil {
			return err
		}
		if count == 0 {
			return errors.New("Droits insuffisant pour l'opération")
		}
	}
	var res sql.Result
	if uID == 0 {
		var opID int64
		if err = db.QueryRow(`SELECT id FROM physical_op WHERE number = $1`,
			op.Number).Scan(&opID); err != nil {
			return err
		}
		if opID != op.ID {
			return errors.New("Numéro d'opération existant")
		}
		res, err = db.Exec(`UPDATE physical_op SET number=$1, name=$2, descript=$3,
	isr=$4, value=$5, valuedate=$6, length=$7, tri=$8, van=$9,
	budget_action_id=$10, payment_types_id=$11, plan_line_id=$12,
	step_id=$13, category_id=$14 WHERE id = $15`, op.Number, op.Name, op.Descript,
			op.Isr, op.Value, op.ValueDate, op.Length, op.TRI, op.VAN, op.BudgetActionID,
			op.PaymentTypeID, op.PlanLineID, op.StepID, op.CategoryID, op.ID)
	} else {
		res, err = db.Exec(`UPDATE physical_op SET descript=$1, isr=$2, value=$3, 
		valuedate=$4, length=$5, tri=$6, van=$7 WHERE id = $8`, op.Descript,
			op.Isr, op.Value, op.ValueDate, op.Length, op.TRI, op.VAN, op.ID)
	}
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Opération introuvable")
	}
	return err
}

// LinkFinancialCommitments updates the financial commitments linked to a physical operation in database.
func (op *PhysicalOp) LinkFinancialCommitments(fcIDs []int64, db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE financial_commitment SET physical_op_id = $1 
	WHERE id = ANY($2)`, op.ID, pq.Array(fcIDs))
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != int64(len(fcIDs)) {
		return errors.New("Opération ou engagements incorrects")
	}
	return nil
}

// Delete removes a physical operation from database.
func (op *PhysicalOp) Delete(db *sql.DB) (err error) {
	res, err := db.Exec("DELETE FROM physical_op WHERE id = $1", op.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Opération introuvable")
	}
	return nil
}

// Save insert or update into database the batch of physical operations sent.
func (op *PhysicalOpsBatch) Save(db *sql.DB) (err error) {
	if len(op.PhysicalOps) == 0 {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err = tx.Exec("DROP TABLE IF EXISTS temp_physical_op"); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec(`CREATE TABLE temp_physical_op (number varchar(10), 
		name varchar(255), descript text, isr boolean, value bigint, 
		valuedate date, length bigint, tri integer, van bigint, action varchar(11), 
		step varchar(50), category varchar(50), payment_types_id integer, 
		plan_line_id integer)`); err != nil {
		tx.Rollback()
		return err
	}
	var value string
	var values []string
	for _, o := range op.PhysicalOps {
		if len(o.Number) != 7 {
			tx.Rollback()
			return errors.New("Number " + o.Number + " incorrect")
		}
		if len(o.Name) == 0 {
			tx.Rollback()
			return errors.New("Name vide")
		}
		value = "(" + toSQL(o.Number) + "," + toSQL(o.Name) + "," + toSQL(o.Descript) + "," +
			toSQL(o.Isr) + "," + toSQL(o.Value) + "," + toSQL(o.Valuedate) + "," +
			toSQL(o.Length) + "," + toSQL(o.Step) + "," + toSQL(o.Category) + "," +
			toSQL(o.TRI) + "," + toSQL(o.VAN) + "," + toSQL(o.Action) + "," +
			toSQL(o.PaymentTypeID) + "," + toSQL(o.PlanLineID) + ")"
		values = append(values, value)
	}
	if _, err = tx.Exec(`INSERT INTO temp_physical_op (number, name, descript, isr, value, 
		valuedate, length, step, category, tri, van, action, payment_types_id, plan_line_id)
		VALUES ` + strings.Join(values, ",")); err != nil {
		tx.Rollback()
		return err
	}
	queries := []string{
		`WITH new AS (
			SELECT p.id, t.number, t.name, t.descript, t.isr, t.value, t.valuedate, t.length, 
				t.tri, t.van, b.id AS budget_action_id, t.payment_types_id, t.plan_line_id, 
				s.id AS step_id, c.id AS category_id 
			FROM temp_physical_op t
			LEFT JOIN physical_op p ON t.number = p.number
			LEFT OUTER JOIN 
				(SELECT ba.id, bp.code_contract||bp.code_function||bp.code_number||ba.code AS code
				 FROM budget_action ba, budget_program bp 
				 WHERE ba.program_id = bp.id) b 
			ON b.code = t.action
			LEFT OUTER JOIN step s ON s.name = t.step
			LEFT OUTER JOIN category c ON c.name = t.category)
		UPDATE physical_op AS op SET 
			name=new.name, descript=COALESCE(new.descript, op.descript), isr = COALESCE(new.isr, op.isr),
			value=COALESCE(new.value, op.value), valuedate=COALESCE(new.valuedate, op.valuedate),
			length=COALESCE(new.length, op.length), tri=COALESCE(new.tri, op.tri), van=COALESCE(new.van, op.van), 
			budget_action_id=COALESCE(new.budget_action_id, op.budget_action_id),
			payment_types_id=COALESCE(new.payment_types_id, op.payment_types_id),
			plan_line_id=COALESCE(new.plan_line_id, op.plan_line_id),
			step_id = COALESCE(new.step_id, op.step_id),
			category_id = COALESCE(new.category_id, op.category_id)
		FROM new WHERE op.id = new.id`,
		`INSERT INTO physical_op (number, name, descript, isr, value, valuedate, length,
			tri, van, payment_types_id, budget_action_id, plan_line_id, step_id, category_id)
		SELECT t.number, t.name, t.descript, t.isr, t.value, t.valuedate, t.length, t.tri, 
			t.van, t.payment_types_id, b.id AS budget_action_id, t.plan_line_id, s.id, c.id
		FROM temp_physical_op t
		LEFT OUTER JOIN (SELECT ba.id, bp.code_contract||bp.code_function||bp.code_number||ba.code AS code
			FROM budget_action ba, budget_program bp
			WHERE ba.program_id = bp.id) b 
		ON b.code = t.action
		LEFT OUTER JOIN step s ON s.name = t.step
		LEFT OUTER JOIN category c ON c.name = t.category
		WHERE t.number NOT IN (SELECT DISTINCT number FROM physical_op)`,
		`DROP TABLE IF EXISTS temp_physical_op`}
	for _, qry := range queries {
		if _, err = tx.Exec(qry); err != nil {
			tx.Rollback()
			return err
		}
	}
	err = tx.Commit()
	return err
}

// GetOpAll fetches all commitments previsions of a physical operation.
func (o *OpCommitments) GetOpAll(opID int64, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT f.id, f.date, f.iris_code, f.name AS name, b.name AS beneficiary, f.value, 
	f.lapse_date, f.value - COALESCE(SUM(p.value - p.cancelled_value),0) AS available
	FROM financial_commitment f
	JOIN beneficiary b ON b.code = f.beneficiary_code
	LEFT JOIN payment p ON p.financial_commitment_id = f.id
	WHERE f.physical_op_id = $1 GROUP BY 1,2,3,5,6,7 ORDER BY 2`, opID)
	if err != nil {
		return err
	}
	var r OpCommitment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Date, &r.IrisCode, &r.Name, &r.Beneficiary,
			&r.Value, &r.LapseDate, &r.Available); err != nil {
			return err
		}
		o.OpCommitments = append(o.OpCommitments, r)
	}
	err = rows.Err()
	if len(o.OpCommitments) == 0 {
		o.OpCommitments = []OpCommitment{}
	}
	return err
}

// SetPrevisions update and create previsions attached to a physical operation into database.
func (op *PhysicalOp) SetPrevisions(o *OpPrevisions, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err = tx.Exec("DELETE FROM prev_commitment WHERE physical_op_id = $1", op.ID); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec("DELETE FROM prev_payment WHERE physical_op_id = $1", op.ID); err != nil {
		tx.Rollback()
		return err
	}
	var value string
	var values []string
	for _, pc := range o.Commitments {
		value = "(" + toSQL(pc.Year) + "," + toSQL(pc.Value) + "," + toSQL(pc.Descript) +
			"," + toSQL(pc.TotalValue) + "," + toSQL(pc.StateRatio) + "," + toSQL(op.ID) + ")"
		values = append(values, value)
	}
	if len(values) > 0 {
		if _, err = tx.Exec("INSERT INTO prev_commitment (year, value, descript, total_value, state_ratio, physical_op_id) VALUES" + strings.Join(values, ",")); err != nil {
			tx.Rollback()
			return err
		}
	}
	values = nil
	for _, p := range o.Payments {
		value = "(" + toSQL(p.Year) + "," + toSQL(p.Value) + "," + toSQL(p.Descript) +
			"," + toSQL(op.ID) + ")"
		values = append(values, value)
	}
	if len(values) > 0 {
		if _, err = tx.Exec("INSERT INTO prev_payment (year, value, descript, physical_op_id) VALUES" + strings.Join(values, ",")); err != nil {
			tx.Rollback()
			return err
		}
	}
	err = tx.Commit()
	return err
}

// GetPrevCommitments fetches all prevision commitments linked to a physical operation.
func (op *PhysicalOp) GetPrevCommitments(p *PrevCommitments, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id, physical_op_id, year, value, descript, state_ratio, total_value
	FROM prev_commitment WHERE physical_op_id = $1`, op.ID)
	if err != nil {
		return err
	}
	var r PrevCommitment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.PhysicalOpID, &r.Year, &r.Value, &r.Descript, &r.StateRatio,
			&r.TotalValue); err != nil {
			return err
		}
		p.PrevCommitments = append(p.PrevCommitments, r)
	}
	err = rows.Err()
	if len(p.PrevCommitments) == 0 {
		p.PrevCommitments = []PrevCommitment{}
	}
	return err
}

// GetPrevPayments fetches all prevision payments attached to a physical operation.
func (op *PhysicalOp) GetPrevPayments(p *PrevPayments, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id, physical_op_id, year, value, descript
	FROM prev_payment WHERE physical_op_id = $1`, op.ID)
	if err != nil {
		return err
	}
	var r PrevPayment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.PhysicalOpID, &r.Year, &r.Value, &r.Descript); err != nil {
			return err
		}
		p.PrevPayments = append(p.PrevPayments, r)
	}
	err = rows.Err()
	if len(p.PrevPayments) == 0 {
		p.PrevPayments = []PrevPayment{}
	}
	return err
}

// GetYearPrevCommitments fetches all prevision commitments linked to a physical operation.
func (op *PhysicalOp) GetYearPrevCommitments(p *PrevCommitments, year int64, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id, physical_op_id, year, value, descript, state_ratio, total_value
	FROM prev_commitment WHERE physical_op_id = $1 AND year >= $2`, op.ID, year)
	if err != nil {
		return err
	}
	var r PrevCommitment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.PhysicalOpID, &r.Year, &r.Value, &r.Descript, &r.StateRatio,
			&r.TotalValue); err != nil {
			return err
		}
		p.PrevCommitments = append(p.PrevCommitments, r)
	}
	err = rows.Err()
	if len(p.PrevCommitments) == 0 {
		p.PrevCommitments = []PrevCommitment{}
	}
	return err
}

// GetYearPrevPayments fetches all prevision payments attached to a physical operation.
func (op *PhysicalOp) GetYearPrevPayments(p *PrevPayments, year int64, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id, physical_op_id, year, value, descript
	FROM prev_payment WHERE physical_op_id = $1 AND year >= $2`, op.ID, year)
	if err != nil {
		return err
	}
	var r PrevPayment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.PhysicalOpID, &r.Year, &r.Value, &r.Descript); err != nil {
			return err
		}
		p.PrevPayments = append(p.PrevPayments, r)
	}
	err = rows.Err()
	if len(p.PrevPayments) == 0 {
		p.PrevPayments = []PrevPayment{}
	}
	return err
}

// GetOpPendings calculates pending sum attached to a physical operation.
func (op *PhysicalOp) GetOpPendings(year int64, db *sql.DB) (o OpPendings, err error) {
	var value NullInt64
	err = db.QueryRow(`SELECT SUM(proposed_value) FROM pending_commitments 
	WHERE physical_op_id = $1 AND EXTRACT(YEAR from commission_date)=$2`, op.ID, year).Scan(&value)
	if err == nil {
		o.OpPendings = []OpPending{{Value: value}}
	}
	return o, err
}

// GetAll fetches all physical operations from database.
func (ops *PhysicalOps) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id, number, name, descript, isr, value, valuedate,
	length, tri, van, budget_action_id, payment_types_id, plan_line_id, step_id,
	category_id FROM physical_op`)
	if err != nil {
		return err
	}
	var r PhysicalOp
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Number, &r.Name, &r.Descript, &r.Isr, &r.Value,
			&r.ValueDate, &r.Length, &r.TRI, &r.VAN, &r.BudgetActionID, &r.PaymentTypeID,
			&r.PlanLineID, &r.StepID, &r.CategoryID); err != nil {
			return err
		}
		ops.PhysicalOps = append(ops.PhysicalOps, r)
	}
	err = rows.Err()
	if len(ops.PhysicalOps) == 0 {
		ops.PhysicalOps = []PhysicalOp{}
	}
	return err
}
