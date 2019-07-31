package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"
)

// FinancialCommitment model
type FinancialCommitment struct {
	ID              int64     `json:"id"`
	PhysicalOpID    NullInt64 `json:"physical_op_id"`
	PlanLineID      NullInt64 `json:"plan_line_id"`
	Chapter         string    `json:"chapter"`
	Action          string    `json:"action"`
	IrisCode        string    `json:"iris_code"`
	CoriolisYear    string    `json:"coriolis_year"`
	CoriolisEgtCode string    `json:"coriolis_egt_code"`
	CoriolisEgtNum  string    `json:"coriolis_egt_num"`
	CoriolisEgtLine string    `json:"coriolis_egt_line"`
	Name            string    `json:"name"`
	BeneficiaryCode int       `json:"beneficiary_code"`
	Date            time.Time `json:"date"`
	Value           int64     `json:"value"`
	ActionID        NullInt64 `json:"action_id"`
	LapseDate       NullTime  `json:"lapse_date"`
}

// FinancialCommitments embeddes an array of FinancialCommitment for json export.
type FinancialCommitments struct {
	FinancialCommitments []FinancialCommitment `json:"FinancialCommitment"`
}

// UnlinkedFinancialCommitment embeddes a row for the query.
type UnlinkedFinancialCommitment struct {
	ID          int       `json:"id"`
	Value       int64     `json:"value"`
	IrisCode    string    `json:"iris_code"`
	Name        string    `json:"name"`
	Date        time.Time `json:"date"`
	Beneficiary string    `json:"beneficiary"`
}

// UnlinkedFinancialCommitments embeddes an array of UnlinkedFinancialCommitment for json export.
type UnlinkedFinancialCommitments struct {
	Commitments []UnlinkedFinancialCommitment `json:"FinancialCommitment"`
}

// FCSearchPattern embeddes parameters to query unlinked financial commitments.
type FCSearchPattern struct {
	LinkType   string
	SearchText string
	MinDate    time.Time
	Page       int64
}

// PaginatedUnlinkedItems embeddes all datas for unlinked financial commitments query.
type PaginatedUnlinkedItems struct {
	Data        []UnlinkedFinancialCommitment `json:"data"`
	CurrentPage int64                         `json:"current_page"`
	LastPage    int64                         `json:"last_page"`
}

// OpLinkedFinancialCommitment embeddes a row for the query.
type OpLinkedFinancialCommitment struct {
	FcID          int       `json:"fcId"`
	FcValue       int64     `json:"fcValue"`
	FcName        string    `json:"fcName"`
	IrisCode      string    `json:"iris_code"`
	FcDate        time.Time `json:"fcDate"`
	OpNumber      string    `json:"opNumber"`
	OpName        string    `json:"opName"`
	FcBeneficiary string    `json:"fcBeneficiary"`
}

// OpLinkedFinancialCommitments embeddes an array of OpLinkedFinancialCommitment for json export.
type OpLinkedFinancialCommitments struct {
	FinancialCommitments []OpLinkedFinancialCommitment `json:"FinancialCommitment"`
}

// PaginatedOpLinkedItems embeddes all datas for financial commitments linked to a physical operation query.
type PaginatedOpLinkedItems struct {
	Data        []OpLinkedFinancialCommitment `json:"data"`
	CurrentPage int64                         `json:"current_page"`
	LastPage    int64                         `json:"last_page"`
}

// PlanLineLinkedFinancialCommitment is used to query financial commitment linked to a plan line.
type PlanLineLinkedFinancialCommitment struct {
	FcID          int       `json:"fcId"`
	FcValue       int64     `json:"fcValue"`
	FcName        string    `json:"fcName"`
	IrisCode      string    `json:"iris_code"`
	FcDate        time.Time `json:"fcDate"`
	PlName        string    `json:"plName"`
	FcBeneficiary string    `json:"fcBeneficiary"`
}

// PlanLineLinkedFinancialCommitments embeddes an array of PlanLineLinkedFinancialCommitment.
type PlanLineLinkedFinancialCommitments struct {
	FinancialCommitments []PlanLineLinkedFinancialCommitment `json:"FinancialCommitment"`
}

// PaginatedPlanLineLinkedItems embeddes all datas for financial commitments linked to a plan line query.
type PaginatedPlanLineLinkedItems struct {
	Data        []PlanLineLinkedFinancialCommitment `json:"data"`
	CurrentPage int64                               `json:"current_page"`
	LastPage    int64                               `json:"last_page"`
}

// FinancialCommitmentLine embeddes a line of financial commitment batch request.
type FinancialCommitmentLine struct {
	Chapter         string    `json:"chapter"`
	Action          string    `json:"action"`
	IrisCode        string    `json:"iris_code"`
	CoriolisYear    string    `json:"coriolis_year"`
	CoriolisEgtCode string    `json:"coriolis_egt_code"`
	CoriolisEgtNum  string    `json:"coriolis_egt_num"`
	CoriolisEgtLine string    `json:"coriolis_egt_line"`
	Name            string    `json:"name"`
	Beneficiary     string    `json:"beneficiary"`
	BeneficiaryCode int       `json:"beneficiary_code"`
	Date            ExcelDate `json:"date"`
	Value           float64   `json:"value"`
	LapseDate       ExcelDate `json:"lapse_date"`
}

// FinancialCommitmentsBatch embeddes the data sent by a financial commitments batch request.
type FinancialCommitmentsBatch struct {
	FinancialCommitments []FinancialCommitmentLine `json:"FinancialCommitment"`
}

// Unlink set to null financial commitments links to a physical operation in database.
func (f *FinancialCommitment) Unlink(LinkType string, fcIDs []int64, db *sql.DB) (err error) {
	var IDQryPart string
	if LinkType == "PhysicalOp" {
		IDQryPart = "physical_op_id"
	} else {
		IDQryPart = "plan_line_id"
	}
	res, err := db.Exec(`UPDATE financial_commitment SET `+IDQryPart+` = NULL 
	WHERE id = ANY($1)`, pq.Array(fcIDs))
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != int64(len(fcIDs)) {
		return errors.New("Engagements incorrects")
	}
	return nil
}

// GetOpAll fetches all financial commitments linked to a physical operation from database.
func (f *FinancialCommitments) GetOpAll(opID int64, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id,	physical_op_id,	plan_line_id,	chapter, action,
	 iris_code, coriolis_year, coriolis_egt_code, coriolis_egt_num, coriolis_egt_line,
	 name, beneficiary_code, date, value, action_id, lapse_date FROM financial_commitment
	 WHERE physical_op_id=$1`, opID)
	if err != nil {
		return err
	}
	var r FinancialCommitment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.PhysicalOpID, &r.PlanLineID, &r.Chapter, &r.Action,
			&r.IrisCode, &r.CoriolisYear, &r.CoriolisEgtCode, &r.CoriolisEgtNum,
			&r.CoriolisEgtLine, &r.Name, &r.BeneficiaryCode, &r.Date, &r.Value, &r.ActionID,
			&r.LapseDate); err != nil {
			return err
		}
		f.FinancialCommitments = append(f.FinancialCommitments, r)
	}
	err = rows.Err()
	if len(f.FinancialCommitments) == 0 {
		f.FinancialCommitments = []FinancialCommitment{}
	}
	return err
}

// getPageOffset returns the correct offset and page according to total number of rows.
func getPageOffset(page int64, count int64) (offset int64, newPage int64, lastPage int64) {
	if count == 0 {
		return 0, 0, 1
	}
	offset = (page - 1) * 15
	newPage = 1
	if offset < 0 {
		offset = 0
	}
	if offset >= count {
		offset = (count - 1) - ((count - 1) % 15)
		newPage = offset/15 + 1
	}
	lastPage = (count-1)/15 + 1
	return offset, newPage, lastPage
}

// GetUnlinked fetches all financial commitments not linked to a physical operation or a plan line
// according to linkType parameter and matching search pattern.
func (p *PaginatedUnlinkedItems) GetUnlinked(pattern FCSearchPattern, db *sql.DB) (err error) {
	var idQryPart string
	var count int64
	if pattern.LinkType == "PhysicalOp" {
		idQryPart = "physical_op_id"
	} else {
		idQryPart = "plan_line_id"
	}
	if err = db.QueryRow(`SELECT count(f.id) count FROM financial_commitment f, beneficiary b 
	WHERE f.beneficiary_code = b.code AND f.date >= $1 AND `+idQryPart+` ISNULL AND
	(f.name ILIKE $2 OR b.name ILIKE $2 OR f.iris_code ILIKE $2)`,
		pattern.MinDate, pattern.SearchText).Scan(&count); err != nil {
		return err
	}
	offset, currentPage, lastPage := getPageOffset(pattern.Page, count)
	p.CurrentPage = currentPage
	p.LastPage = lastPage
	rows, err := db.Query(`SELECT f.id as id, f.value as value, f.iris_code as iris_code, 
	f.name as name, f.date as date, b.name as beneficiary 
	FROM financial_commitment f, beneficiary b
	WHERE f.beneficiary_code = b.code AND f.date >= $1 AND `+idQryPart+` ISNULL
	AND (f.name ILIKE $2 OR b.name ILIKE $2 OR f.iris_code ILIKE $2)
	ORDER BY 1 LIMIT 15 OFFSET $3`, pattern.MinDate, pattern.SearchText, offset)
	if err != nil {
		return err
	}
	var r UnlinkedFinancialCommitment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Value, &r.IrisCode, &r.Name, &r.Date,
			&r.Beneficiary); err != nil {
			return err
		}
		p.Data = append(p.Data, r)
	}
	err = rows.Err()
	if len(p.Data) == 0 {
		p.Data = []UnlinkedFinancialCommitment{}
	}
	return err
}

// GetLinked fetches all financial commitments linked to a physical operation and matching search pattern.
func (p *PaginatedOpLinkedItems) GetLinked(pattern FCSearchPattern, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	var count int64
	if err = tx.QueryRow(`SELECT count(f.id) 
	FROM financial_commitment f, beneficiary b, physical_op op
	WHERE f.physical_op_id = op.id AND f.beneficiary_code = b.code AND f.physical_op_id NOTNULL
	AND f.date > $1 AND (f.name ILIKE $2 OR b.name ILIKE $2 OR op.name ILIKE $2
	OR op.number ILIKE $2)`, pattern.MinDate, pattern.SearchText).Scan(&count); err != nil {
		tx.Rollback()
		return err
	}
	offset, currentPage, lastPage := getPageOffset(pattern.Page, count)
	p.CurrentPage = currentPage
	p.LastPage = lastPage
	rows, err := tx.Query(`SELECT f.id as fc_iD, f.value as fc_value, f.name as fc_name, 
	f.iris_code, f.date as fc_date, b.Name fc_beneficiary, op.number op_number, op.name op_name
	FROM financial_commitment f, beneficiary b, physical_op op
	WHERE f.physical_op_id = op.id AND f.beneficiary_code = b.code AND f.physical_op_id NOTNULL
	AND f.date > $1 AND (f.name ILIKE $2 OR b.name ILIKE $2 OR op.name ILIKE $2 
	OR op.number ILIKE $2)
	ORDER BY 1 LIMIT 15 OFFSET $3`, pattern.MinDate, pattern.SearchText, offset)
	if err != nil {
		tx.Rollback()
		return err
	}
	var r OpLinkedFinancialCommitment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.FcID, &r.FcValue, &r.FcName, &r.IrisCode, &r.FcDate,
			&r.FcBeneficiary, &r.OpNumber, &r.OpName); err != nil {
			return err
		}
		p.Data = append(p.Data, r)
	}
	err = rows.Err()
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if len(p.Data) == 0 {
		p.Data = []OpLinkedFinancialCommitment{}
	}
	return err
}

// GetLinked fetches all financial commitments linked to a physical operation and matching search pattern.
func (p *PaginatedPlanLineLinkedItems) GetLinked(pattern FCSearchPattern, db *sql.DB) (err error) {
	tx, err := db.Begin()
	var count int64
	if err != nil {
		return err
	}
	if err = tx.QueryRow(`SELECT count(f.id) 
	FROM financial_commitment f, beneficiary b, plan_line pl
	WHERE f.plan_line_id = pl.id AND f.beneficiary_code = b.code AND f.plan_line_id NOTNULL
	AND f.date > $1 AND (f.name ILIKE $2 OR b.name ILIKE $2 OR pl.name ILIKE $2)`,
		pattern.MinDate, pattern.SearchText).Scan(&count); err != nil {
		tx.Rollback()
		return err
	}
	offset, currentPage, lastPage := getPageOffset(pattern.Page, count)
	p.CurrentPage = currentPage
	p.LastPage = lastPage
	rows, err := tx.Query(`SELECT f.id as fc_id, f.value as fc_value, f.name as fc_name, 
	f.iris_code, f.date as fc_date, b.Name fc_beneficiary, pl.name pl_name
	FROM financial_commitment f, beneficiary b, plan_line pl
	WHERE f.plan_line_id = pl.id AND f.beneficiary_code = b.code AND f.plan_line_id NOTNULL
	AND f.date > $1 AND (f.name ILIKE $2 OR b.name ILIKE $2 OR pl.name ILIKE $2)
	ORDER BY 1 LIMIT 15 OFFSET $3`, pattern.MinDate, pattern.SearchText, offset)
	if err != nil {
		tx.Rollback()
		return err
	}
	var r PlanLineLinkedFinancialCommitment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.FcID, &r.FcValue, &r.FcName, &r.IrisCode, &r.FcDate,
			&r.PlName, &r.FcBeneficiary); err != nil {
			return err
		}
		p.Data = append(p.Data, r)
	}
	err = rows.Err()
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if len(p.Data) == 0 {
		p.Data = []PlanLineLinkedFinancialCommitment{}
	}
	return err
}

// Save a batch of financial commitments into database.
func (f *FinancialCommitmentsBatch) Save(db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE from temp_commitment`); err != nil {
		tx.Rollback()
		return err
	}
	var values []string
	var value string
	for _, fc := range f.FinancialCommitments {
		value = "(" + toSQL(fc.Chapter) + "," + toSQL(fc.Action) + "," +
			toSQL(fc.IrisCode) + "," + toSQL(fc.CoriolisYear) + "," +
			toSQL(fc.CoriolisEgtCode) + "," + toSQL(fc.CoriolisEgtNum) + "," +
			toSQL(fc.CoriolisEgtLine) + "," + toSQL(fc.Name) + "," +
			toSQL(fc.Beneficiary) + "," + toSQL(fc.BeneficiaryCode) + "," +
			toSQL(fc.Date) + "," + toSQL(int64(100*fc.Value)) + "," + toSQL(fc.LapseDate) + ")"
		values = append(values, value)
	}
	if _, err = tx.Exec(`INSERT INTO temp_commitment (chapter,action,iris_code,
		coriolis_year,coriolis_egt_code,coriolis_egt_num,coriolis_egt_line,name,
		beneficiary,beneficiary_code,date,value,lapse_date) VALUES ` + strings.Join(values, ",")); err != nil {
		tx.Rollback()
		return err
	}
	queries := []string{
		`WITH new AS (
			SELECT f.id, t.chapter, t.action, t.iris_code, t.name, t.beneficiary_code, t.date, t.value, t.lapse_date
			FROM temp_commitment t LEFT JOIN financial_commitment f ON t.iris_code = f.iris_code 
			 WHERE (f.value <> t.value OR f.chapter <> t.chapter OR f.action <> t.action OR f.name <> t.name OR 
							 f.coriolis_year <> t.coriolis_year OR  f.coriolis_egt_code <> t.coriolis_egt_code OR 
							 f.coriolis_egt_num <> t.coriolis_egt_num OR f.coriolis_egt_line <> t.coriolis_egt_line OR 
							 f.beneficiary_code <> t.beneficiary_code OR f.lapse_date IS DISTINCT FROM t.lapse_date) 
							 AND f.date = t.date) 
		UPDATE financial_commitment SET 
		chapter = new.chapter,  action = new.action, name = new.name, beneficiary_code = new.beneficiary_code, 
		 value = new.value, lapse_date = new.lapse_date 
		FROM new WHERE financial_commitment.id = new.id`,
		`INSERT INTO financial_commitment (physical_op_id, chapter, action, iris_code,
			coriolis_year, coriolis_egt_code, coriolis_egt_num, coriolis_egt_line, name, beneficiary_code, date,
			value, lapse_date) 
		SELECT NULL as physical_op_id, chapter, action, iris_code, coriolis_year, coriolis_egt_code, coriolis_egt_num, 
			coriolis_egt_line, name, beneficiary_code, date, value, lapse_date
			FROM temp_commitment t 
		WHERE (t.iris_code, t.date) NOT IN (SELECT iris_code, date FROM financial_commitment)`,
		`WITH new AS (
			SELECT t.beneficiary_code, t.beneficiary, t.date FROM temp_commitment t
			WHERE t.beneficiary_code NOT IN (SELECT code FROM beneficiary) )
		INSERT INTO beneficiary (code, name) SELECT beneficiary_code, beneficiary FROM new
			WHERE (date, beneficiary_code) IN (SELECT Max(date), beneficiary_code FROM temp_commitment GROUP BY 2)`,
		` WITH duplicated AS (SELECT id from financial_commitment WHERE iris_code IN
	(SELECT iris_code FROM financial_commitment WHERE iris_code in
		(SELECT iris_code FROM
			(SELECT SUM(1) as count, iris_code FROM financial_commitment GROUP BY 2) fcCount WHERE fcCount.count > 1)
					AND coriolis_egt_line <> '1') AND coriolis_egt_line = '1')
UPDATE financial_commitment SET value = 0 FROM duplicated WHERE financial_commitment.id=duplicated.id`,
		`WITH correspond AS (SELECT fc_extract.fc_id, ba_full.ba_id FROM 
	(SELECT fc.id AS fc_id, substring (fc.action FROM '^[0-9sS]+') AS fc_action FROM financial_commitment fc) fc_extract,
(SELECT ba.id AS ba_id, bp.code_contract || bp.code_function || bp.code_number || ba.code AS ba_code 
FROM budget_action ba, budget_program bp WHERE ba.program_id = bp.id) ba_full
WHERE fc_extract.fc_action = ba_full.ba_code)
UPDATE financial_commitment SET action_id = correspond.ba_id
FROM correspond WHERE financial_commitment.id = correspond.fc_id`}
	for _, qry := range queries {
		if _, err := tx.Exec(qry); err != nil {
			tx.Rollback()
			return err
		}
	}
	if _, err := tx.Exec("DELETE from temp_commitment"); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec(`INSERT INTO import_logs (category,last_date) 
		VALUES ('FinancialCommitments',$1)
		ON CONFLICT (category) DO UPDATE SET last_date = EXCLUDED.last_date;`,
		time.Now()); err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}
