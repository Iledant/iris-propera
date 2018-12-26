package models

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
)

// MultiannualProg embeddes an array of bytes for json export.
type MultiannualProg struct {
	MultiannualProg json.RawMessage `json:"MultiannualProg"`
}

// GetAll fetches multi annual programmation from database.
func (m *MultiannualProg) GetAll(firstYear int64, db *sql.DB) (err error) {
	var lastYear int64
	if err = db.QueryRow("SELECT max(year) FROM prev_commitment").Scan(&lastYear); err != nil {
		return err
	}
	if lastYear < firstYear+4 {
		lastYear = firstYear + 4
	}
	if _, err := db.Exec("CREATE EXTENSION IF NOT EXISTS tablefunc"); err != nil {
		return err
	}
	var columnNames, typesNames, jsonNames []string
	for i := firstYear; i <= lastYear; i++ {
		year := strconv.FormatInt(i-firstYear, 10)
		columnNames = append(columnNames, `"`+strconv.FormatInt(i, 10)+`" AS y`+year)
		typesNames = append(typesNames, `"`+strconv.FormatInt(i, 10)+`" VARCHAR`)
		jsonNames = append(jsonNames, `'y`+year+`', q.y`+year)
	}
	qry := `SELECT json_build_object('number',q.number,'name',q.name,'step_name', q.step_name, 'category_name',
	 q.category_name,` + strings.Join(jsonNames, ",") + `) FROM
	(SELECT op.number, op.name, s.name as step_name, cat.name as category_name, ` + strings.Join(columnNames, ",") + ` FROM 
	crosstab('SELECT physical_op_id, year, row_to_json((SELECT d FROM (SELECT value, total_value, state_ratio) d)) 
						FROM prev_commitment ORDER BY 1,2', 
					'SELECT m FROM generate_series(` + strconv.FormatInt(firstYear, 10) + `,` + strconv.FormatInt(lastYear, 10) + `) AS m') AS
					(op_id INTEGER, ` + strings.Join(typesNames, ",") + `)
	JOIN physical_op op ON op.id = op_id
	LEFT OUTER JOIN step s ON op.step_id = s.id
	LEFT OUTER JOIN category cat ON op.category_id = cat.id) q`
	rows, err := db.Query(qry)
	if err != nil {
		return err
	}
	var r string
	var rr []string
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r); err != nil {
			return err
		}
		rr = append(rr, r)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	m.MultiannualProg = json.RawMessage("[" + strings.Join(rr, ",") + "]")
	return nil
}
