package models

import (
	"database/sql"
	"strconv"
)

// Programming model
type Programming struct {
	ID           int64       `json:"id" gorm:"column:id"`
	Value        int64       `json:"value" gorm:"column:value"`
	PhysicalOpID int64       `json:"physical_op_id" gorm:"column:physical_op_id"`
	CommissionID int64       `json:"commission_id" gorm:"column:commission_id"`
	Year         NullInt64   `json:"year" gorm:"column:year"`
	StateRatio   NullFloat64 `json:"state_ratio" gorm:"column:state_ratio"`
	TotalValue   NullInt64   `json:"total_value" gorm:"column:total_value"`
}

// ProgrammingFullDatas embeddes physical operations and linked programmings for json expert.
type ProgrammingFullDatas struct {
	ID                  NullInt64   `json:"id"`
	Value               NullInt64   `json:"value"`
	TotalValue          NullInt64   `json:"total_value"`
	StateRatio          NullFloat64 `json:"state_ratio"`
	PhysicalOpID        int64       `json:"physical_op_id"`
	CommissionID        NullInt64   `json:"commission_id"`
	OpNumber            string      `json:"op_number"`
	OpName              string      `json:"op_name"`
	Prevision           NullInt64   `json:"prevision"`
	TotalPrevision      NullInt64   `json:"total_prevision"`
	StateRatioPrevision NullFloat64 `json:"state_ratio_prevision"`
	PreProgValue        NullInt64   `json:"pre_prog_value"`
	PreProgTotalValue   NullInt64   `json:"pre_prog_total_value"`
	PreProgStateRatio   NullFloat64 `json:"pre_prog_state_ratio"`
	PreProgDescript     NullString  `json:"pre_prog_descript"`
	PlanName            NullString  `json:"plan_name"`
	PlanLineName        NullString  `json:"plan_line_name"`
	PlanLineValue       NullInt64   `json:"plan_line_value"`
	PlanLineTotalValue  NullInt64   `json:"plan_line_total_value"`
}

// Programmings embeddes an array of ProgrammingFullDatas for json export.
type Programmings struct {
	Programmings []ProgrammingFullDatas `json:"Programmings"`
}

// ProgrammingLine is used to decode a line of batch programming sent.
type ProgrammingLine struct {
	Value        int64       `json:"value"`
	PhysicalOpID int64       `json:"physical_op_id"`
	CommissionID int64       `json:"commission_id"`
	Year         int64       `json:"year"`
	TotalValue   NullInt64   `json:"total_value"`
	StateRatio   NullFloat64 `json:"state_ratio"`
}

// ProgrammingBatch programming batch payload.
type ProgrammingBatch struct {
	Programmings []ProgrammingLine `json:"Programmings"`
	Year         int64             `json:"year"`
}

// GetAll fetches operations and attached programmations datas of the given year.
func (p *Programmings) GetAll(year int64, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT pr.id, pr.value, pr.total_value, pr.state_ratio, op.id AS physical_op_id, 
	pr.commission_id, op.number as op_number, op.name as op_name, pc.value as prevision, 
	pc.total_value as total_prevision, pc.state_ratio as state_ratio_prevision,
	pp.value AS pre_prog_value, pp.total_value AS pre_prog_total_value,
	pp.state_ratio AS pre_prog_state_ratio, pp.descript AS pre_prog_descript, pl.plan_name, 
	pl.plan_line_name, pl.plan_line_value, pl.plan_line_total_value
FROM physical_op op
LEFT OUTER JOIN (SELECT pl.id, pl.name AS plan_line_name, pl.value AS plan_line_value, 
		pl.total_value AS plan_line_total_value, p.name AS plan_name 
	FROM plan_line pl, plan p WHERE pl.plan_id = p.id) pl ON op.plan_line_id = pl.id
LEFT OUTER JOIN (SELECT * FROM programmings WHERE year=$1) pr ON pr.physical_op_id = op.id
LEFT OUTER JOIN (SELECT * FROM prev_commitment WHERE year=$1) pc ON pc.physical_op_id = op.id
LEFT OUTER JOIN (SELECT * FROM pre_programmings WHERE year=$1) pp ON op.id = pp.physical_op_id`, year)
	if err != nil {
		return err
	}
	defer rows.Close()
	var r ProgrammingFullDatas
	for rows.Next() {
		err = rows.Scan(&r.ID, &r.Value, &r.TotalValue, &r.StateRatio, &r.PhysicalOpID,
			&r.CommissionID, &r.OpNumber, &r.OpName, &r.Prevision, &r.TotalPrevision, &r.StateRatioPrevision,
			&r.PreProgValue, &r.PreProgTotalValue, &r.PreProgStateRatio, &r.PreProgDescript, &r.PlanName,
			&r.PlanLineName, &r.PlanLineValue, &r.PlanLineTotalValue)
		if err != nil {
			return err
		}
		p.Programmings = append(p.Programmings, r)
	}
	err = rows.Err()
	return err
}

// ProgrammingsYear embeddes one year programmings for json export.
type ProgrammingsYear struct {
	Year int `json:"year"`
}

// ProgrammingsYears embeddes all years of programmings table for json export.
type ProgrammingsYears struct {
	Years []ProgrammingsYear `json:"ProgrammingsYears"`
}

// GetAll fetches all years of programmings tables.
func (p *ProgrammingsYears) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query("SELECT DISTINCT year from programmings")
	if err != nil {
		return err
	}
	defer rows.Close()
	var row ProgrammingsYear
	for rows.Next() {
		err = rows.Scan(&row.Year)
		if err != nil {
			return err
		}
		p.Years = append(p.Years, row)
	}
	err = rows.Err()
	return err
}

// ProgrammingsPerMonth embeddes programmings value of each month for json export.
type ProgrammingsPerMonth struct {
	Month int   `json:"month"`
	Value int64 `json:"value"`
}

// ProgrammingsPerMonthes embeddes programmings value of all monthes for json export.
type ProgrammingsPerMonthes struct {
	ProgrammingsPerMonth []ProgrammingsPerMonth `json:"ProgrammingsPerMonth"`
}

// GetAll fetches programmings values of a given year.
func (p *ProgrammingsPerMonthes) GetAll(year int, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT extract(month from c.date)::integer as month, sum(p.value)::bigint as value
	FROM commissions c, programmings p WHERE p.commission_id=c.id AND year = ` + strconv.Itoa(year) +
		` GROUP BY 1 ORDER BY 1`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var row ProgrammingsPerMonth
	for rows.Next() {
		err = rows.Scan(&row.Month, &row.Value)
		if err != nil {
			return err
		}
		p.ProgrammingsPerMonth = append(p.ProgrammingsPerMonth, row)
	}
	err = rows.Err()
	return err
}

// Save resets programmings into database according to batch sent.
func (p *ProgrammingBatch) Save(db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE from programmings WHERE year = $1", p.Year); err != nil {
		tx.Rollback()
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO programmings (value, physical_op_id, commission_id, year, 
		total_value, state_ratio) VALUES ($1,$2,$3,$4,$5,$6)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, p := range p.Programmings {
		if _, err := stmt.Exec(p.Value, p.PhysicalOpID, p.CommissionID, p.Year, p.TotalValue, p.StateRatio); err != nil {
			tx.Rollback()
			return err
		}
	}
	err = tx.Commit()
	return err
}
