package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

// Scenario model
type Scenario struct {
	ID       int64      `json:"id"`
	Name     string     `json:"name"`
	Descript NullString `json:"descript"`
}

// Scenarios embeddes an array of Scenario for json export.
type Scenarios struct {
	Scenarios []Scenario `json:"Scenario"`
}

// ScenarioDatas embeddes results of queries dedicated to operations previsions and
// scenario previsions according to it's operation list and offsets
type ScenarioDatas struct {
	OperationCrossTable json.RawMessage `json:"OperationCrossTable"`
	ScenarioCrossTable  json.RawMessage `json:"ScenarioCrossTable"`
}

// MABScenarioLine is used to decode one line of the multi annual budget
// scenario query that calculates commitments per budget entities.
type MABScenarioLine struct {
	Number      string     `json:"number"`
	Name        string     `json:"name"`
	Chapter     NullInt64  `json:"chapter"`
	Sector      NullString `json:"sector"`
	Subfunction NullString `json:"subfunction"`
	Program     NullString `json:"program"`
	Action      NullString `json:"action"`
	Y0          NullInt64  `json:"y0"`
	Y1          NullInt64  `json:"y1"`
	Y2          NullInt64  `json:"y2"`
	Y3          NullInt64  `json:"y3"`
	Y4          NullInt64  `json:"y4"`
}

// MultiAnnualBudgetScenario embeddes an array of MABScenarioLine to fetch
// the dedicated query.
type MultiAnnualBudgetScenario struct {
	MultiAnnualBudgetScenario []MABScenarioLine `json:"MultiannualBudgetScenario"`
}

// GetAll fetches all scenarios from database.
func (s *Scenarios) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id, name, descript FROM scenario`)
	if err != nil {
		return err
	}
	var r Scenario
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Name, &r.Descript); err != nil {
			return err
		}
		s.Scenarios = append(s.Scenarios, r)
	}
	err = rows.Err()
	if len(s.Scenarios) == 0 {
		s.Scenarios = []Scenario{}
	}
	return err
}

// Invalid checks of Scenario's field can be saved to database.
func (s *Scenario) Invalid() bool {
	return s.Name == "" || len(s.Name) > 255
}

// Create insert a new scenario into database.
func (s *Scenario) Create(db *sql.DB) (err error) {
	err = db.QueryRow("INSERT INTO scenario (name,descript) VALUES($1,$2) RETURNING id",
		s.Name, s.Descript).Scan(&s.ID)
	return err
}

// Update modifies a scenario into database.
func (s *Scenario) Update(db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE scenario SET name=$1, descript=$2 WHERE id = $3`,
		s.Name, s.Descript, s.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Scenario introuvable")
	}
	return err
}

// Delete remote scenario from database.
func (s *Scenario) Delete(db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err = tx.Exec("DELETE from scenario_offset WHERE scenario_id = $1",
		s.ID); err != nil {
		tx.Rollback()
		return
	}
	res, err := tx.Exec("DELETE FROM scenario WHERE id = $1", s.ID)
	if err != nil {
		tx.Rollback()
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if count != 1 {
		tx.Rollback()
		return errors.New("Scenario introuvable")
	}
	return nil
}

// Populate calculates datas linked to a scenario.
func (d *ScenarioDatas) Populate(sID int64, firstYear int64, db *sql.DB) (err error) {
	var lastYear NullInt64
	if err = db.QueryRow(`SELECT max(p.year+s.offset)::bigint AS year 
	 FROM prev_commitment p, scenario_offset s 
	 WHERE s.physical_op_id = p.physical_op_id AND s.scenario_id= $1`, sID).
		Scan(&lastYear); err != nil {
		return err
	}
	if !lastYear.Valid {
		lastYear.Int64 = firstYear + 4
	}
	var columnNames, typesNames, jsonNames []string

	for i := firstYear; i <= lastYear.Int64; i++ {
		sy := strconv.FormatInt(i-firstYear, 10)
		columnNames = append(columnNames, `"`+strconv.FormatInt(i, 10)+`" AS y`+sy)
		typesNames = append(typesNames, `"`+strconv.FormatInt(i, 10)+`" NUMERIC`)
		jsonNames = append(jsonNames, `'y`+sy+`', q.y`+sy)
	}
	sfy := strconv.FormatInt(firstYear, 10)
	sly := strconv.FormatInt(lastYear.Int64, 10)
	operationCrossQuery := `SELECT json_build_object('id',q.id,'number',q.number, 
		'name', q.name,` + strings.Join(jsonNames, ",") + ` ) FROM
	(SELECT op.id, op.number, op.name, ` + strings.Join(columnNames, ",") + ` 
		FROM physical_op op, 
			(SELECT * FROM 
				crosstab ('SELECT p.id, c.year, c.value FROM physical_op p
									 LEFT OUTER JOIN prev_commitment c ON c.physical_op_id=p.id ORDER BY 1',
									'SELECT m FROM generate_series(` + sfy + `, ` + sly + `) AS m')
				AS ( id INTEGER, ` + strings.Join(typesNames, ",") + `)
			) AS ppi WHERE op.id=ppi.id) q;`
	scenarioCrossQuery := `SELECT json_build_object('id',q.id,'number',q.number, 
		'name', q.name, 'offset', q.offset, ` + strings.Join(jsonNames, ",") + ` ) FROM
	(SELECT op.id, op.number, op.name, sc.offset, ` + strings.Join(columnNames, ",") + `
		FROM physical_op op, scenario_offset sc, 
			(SELECT * FROM 
				crosstab ('SELECT p.id, c.year, c.value FROM physical_op p 
									 LEFT OUTER JOIN prev_commitment c ON c.physical_op_id=p.id ORDER BY 1',
		 							'SELECT m FROM generate_series(` + sfy + `, ` + sly + `) m')
				AS (id INTEGER, ` + strings.Join(typesNames, ",") + `)
			) AS ppi WHERE op.id=ppi.id AND sc.physical_op_id=op.id AND sc.scenario_id = $1) q;`
	lines, line := []string{}, ""
	rows, err := db.Query(operationCrossQuery)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&line); err != nil {
			return err
		}
		lines = append(lines, line)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	d.OperationCrossTable = json.RawMessage("[" + strings.Join(lines, ",") + "]")
	rows, err = db.Query(scenarioCrossQuery, sID)
	if err != nil {
		return err
	}
	lines = nil
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&line); err != nil {
			return err
		}
		lines = append(lines, line)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	d.ScenarioCrossTable = json.RawMessage("[" + strings.Join(lines, ",") + "]")
	return nil
}

// GetAll populate MultiAnnualBudgetScenario from database
func (m *MultiAnnualBudgetScenario) GetAll(year int64, scenarioID int64, db *sql.DB) (err error) {
	sy := strconv.FormatInt(year, 10)
	ssID := strconv.FormatInt(scenarioID, 10)
	rows, err := db.Query(`SELECT op.number, op.name, bt.chapter, bt.sector, bt.subfunction, 
		bt.program, bt.action, sc.y0, sc.y1, sc.y2, sc.y3, sc.y4
	FROM physical_op AS op
	JOIN (SELECT * FROM
	crosstab ('SELECT op.id, p.year+s.offset AS year, p.value FROM physical_op op 
					JOIN scenario_offset s 
					ON s.physical_op_id = op.id AND s.scenario_id = ` + ssID + ` 
					LEFT OUTER JOIN prev_commitment p ON p.physical_op_id = op.id ORDER BY 1,2',
					'SELECT m FROM generate_series(` + sy + `, ` + sy + ` + 4) AS m')
	AS (id INTEGER, y0 BIGINT, y1 BIGINT, y2 BIGINT, y3 BIGINT, y4 BIGINT)) sc 
	ON sc.id = op.id
	LEFT OUTER JOIN 
	(SELECT ba.id, bc.code AS chapter, bs.code AS sector,
		bp.code_function || bp.code_subfunction AS subfunction,
		bp.code_contract || bp.code_function || bp.code_number AS program, 
		bp.code_contract || bp.code_function || bp.code_number || ba.code AS action 
	FROM budget_chapter AS bc, budget_program AS bp, budget_action AS ba, budget_sector AS bs 
	WHERE bp.id = ba.program_id  AND bc.id=bp.chapter_id AND bs.id = ba.sector_id) AS bt
	ON bt.id = op.budget_action_id 
	ORDER BY chapter, sector, subfunction, program, action, number`)
	if err != nil {
		return err
	}
	var r MABScenarioLine
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.Number, &r.Name, &r.Chapter, &r.Sector, &r.Subfunction,
			&r.Program, &r.Action, &r.Y0, &r.Y1, &r.Y2, &r.Y3, &r.Y4); err != nil {
			return err
		}
		m.MultiAnnualBudgetScenario = append(m.MultiAnnualBudgetScenario, r)
	}
	err = rows.Err()
	if len(m.MultiAnnualBudgetScenario) == 0 {
		m.MultiAnnualBudgetScenario = []MABScenarioLine{}
	}
	return err
}
