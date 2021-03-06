package models

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

// PreProgramming model
type PreProgramming struct {
	ID           int         `json:"id"`
	Year         int         `json:"year"`
	PhysicalOpID int         `json:"physical_op_id"`
	CommissionID int         `json:"commission_id"`
	Value        int64       `json:"value"`
	StateRatio   NullFloat64 `json:"state_ratio"`
	TotalValue   NullInt64   `json:"total_value"`
	Descript     NullString  `json:"descript"`
}

// FullPreProgramming is used to scan the select pre programming query results
type FullPreProgramming struct {
	PhysicalOpID        int64       `json:"physical_op_id"`
	PhysicalOpNumber    string      `json:"physical_op_number"`
	PhysicalOpName      string      `json:"physical_op_name"`
	PrevValue           NullInt64   `json:"prev_value"`
	PrevStateRatio      NullFloat64 `json:"prev_state_ratio"`
	PrevTotalValue      NullInt64   `json:"prev_total_value"`
	PrevDescript        NullString  `json:"prev_descript"`
	PreProgID           NullInt64   `json:"pre_prog_id"`
	PreProgValue        NullInt64   `json:"pre_prog_value"`
	PreProgYear         NullInt64   `json:"pre_prog_year"`
	PreProgCommissionID NullInt64   `json:"pre_prog_commission_id"`
	PreProgStateRatio   NullFloat64 `json:"pre_prog_state_ratio"`
	PreProgTotalValue   NullInt64   `json:"pre_prog_total_value"`
	PreProgDescript     NullString  `json:"pre_prog_descript"`
	PlanName            NullString  `json:"plan_name"`
	PlanLineName        NullString  `json:"plan_line_name"`
	PlanLineValue       NullInt64   `json:"plan_line_value"`
	PlanLineTotalValue  NullInt64   `json:"plan_line_total_value"`
}

// FullPreProgrammings embeddes an array of FullPreProgramming
type FullPreProgrammings struct {
	FullPreProgrammings []FullPreProgramming `json:"PreProgrammings"`
}

// PreProgrammingLine is used to decode a line par pre programming sent.
type PreProgrammingLine struct {
	PhysicalOpID int64       `json:"physical_op_id"`
	ID           NullInt64   `json:"pre_prog_id"`
	Year         int64       `json:"pre_prog_year"`
	Value        int64       `json:"pre_prog_value"`
	CommissionID int64       `json:"pre_prog_commission_id"`
	TotalValue   NullInt64   `json:"pre_prog_total_value"`
	StateRatio   NullFloat64 `json:"pre_prog_state_ratio"`
}

// PreProgrammingBatch is used to decode sent payload.
type PreProgrammingBatch struct {
	PreProgrammings []PreProgrammingLine `json:"PreProgrammings"`
	Year            int64                `json:"year"`
}

// GetAll fetches pre programmings with all datas from database.
func (f *FullPreProgrammings) GetAll(uID int64, year int64, db *sql.DB) (err error) {
	fromQry := ` physical_op op `
	if uID != 0 {
		fromQry = ` (SELECT * FROM physical_op WHERE id IN (SELECT physical_op_id FROM rights WHERE users_id = $2)) op `
	}
	query := `SELECT op.id AS physical_op_id, op.number AS physical_op_number, 
	op.name AS physical_op_name, pc.value AS prev_value, pc.state_ratio AS prev_state_ratio, 
	pc.total_value AS prev_total_value, pc.descript AS prev_descript, pp.id AS pre_prog_id,
	pp.value AS pre_prog_value, pp.year AS pre_prog_year, pp.commission_id AS pre_prog_commission_id, 
	pp.state_ratio AS pre_prog_state_ratio, pp.total_value AS pre_prog_total_value, 
	pp.descript AS pre_prog_descript, pl.plan_name, pl.plan_line_name, pl.plan_line_value,
	 pl.plan_line_total_value 
	FROM` + fromQry +
		`LEFT OUTER JOIN (SELECT pl.id, pl.name AS plan_line_name, pl.value AS plan_line_value, 
		pl.total_value AS plan_line_total_value, p.name AS plan_name
		FROM plan_line pl, plan p WHERE pl.plan_id = p.id) pl ON op.plan_line_id = pl.id
LEFT OUTER JOIN (SELECT * FROM prev_commitment WHERE year = $1) pc ON op.id = pc.physical_op_id
LEFT OUTER JOIN (SELECT * FROM pre_programmings WHERE year = $1) pp ON op.id = pp.physical_op_id`
	var rows *sql.Rows
	if uID == 0 {
		rows, err = db.Query(query, year)
	} else {
		rows, err = db.Query(query, year, uID)
	}
	if err != nil {
		return err
	}
	var r FullPreProgramming
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.PhysicalOpID, &r.PhysicalOpNumber, &r.PhysicalOpName,
			&r.PrevValue, &r.PrevStateRatio, &r.PrevTotalValue, &r.PrevDescript,
			&r.PreProgID, &r.PreProgValue, &r.PreProgYear, &r.PreProgCommissionID,
			&r.PreProgStateRatio, &r.PreProgTotalValue, &r.PreProgDescript, &r.PlanName,
			&r.PlanLineName, &r.PlanLineValue, &r.PlanLineTotalValue); err != nil {
			return err
		}
		f.FullPreProgrammings = append(f.FullPreProgrammings, r)
	}
	err = rows.Err()
	if len(f.FullPreProgrammings) == 0 {
		f.FullPreProgrammings = []FullPreProgramming{}
	}
	return err
}

// Save insert the batch of pre programmings into the database.
func (p *PreProgrammingBatch) Save(uID int64, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`CREATE TABLE IF NOT EXISTS temp_pre_programmings 
	(id integer, year integer NOT NULL, physical_op_id integer NOT NULL, 
		commission_id integer NOT NULL, value bigint NOT NULL, total_value bigint,
		 state_ratio double precision, descript text)`); err != nil {
		tx.Rollback()
		return fmt.Errorf("create temp table %v", err)
	}
	if len(p.PreProgrammings) > 0 {

		stmt, err := tx.Prepare(pq.CopyIn("temp_pre_programmings", "id", "year",
			"physical_op_id", "commission_id", "value", "total_value", "state_ratio",
			"descript"))
		if err != nil {
			return fmt.Errorf("prepare stmt %v", err)
		}
		defer stmt.Close()
		for _, r := range p.PreProgrammings {
			if _, err = stmt.Exec(r.ID, r.Year, r.PhysicalOpID, r.CommissionID, r.Value,
				r.TotalValue, r.StateRatio, NullString{Valid: false}); err != nil {
				tx.Rollback()
				return fmt.Errorf("insertion de %+v  %v", r, err)
			}
		}
		if _, err = stmt.Exec(); err != nil {
			tx.Rollback()
			return fmt.Errorf("statement exec flush %v", err)
		}

		if _, err = tx.Exec(`UPDATE pre_programmings SET value = t.value, 
	physical_op_id = t.physical_op_id, commission_id = t.commission_id,
	year = t.year, total_value = t.total_value, state_ratio = t.state_ratio,
	descript = t.descript
  FROM temp_pre_programmings t WHERE pre_programmings.id = t.id`); err != nil {
			tx.Rollback()
			return fmt.Errorf("update %v", err)
		}
	}
	if uID == 0 {
		if _, err = tx.Exec(`DELETE FROM pre_programmings pp 
		WHERE pp.physical_op_id IN (SELECT id FROM physical_op op)
		 AND pp.id NOT IN (SELECT id FROM temp_pre_programmings) AND pp.year = $1`,
			p.Year); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete %v", err)
		}
	} else {
		if _, err = tx.Exec(`DELETE FROM pre_programmings pp 
		WHERE pp.physical_op_id IN (SELECT id FROM physical_op
			WHERE id IN (SELECT physical_op_id FROM rights WHERE users_id = $1))
				AND pp.id NOT IN (SELECT id FROM temp_pre_programmings) AND pp.year = $2`,
			uID, p.Year); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete %v", err)
		}
	}
	if _, err = tx.Exec(`INSERT INTO pre_programmings (value,physical_op_id,
		commission_id,year,total_value,state_ratio,descript)
	(SELECT value,physical_op_id,commission_id,year,total_value,state_ratio,descript 
		FROM temp_pre_programmings 
		WHERE id ISNULL OR id NOT IN (SELECT DISTINCT id FROM pre_programmings))`); err != nil {
		tx.Rollback()
		return fmt.Errorf("insert table %v", err)
	}
	if _, err = tx.Exec(`DROP TABLE IF EXISTS temp_pre_programmings`); err != nil {
		tx.Rollback()
		return fmt.Errorf("droptable %v", err)
	}
	tx.Commit()
	return err
}
