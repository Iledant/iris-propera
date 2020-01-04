package models

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
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

	stmt, err := tx.Prepare(pq.CopyIn("scenario_offset", "offset",
		"physical_op_id", "scenario_id"))
	if err != nil {
		return fmt.Errorf("prepare stmt %v", err)
	}
	defer stmt.Close()
	for _, r := range s.ScenarioOffsets {
		if _, err = stmt.Exec(r.Offset, r.PhysicalOpID, sID); err != nil {
			tx.Rollback()
			return fmt.Errorf("insertion de %+v  %v", r, err)
		}
	}
	if _, err = stmt.Exec(); err != nil {
		tx.Rollback()
		return fmt.Errorf("statement exec flush %v", err)
	}

	err = tx.Commit()
	return err
}
