package models

import (
	"database/sql"
	"strconv"
	"strings"
	"time"
)

// OpDptRatio model
type OpDptRatio struct {
	ID           int64   `json:"id" gorm:"column:id"`
	PhysicalOpID int64   `json:"physical_op_id" gorm:"column:physical_op_id"`
	R75          float64 `json:"r75" gorm:"column:r75"`
	R77          float64 `json:"r77" gorm:"column:r77"`
	R78          float64 `json:"r78" gorm:"column:r78"`
	R91          float64 `json:"r91" gorm:"column:r91"`
	R92          float64 `json:"r92" gorm:"column:r92"`
	R93          float64 `json:"r93" gorm:"column:r93"`
	R94          float64 `json:"r94" gorm:"column:r94"`
	R95          float64 `json:"r95" gorm:"column:r95"`
}

// OpDptRatioLine embeddes a line of sent datas for a batch of datas.
type OpDptRatioLine struct {
	PhysicalOpID int64   `json:"physical_op_id"`
	R75          float64 `json:"r75"`
	R77          float64 `json:"r77"`
	R78          float64 `json:"r78"`
	R91          float64 `json:"r91"`
	R92          float64 `json:"r92"`
	R93          float64 `json:"r93"`
	R94          float64 `json:"r94"`
	R95          float64 `json:"r95"`
}

// OpDptRatioBatch embeddes a batch of opDptRatioLine to upload into database.
type OpDptRatioBatch struct {
	OpDptRatioLines []OpDptRatioLine `json:"OpDptRatios"`
}

// FCPerDpt is used to decode one row of financial commitment per department query.
type FCPerDpt struct {
	Total NullInt64 `json:"total"`
	FC75  NullInt64 `json:"fc75"`
	FC77  NullInt64 `json:"fc77"`
	FC78  NullInt64 `json:"fc78"`
	FC91  NullInt64 `json:"fc91"`
	FC92  NullInt64 `json:"fc92"`
	FC93  NullInt64 `json:"fc93"`
	FC94  NullInt64 `json:"fc94"`
	FC95  NullInt64 `json:"fc95"`
}

// FCPerDepartments embeddes an array ofFcPerDpt for json export.
type FCPerDepartments struct {
	FCPerDepartments []FCPerDpt `json:"FinancialCommitmentPerDpt"`
}

// DetailedFCPerDpt is used to decode one row of detailed financial commitment per department query
type DetailedFCPerDpt struct {
	FCPerDpt
	ID     int64  `json:"id" gorm:"id"`
	Number string `json:"number" gorm:"number"`
	Name   string `json:"name" gorm:"name"`
}

// DetailedFCPerDepartments embeddes an array ofFcPerDpt for json export.
type DetailedFCPerDepartments struct {
	DetailedFCPerDepartments []DetailedFCPerDpt `json:"DetailedFinancialCommitmentPerDpt"`
}

// OpWithDptRatio embeddes a row of query that fetches physical operations and ratios par department.
type OpWithDptRatio struct {
	ID     int64       `json:"id"`
	Number string      `json:"number"`
	Name   string      `json:"name"`
	R75    NullFloat64 `json:"r75"`
	R77    NullFloat64 `json:"r77"`
	R78    NullFloat64 `json:"r78"`
	R91    NullFloat64 `json:"r91"`
	R92    NullFloat64 `json:"r92"`
	R93    NullFloat64 `json:"r93"`
	R94    NullFloat64 `json:"r94"`
	R95    NullFloat64 `json:"r95"`
}

// OpWithDptRatios embeddes an array of OpWithDptRatio for json export.
type OpWithDptRatios struct {
	OpWithDptRatios []OpWithDptRatio `json:"OpsWithDptRatios"`
}

// DetailedPrgPerDpt  is used to decode one row of detailed programmings per department query
type DetailedPrgPerDpt struct {
	Date   time.Time `json:"date"`
	ID     int       `json:"id"`
	Number string    `json:"number"`
	Name   string    `json:"name"`
	Total  NullInt64 `json:"total"`
	PR75   NullInt64 `json:"pr75"`
	PR77   NullInt64 `json:"pr77"`
	PR78   NullInt64 `json:"pr78"`
	PR91   NullInt64 `json:"pr91"`
	PR92   NullInt64 `json:"pr92"`
	PR93   NullInt64 `json:"pr93"`
	PR94   NullInt64 `json:"pr94"`
	PR95   NullInt64 `json:"pr95"`
}

// DetailedPrgPerDepartments embeddes an array of DetailedPrgPerDpt for json export.
type DetailedPrgPerDepartments struct {
	DetailedPrgPerDepartments []DetailedPrgPerDpt `json:"DetailedProgrammingsPerDpt"`
}

// GetAll fetches OpWithDptRatio from database according to user role with convention
// that uID is null for ADMINS or OBSERVERS and other for USERS
func (o *OpWithDptRatios) GetAll(uID int64, db *sql.DB) (err error) {
	var whereClause string
	if uID != 0 {
		whereClause = "WHERE op.id IN (SELECT physical_op_id FROM rights WHERE users_id = " +
			strconv.FormatInt(uID, 10) + " ) "
	}
	rows, err := db.Query(`SELECT op.id, op.number, op.name, r.r75, r.r77, r.r78, 
	r.r91, r.r92, r.r93, r.r94, r.r95
	FROM physical_op op
	LEFT OUTER JOIN op_dpt_ratios r ON r.physical_op_id = op.id` + whereClause)
	if err != nil {
		return err
	}
	var r OpWithDptRatio
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Number, &r.Name, &r.R75, &r.R77, &r.R78, &r.R91, &r.R92,
			&r.R93, &r.R94, &r.R95); err != nil {
			return err
		}
		o.OpWithDptRatios = append(o.OpWithDptRatios, r)
	}
	err = rows.Err()
	return err
}

// Save a batch of OpDptRatio into database.
func (o *OpDptRatioBatch) Save(uID int64, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	var andClause, andInsertClause string
	if uID != 0 {
		andClause = "AND op_dpt_ratios.physical_op_id IN (SELECT physical_op_id FROM rights WHERE users_id = " + strconv.FormatInt(uID, 10) + ")"
		andInsertClause = "AND physical_op_id IN (SELECT physical_op_id FROM rights WHERE users_id = " + strconv.FormatInt(uID, 10) + ")"
	}
	if _, err = tx.Exec("DROP TABLE IF EXISTS temp_op_dpt_ratios"); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec(`CREATE TABLE temp_op_dpt_ratios ( 
			physical_op_id integer, r75 double precision, r77 double precision, 
			r78 double precision, r91 double precision, r92 double precision, 
			r93 double precision, r94 double precision, r95 double precision );`); err != nil {
		tx.Rollback()
		return err
	}
	var values []string
	for _, o := range o.OpDptRatioLines {
		values = append(values, "("+toSQL(o.PhysicalOpID)+","+toSQL(o.R75)+","+
			toSQL(o.R77)+","+toSQL(o.R78)+","+toSQL(o.R91)+","+toSQL(o.R92)+","+
			toSQL(o.R93)+","+toSQL(o.R94)+","+toSQL(o.R95)+")")
	}
	if _, err = tx.Exec(`INSERT INTO temp_op_dpt_ratios VALUES` + strings.Join(values, ",")); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec(`DELETE FROM op_dpt_ratios WHERE physical_op_id NOT IN
		(SELECT physical_op_id FROM temp_op_dpt_ratios)` + andClause); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec(`UPDATE op_dpt_ratios SET r75 = t.r75, r77 = t.r77, r78 = t.r78,
		r91 = t.r91, r92 = t.r92, r93 = t.r93, r94 = t.r94, r95 = t.r95
		FROM temp_op_dpt_ratios t WHERE op_dpt_ratios.physical_op_id = t.physical_op_id` +
		andClause); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec(`INSERT INTO op_dpt_ratios (physical_op_id,r75,r77,r78,r91,r92,r93,r94,r95) 
		SELECT * FROM temp_op_dpt_ratios t
			WHERE t.physical_op_id NOT IN (SELECT physical_op_id FROM op_dpt_ratios)` +
		andInsertClause); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec("DROP TABLE IF EXISTS temp_op_dpt_ratios"); err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}

// GetAll fetches financial commitments per departments from database.
func (f *FCPerDepartments) GetAll(firstYear int, lastYear int, db *sql.DB) (err error) {
	sy0 := strconv.Itoa(firstYear)
	sy1 := strconv.Itoa(lastYear)
	query := `SELECT SUM(fc.value)::bigint AS total, SUM(fc.value*r.r75)::bigint AS fc75,
	SUM(fc.value*r.r77)::bigint AS fc77, SUM(fc.value*r.r78)::bigint AS fc78,
	SUM(fc.value*r.r91)::bigint AS fc91, SUM(fc.value*r.r92)::bigint AS fc92,
	SUM(fc.value*r.r93)::bigint AS fc93, SUM(fc.value*r.r94)::bigint AS fc94,
	SUM(fc.value*r.r95)::bigint AS fc95
	FROM financial_commitment fc
	LEFT OUTER JOIN physical_op op ON fc.physical_op_id = op.id
	LEFT OUTER JOIN op_dpt_ratios r ON r.physical_op_id = op.id
	WHERE extract(year FROM fc.date) >= ` + sy0 + ` AND extract(year FROM fc.date) <= ` + sy1
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	var r FCPerDpt
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Total, &r.FC75, &r.FC77, &r.FC78, &r.FC91, &r.FC92, &r.FC93,
			&r.FC94, &r.FC95); err != nil {
			return err
		}
		f.FCPerDepartments = append(f.FCPerDepartments, r)
	}
	err = rows.Err()
	return err
}

// GetAll fetches detailed financial commitments per departments from database.
func (f *DetailedFCPerDepartments) GetAll(firstYear int, lastYear int, db *sql.DB) (err error) {
	sy0 := strconv.Itoa(firstYear)
	sy1 := strconv.Itoa(lastYear)
	query := `SELECT op.id, op.number, op.name, SUM(fc.value)::bigint AS total,
	SUM(fc.value*r.r75)::bigint AS fc75, SUM(fc.value*r.r77)::bigint AS fc77, 
	SUM(fc.value*r.r78)::bigint AS fc78, SUM(fc.value*r.r91)::bigint AS fc91,
	SUM(fc.value*r.r92)::bigint AS fc92, SUM(fc.value*r.r93)::bigint AS fc93,
	SUM(fc.value*r.r94)::bigint AS fc94, SUM(fc.value*r.r95)::bigint AS fc95
	FROM financial_commitment fc
	LEFT OUTER JOIN physical_op op ON fc.physical_op_id = op.id
	LEFT OUTER JOIN op_dpt_ratios r ON r.physical_op_id = op.id
	WHERE extract(year FROM fc.date) >= ` + sy0 + ` AND extract(year FROM fc.date) <= ` +
		sy1 + ` GROUP BY 1,2,3 ORDER BY 2,3`
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	var r DetailedFCPerDpt
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Number, &r.Name, &r.Total, &r.FC75, &r.FC77, &r.FC78, &r.FC91,
			&r.FC92, &r.FC93, &r.FC94, &r.FC95); err != nil {
			return err
		}
		f.DetailedFCPerDepartments = append(f.DetailedFCPerDepartments, r)
	}
	err = rows.Err()
	return err
}

// GetAll fetches detailed programmings per departments from database.
func (f *DetailedPrgPerDepartments) GetAll(year int, db *sql.DB) (err error) {
	sy := strconv.Itoa(year)
	query := `SELECT c.date, op.id, op.number, op.name, SUM(pr.value)::bigint AS total,
	SUM(pr.value*r.r75)::bigint AS pr75, SUM(pr.value*r.r77)::bigint AS pr77, 
	SUM(pr.value*r.r78)::bigint AS pr78, SUM(pr.value*r.r91)::bigint AS pr91,
	SUM(pr.value*r.r92)::bigint AS pr92, SUM(pr.value*r.r93)::bigint AS pr93,
	SUM(pr.value*r.r94)::bigint AS pr94, SUM(pr.value*r.r95)::bigint AS pr95
	FROM programmings pr
	LEFT JOIN commissions c ON pr.commission_id = c.id
	LEFT OUTER JOIN physical_op op ON pr.physical_op_id = op.id
	LEFT OUTER JOIN op_dpt_ratios r ON r.physical_op_id = op.id
	WHERE pr.year = ` + sy + ` GROUP BY 1,2,3,4 ORDER BY 1,3,4`
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	var r DetailedPrgPerDpt
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Date, &r.ID, &r.Number, &r.Name, &r.Total, &r.PR75, &r.PR77,
			&r.PR78, &r.PR91, &r.PR92, &r.PR93, &r.PR94, &r.PR95); err != nil {
			return err
		}
		f.DetailedPrgPerDepartments = append(f.DetailedPrgPerDepartments, r)
	}
	err = rows.Err()
	return err
}
