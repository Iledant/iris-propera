package models

import (
	"database/sql"
	"errors"
	"fmt"
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
	APP             bool      `json:"app"`
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

// Pagination is the embeddes the common fields for paginated commitments.
type Pagination struct {
	CurrentPage int64 `json:"current_page"`
	ItemsCount  int64 `json:"items_count"`
	Offset      int64 `json:"-"`
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
	UnlinkedFinancialCommitments
	Pagination
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
	Commitments []OpLinkedFinancialCommitment `json:"FinancialCommitment"`
}

// PaginatedOpLinkedItems embeddes all datas for financial commitments linked to
//  a physical operation query.
type PaginatedOpLinkedItems struct {
	OpLinkedFinancialCommitments
	Pagination
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
	Commitments []PlanLineLinkedFinancialCommitment `json:"FinancialCommitment"`
}

// PaginatedPlanLineLinkedItems embeddes all datas for financial commitments linked to a plan line query.
type PaginatedPlanLineLinkedItems struct {
	PlanLineLinkedFinancialCommitments
	Pagination
}

// FinancialCommitmentLine embeddes a line of financial commitment batch request.
type FinancialCommitmentLine struct {
	Chapter         string     `json:"chapter"`
	Action          string     `json:"action"`
	IrisCode        string     `json:"iris_code"`
	CoriolisYear    string     `json:"coriolis_year"`
	CoriolisEgtCode string     `json:"coriolis_egt_code"`
	CoriolisEgtNum  string     `json:"coriolis_egt_num"`
	CoriolisEgtLine string     `json:"coriolis_egt_line"`
	Name            string     `json:"name"`
	Beneficiary     string     `json:"beneficiary"`
	BeneficiaryCode int        `json:"beneficiary_code"`
	Date            ExcelDate  `json:"date"`
	Value           float64    `json:"value"`
	LapseDate       ExcelDate  `json:"lapse_date"`
	APP             bool       `json:"app"`
	OpName          NullString `json:"op_name"`
}

// FinancialCommitmentsBatch embeddes the data sent by a financial commitments batch request.
type FinancialCommitmentsBatch struct {
	FinancialCommitments []FinancialCommitmentLine `json:"FinancialCommitment"`
}

// CmtOpProposal is used to propose a link between a newly imported commitment
// and a physical operation using the name field
type CmtOpProposal struct {
	CommitmentID   int64  `json:"commitment_id"`
	CommitmentName string `json:"commitment_name"`
	IRISOpName     string `json:"iris_op_name"`
	OpID           int64  `json:"op_id"`
	OpNumber       string `json:"op_number"`
	OpName         string `json:"op_name"`
}

// CmtOpProposals embeddes an array of CmtOpProposal
type CmtOpProposals struct {
	Lines []CmtOpProposal `json:"CmtOpProposal"`
}

// CmtOpLink is used to link financial commitments and physical operations
// after a financial commitments batch import and a proposal validation
type CmtOpLink struct {
	CommitmentID int64 `json:"commitment_id"`
	OpID         int64 `json:"op_id"`
}

// CmtOpLinks embeddes an array of CmtOpLink for json upload
type CmtOpLinks struct {
	Lines []CmtOpLink `json:"CmtOpLink"`
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
	 name, beneficiary_code, date, value, action_id, lapse_date, app
	 FROM financial_commitment WHERE physical_op_id=$1`, opID)
	if err != nil {
		return err
	}
	var r FinancialCommitment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.PhysicalOpID, &r.PlanLineID, &r.Chapter, &r.Action,
			&r.IrisCode, &r.CoriolisYear, &r.CoriolisEgtCode, &r.CoriolisEgtNum,
			&r.CoriolisEgtLine, &r.Name, &r.BeneficiaryCode, &r.Date, &r.Value, &r.ActionID,
			&r.LapseDate, &r.APP); err != nil {
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
func (p *Pagination) getPageOffset() {
	if p.ItemsCount == 0 {
		p.Offset = 0
		p.CurrentPage = 1
		return
	}
	p.Offset = (p.CurrentPage - 1) * 10
	if p.Offset < 0 {
		p.Offset = 0
	}
	if p.Offset >= p.ItemsCount {
		p.Offset = (p.ItemsCount - 1) - ((p.ItemsCount - 1) % 10)
	}
	p.CurrentPage = p.Offset/10 + 1
}

// GetUnlinked fetches all financial commitments not linked to a physical operation or a plan line
// according to linkType parameter and matching search pattern.
func (p *PaginatedUnlinkedItems) GetUnlinked(pattern FCSearchPattern, db *sql.DB) (err error) {
	var idQryPart string
	if pattern.LinkType == "PhysicalOp" {
		idQryPart = "physical_op_id"
	} else {
		idQryPart = "plan_line_id"
	}
	if err = db.QueryRow(`SELECT count(f.id) count FROM financial_commitment f, beneficiary b 
	WHERE f.beneficiary_code = b.code AND f.date >= $1 AND `+idQryPart+` ISNULL AND
	(f.name ILIKE $2 OR b.name ILIKE $2 OR f.iris_code ILIKE $2)`,
		pattern.MinDate, pattern.SearchText).Scan(&p.ItemsCount); err != nil {
		return err
	}
	p.CurrentPage = pattern.Page
	p.getPageOffset()
	rows, err := db.Query(`SELECT DISTINCT f.id as id, f.value as value, f.iris_code as iris_code, 
	f.name as name, f.date as date, b.name as beneficiary 
	FROM financial_commitment f, beneficiary b
	WHERE f.beneficiary_code = b.code AND f.date >= $1 AND `+idQryPart+` ISNULL
	AND (f.name ILIKE $2 OR b.name ILIKE $2 OR f.iris_code ILIKE $2)
	ORDER BY 1 LIMIT 10 OFFSET $3`, pattern.MinDate, pattern.SearchText, p.Offset)
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
		p.Commitments = append(p.Commitments, r)
	}
	err = rows.Err()
	if len(p.Commitments) == 0 {
		p.Commitments = []UnlinkedFinancialCommitment{}
	}
	return err
}

// GetLinked fetches all financial commitments linked to a physical operation
// and that matches the search pattern.
func (p *PaginatedOpLinkedItems) GetLinked(pattern FCSearchPattern, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if err = tx.QueryRow(`SELECT count(f.id) 
	FROM financial_commitment f, beneficiary b, physical_op op
	WHERE f.physical_op_id = op.id AND f.beneficiary_code = b.code AND f.physical_op_id NOTNULL
	AND f.date > $1 AND (f.name ILIKE $2 OR b.name ILIKE $2 OR op.name ILIKE $2
	OR op.number ILIKE $2)`, pattern.MinDate, pattern.SearchText).Scan(&p.ItemsCount); err != nil {
		tx.Rollback()
		return err
	}
	p.CurrentPage = pattern.Page
	p.getPageOffset()
	rows, err := tx.Query(`SELECT DISTINCT f.id as fc_iD, f.value as fc_value, f.name as fc_name, 
	f.iris_code, f.date as fc_date, b.Name fc_beneficiary, op.number op_number, op.name op_name
	FROM financial_commitment f, beneficiary b, physical_op op
	WHERE f.physical_op_id = op.id AND f.beneficiary_code = b.code AND f.physical_op_id NOTNULL
	AND f.date > $1 AND (f.name ILIKE $2 OR b.name ILIKE $2 OR op.name ILIKE $2 
	OR op.number ILIKE $2)
	ORDER BY 1 LIMIT 10 OFFSET $3`, pattern.MinDate, pattern.SearchText, p.Offset)
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
		p.Commitments = append(p.Commitments, r)
	}
	err = rows.Err()
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if len(p.Commitments) == 0 {
		p.Commitments = []OpLinkedFinancialCommitment{}
	}
	return err
}

// GetLinked fetches all financial commitments linked to a physical operation and matching search pattern.
func (p *PaginatedPlanLineLinkedItems) GetLinked(pattern FCSearchPattern, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if err = tx.QueryRow(`SELECT count(f.id) 
	FROM financial_commitment f, beneficiary b, plan_line pl
	WHERE f.plan_line_id = pl.id AND f.beneficiary_code = b.code AND f.plan_line_id NOTNULL
	AND f.date > $1 AND (f.name ILIKE $2 OR b.name ILIKE $2 OR pl.name ILIKE $2)`,
		pattern.MinDate, pattern.SearchText).Scan(&p.ItemsCount); err != nil {
		tx.Rollback()
		return err
	}
	p.CurrentPage = pattern.Page
	p.getPageOffset()
	rows, err := tx.Query(`SELECT DISTINCT f.id as fc_id, f.value as fc_value, f.name as fc_name, 
	f.iris_code, f.date as fc_date, b.Name fc_beneficiary, pl.name pl_name
	FROM financial_commitment f, beneficiary b, plan_line pl
	WHERE f.plan_line_id = pl.id AND f.beneficiary_code = b.code AND f.plan_line_id NOTNULL
	AND f.date > $1 AND (f.name ILIKE $2 OR b.name ILIKE $2 OR pl.name ILIKE $2)
	ORDER BY 1 LIMIT 10 OFFSET $3`, pattern.MinDate, pattern.SearchText, p.Offset)
	if err != nil {
		tx.Rollback()
		return err
	}
	var r PlanLineLinkedFinancialCommitment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.FcID, &r.FcValue, &r.FcName, &r.IrisCode, &r.FcDate,
			&r.FcBeneficiary, &r.PlName); err != nil {
			return err
		}
		p.Commitments = append(p.Commitments, r)
	}
	err = rows.Err()
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if len(p.Commitments) == 0 {
		p.Commitments = []PlanLineLinkedFinancialCommitment{}
	}
	return err
}

// Save a batch of financial commitments into database.
func (f *FinancialCommitmentsBatch) Save(db *sql.DB) (*CmtOpProposals, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(`DELETE from temp_commitment`); err != nil {
		tx.Rollback()
		return nil, err
	}
	stmt, err := tx.Prepare(pq.CopyIn("temp_commitment", "chapter", "action",
		"iris_code", "coriolis_year", "coriolis_egt_code", "coriolis_egt_num",
		"coriolis_egt_line", "name", "beneficiary", "beneficiary_code", "date",
		"value", "lapse_date", "app", "op_name"))
	if err != nil {
		return nil, fmt.Errorf("prepare stmt %v", err)
	}
	defer stmt.Close()
	for _, r := range f.FinancialCommitments {
		if _, err = stmt.Exec(r.Chapter, r.Action, r.IrisCode, r.CoriolisYear,
			r.CoriolisEgtCode, r.CoriolisEgtNum, r.CoriolisEgtLine, r.Name, r.Beneficiary,
			r.BeneficiaryCode, r.Date.ToDate(), int64(100*r.Value), r.LapseDate.ToDate(),
			r.APP, r.OpName); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("insertion de %+v  %v", r, err)
		}
	}
	if _, err = stmt.Exec(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("statement exec flush %v", err)
	}
	queries := []string{
		// remove duplicated commitment due to IRIS query bug
		`WITH cnt as (SELECT count(1) cnt,value,beneficiary_code,name
				FROM temp_commitment GROUP by 2,3,4),
			dup as (SELECT value,beneficiary_code,name FROM cnt WHERE cnt.cnt > 1),
			max_ids as (SELECT id FROM temp_commitment 
				WHERE (value,beneficiary_code,name,coriolis_year||coriolis_egt_num) in
				(SELECT value,beneficiary_code,name,max(coriolis_year||coriolis_egt_num) 
					FROM temp_commitment
					WHERE (value,beneficiary_code,name) in (SELECT * FROM dup)
					GROUP by 1,2,3)),
			sing_ids as (SELECT id FROM temp_commitment WHERE (value,beneficiary_code,name) in
				(SELECT value,beneficiary_code,name FROM cnt WHERE cnt.cnt = 1))
			DELETE FROM temp_commitment WHERE id not in
				(SELECT * FROM max_ids union all SELECT * FROM sing_ids)`,
		`WITH new AS (
				SELECT f.id,t.chapter,t.action,t.iris_code,t.name,t.beneficiary_code,t.date,
					t.value,t.lapse_date,t.app
				FROM temp_commitment t JOIN financial_commitment f ON t.iris_code=f.iris_code
				 WHERE (f.value<>t.value OR f.chapter<>t.chapter OR f.action<>t.action OR
								f.name<>t.name OR f.coriolis_year<>t.coriolis_year OR
								f.coriolis_egt_code<>t.coriolis_egt_code OR
								f.coriolis_egt_num<>t.coriolis_egt_num OR
								f.coriolis_egt_line<>t.coriolis_egt_line OR
								f.beneficiary_code<>t.beneficiary_code OR
								f.lapse_date IS DISTINCT FROM t.lapse_date OR f.app<>t.app)
								 AND f.date = t.date)
			UPDATE financial_commitment SET
			chapter=new.chapter,action=new.action,name=new.name,value=new.value,
			beneficiary_code=new.beneficiary_code,lapse_date=new.lapse_date,app=new.app
			FROM new WHERE financial_commitment.id = new.id`,
		`INSERT INTO financial_commitment (physical_op_id,chapter,action,iris_code,
				coriolis_year,coriolis_egt_code,coriolis_egt_num,coriolis_egt_line,name,
				beneficiary_code,date,value,lapse_date,app)
			SELECT NULL as physical_op_id,chapter,action,iris_code,coriolis_year,
				coriolis_egt_code,coriolis_egt_num,coriolis_egt_line,name,
				beneficiary_code,date,value,lapse_date,app
				FROM temp_commitment t
			WHERE (t.iris_code,t.date) NOT IN (SELECT iris_code,date FROM financial_commitment)`,
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
			return nil, err
		}
	}
	if _, err := tx.Exec(`INSERT INTO import_logs (category,last_date)
			VALUES ('FinancialCommitments',$1)
			ON CONFLICT (category) DO UPDATE SET last_date = EXCLUDED.last_date;`,
		time.Now()); err != nil {
		tx.Rollback()
		return nil, err
	}
	rows, err := tx.Query(
		`SELECT c.id,c.name,t.op_name,op.id,op.number,op.name FROM temp_commitment t
		JOIN financial_commitment c ON t.iris_code=c.iris_code
			AND t.coriolis_year=c.coriolis_year
			AND t.coriolis_egt_code=c.coriolis_egt_code
			AND t.coriolis_egt_num=c.coriolis_egt_num
			AND t.coriolis_egt_line=c.coriolis_egt_line AND t.date=c.date
		JOIN physical_op op ON t.op_name=op.name
		WHERE c.physical_op_id IS NULL
		UNION ALL
		SELECT c.id,c.name,t.op_name,op.id,op.number,op.name FROM temp_commitment t
		JOIN financial_commitment c ON t.iris_code=c.iris_code
			AND t.coriolis_year=c.coriolis_year
			AND t.coriolis_egt_code=c.coriolis_egt_code
			AND t.coriolis_egt_num=c.coriolis_egt_num
			AND t.coriolis_egt_line=c.coriolis_egt_line AND t.date=c.date
		JOIN physical_op op ON t.op_name<>op.name AND levenshtein(t.op_name,op.name)<2
		WHERE c.physical_op_id IS NULL`)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("select %v", err)
	}
	var (
		line CmtOpProposal
		resp CmtOpProposals
	)
	for rows.Next() {
		if err = rows.Scan(&line.CommitmentID, &line.CommitmentName, &line.IRISOpName,
			&line.OpID, &line.OpNumber, &line.OpName); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("rows scan %v", err)
		}
		resp.Lines = append(resp.Lines, line)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err %v", err)
	}
	if len(resp.Lines) == 0 {
		resp.Lines = []CmtOpProposal{}
	}
	if _, err = tx.Query(`DELETE FROM temp_commitment`); err != nil {
		return nil, fmt.Errorf("delete %v", err)
	}
	err = tx.Commit()
	return &resp, err
}

// GetAll fetches all commitments without a link to a plan line
func (p *UnlinkedFinancialCommitments) GetAll(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	rows, err := tx.Query(`SELECT DISTINCT f.id,f.value,f.name,f.iris_code,
	f.date,b.Name FROM financial_commitment f
	JOIN beneficiary b ON f.beneficiary_code = b.code
	WHERE f.plan_line_id IS NULL ORDER BY 5,4`)
	if err != nil {
		tx.Rollback()
		return err
	}
	var r UnlinkedFinancialCommitment
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Value, &r.Name, &r.IrisCode, &r.Date,
			&r.Beneficiary); err != nil {
			return err
		}
		p.Commitments = append(p.Commitments, r)
	}
	err = rows.Err()
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if len(p.Commitments) == 0 {
		p.Commitments = []UnlinkedFinancialCommitment{}
	}
	return err
}

// Save update the financial commitments to link to the physical operations
func (c *CmtOpLinks) Save(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("tx begin %v", err)
	}
	stmt, err := tx.Prepare(`UPDATE financial_commitment SET physical_op_id=$1
	WHERE id=$2`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("statement prepare %v", err)
	}
	defer stmt.Close()
	for _, l := range c.Lines {
		if _, err = stmt.Exec(l.OpID, l.CommitmentID); err != nil {
			tx.Rollback()
			return fmt.Errorf("statement exec %v", err)
		}
	}
	tx.Commit()
	return nil
}
