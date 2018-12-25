package models

import (
	"database/sql"
	"strings"
)

// ScenarioOffset model
type ScenarioOffset struct {
	ID           int64 `json:"id"`
	ScenarioID   int64 `json:"scenario_id"`
	PhysicalOpID int64 `json:"physical_op_id"`
	Offset       int64 `json:"offset"`
}

// ScenarioOffsets embeddes an array of ScenarioOffset
type ScenarioOffsets struct {
	ScenarioOffsets []ScenarioOffset `json:"offsetList"`
}

// Save replaces offsets of a scenario.
func (s *ScenarioOffsets) Save(sID int64, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err = tx.Exec("DELETE FROM scenario_offset WHERE scenario_id=$1", sID); err != nil {
		tx.Rollback()
		return err
	}
	var values []string
	syID := toSQL(sID)
	for _, o := range s.ScenarioOffsets {
		values = append(values, `(`+toSQL(o.Offset)+`,`+toSQL(o.PhysicalOpID)+`,`+syID+`)`)
	}
	if _, err = tx.Exec(`INSERT INTO scenario_offset ("offset",physical_op_id,scenario_id)
	 VALUES ` + strings.Join(values, ",")); err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}
