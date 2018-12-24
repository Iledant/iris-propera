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
	ID       int64      `json:"id" gorm:"column:id"`
	Name     string     `json:"name" gorm:"column:name"`
	Descript NullString `json:"descript" gorm:"column:descript"`
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
	return err
}

// Invalid checks of Scenario's field can be saved to database.
func (s *Scenario) Invalid() bool {
	return s.Name == "" || len(s.Name) > 255
}

// Create insert a new scenario into database.
func (s *Scenario) Create(db *sql.DB) (err error) {
	err = db.QueryRow("INSERT INTO scenario (name,descript) VALUES($1,$2) RETURNING id", s.Name, s.Descript).Scan(&s.ID)
	return err
}

// Update modifies a scenario into database.
func (s *Scenario) Update(db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE scenario SET name=$1, descript=$2 WHERE id = $3`, s.Name, s.Descript, s.ID)
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
	if _, err = tx.Exec("DELETE from scenario_offset WHERE scenario_id = $1", s.ID); err != nil {
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
	var lastYear int64
	if err = db.QueryRow(`SELECT max(p.year+s.offset)::bigint AS year FROM prev_commitment p, scenario_offset s 
	 WHERE s.physical_op_id = p.physical_op_id AND s.scenario_id= $1`, sID).Scan(&lastYear); err != nil {
		return err
	}
	var columnNames, typesNames, jsonNames []string

	for i := firstYear; i <= lastYear; i++ {
		sy := strconv.FormatInt(i-firstYear, 10)
		columnNames = append(columnNames, `"`+strconv.FormatInt(i, 10)+`" AS y`+sy)
		typesNames = append(typesNames, `"`+strconv.FormatInt(i, 10)+`" NUMERIC`)
		jsonNames = append(jsonNames, `'y`+sy+`', q.y`+sy)
	}
	operationCrossQuery := `SELECT json_build_object('id',q.id,'number',q.number, 'name', q.name,` + strings.Join(jsonNames, ",") + ` ) FROM
	(SELECT op.id, op.number, op.name, ` + strings.Join(columnNames, ",") + ` 
	FROM physical_op op, (SELECT * FROM 
		crosstab ('SELECT p.id, c.year, c.value FROM physical_op p LEFT OUTER JOIN prev_commitment c ON c.physical_op_id = p.id ORDER BY 1',
							'SELECT m FROM generate_series(` + strconv.FormatInt(firstYear, 10) + `, ` + strconv.FormatInt(lastYear, 10) + `) AS m') AS
			( id INTEGER, ` + strings.Join(typesNames, ",") + `))
		AS ppi WHERE op.id=ppi.id) q;`
	scenarioCrossQuery := `SELECT json_build_object('id',q.id,'number',q.number, 'name', q.name, 'offset', q.offset, ` + strings.Join(jsonNames, ",") + ` ) FROM
	(SELECT op.id, op.number, op.name, sc.offset, ` + strings.Join(columnNames, ",") + `
	FROM physical_op op, scenario_offset sc, (SELECT * FROM 
		crosstab ('SELECT p.id, c.year, c.value FROM physical_op p LEFT OUTER JOIN prev_commitment c ON c.physical_op_id = p.id ORDER BY 1',
		 'SELECT m FROM generate_series(` + strconv.FormatInt(firstYear, 10) + `, ` + strconv.FormatInt(lastYear, 10) + `) m') AS
			(id INTEGER, ` + strings.Join(typesNames, ",") + `))
		AS ppi WHERE op.id=ppi.id AND sc.physical_op_id = op.id AND sc.scenario_id = $1) q;`
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
