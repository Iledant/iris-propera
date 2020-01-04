package models

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

// PlanLine model
type PlanLine struct {
	ID         int64      `json:"id"`
	PlanID     int64      `json:"plan_id"`
	Name       string     `json:"name"`
	Descript   NullString `json:"descript"`
	Value      int64      `json:"value"`
	TotalValue NullInt64  `json:"total_value"`
}

// PlanLines embeddes an array of PlanLine for json export.
type PlanLines struct {
	PlanLines []PlanLine `json:"PlanLine"`
}

// PlanLineBatch is used to decode a batch of plan lines that can have variables
// number of fields to store beneficiaries ratios.
type PlanLineBatch struct {
	PlanLines []map[string]interface{} `json:"PlanLine"`
}

// LinkFCs updates the financial commitments linked
// to a physical operation in database.
func (p *PlanLine) LinkFCs(fcIDs []int64, db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE financial_commitment SET plan_line_id = $1 
	WHERE id = ANY($2)`, p.ID, pq.Array(fcIDs))
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != int64(len(fcIDs)) {
		return errors.New("Ligne de plan ou engagements incorrects")
	}
	return nil
}

// Delete removes the plan lines from database including linked plan line ratios.
func (p *PlanLine) Delete(db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err = tx.Exec("DELETE FROM plan_line_ratios WHERE plan_line_id = $1",
		p.ID); err != nil {
		tx.Rollback()
		return err
	}
	res, err := tx.Exec("DELETE FROM plan_line WHERE id=$1", p.ID)
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
		return errors.New("Ligne de plan introuvable")
	}
	tx.Commit()
	return err
}

// Create insert a new plan line and it's linked ratios into database.
func (p *PlanLine) Create(plr *PlanLineRatios, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	err = tx.QueryRow(`INSERT INTO plan_line (plan_id,name,descript,value,total_value) 
	VALUES($1,$2,$3,$4,$5) RETURNING id`, p.PlanID, p.Name, p.Descript,
		p.Value, p.TotalValue).Scan(&p.ID)
	if err != nil {
		tx.Rollback()
		return err
	}
	if err = plr.Save(p.ID, tx); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return err
}

// GetByID fetches the plan line whose ID is given from database.
func (p *PlanLine) GetByID(db *sql.DB) (err error) {
	err = db.QueryRow(`SELECT id, plan_id, name, descript, value, total_value 
	FROM plan_line WHERE id=$1`, p.ID).Scan(&p.ID, &p.PlanID, &p.Name, &p.Descript,
		&p.Value, &p.TotalValue)
	return err
}

// Update modifies a plan line and it's ratio into the database.
func (p *PlanLine) Update(plr *PlanLineRatios, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	res, err := tx.Exec(`UPDATE plan_line SET plan_id=$1,name=$2,descript=$3,
	value=$4,total_value=$5 WHERE id=$6`, p.PlanID, p.Name, p.Descript, p.Value,
		p.TotalValue, p.ID)
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
		return errors.New("Ligne de plan introuvable")
	}
	if err = plr.Save(p.ID, tx); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return err
}

// Save insert plan lines and their beneficiary's ratios into database.
func (p *PlanLineBatch) Save(planID int64, db *sql.DB) (err error) {
	if len(p.PlanLines) == 0 {
		return nil
	}
	var bKeys []string
	for key := range p.PlanLines[0] {
		_, err = strconv.Atoi(key)
		if err == nil {
			bKeys = append(bKeys, key)
		}
	}
	var value, sqlDescript, sqlTotalValue string
	var values []string
	for _, l := range p.PlanLines {
		if _, ok := l["name"]; !ok || l["name"] == nil {
			return errors.New("Colonne name manquante")
		}
		if _, ok := l["value"]; !ok || l["value"] == nil {
			return errors.New("Colonne value manquante")
		}
		if descript, ok := l["descript"]; !ok || descript == nil {
			sqlDescript = "null"
		} else {
			sqlDescript = "'" + descript.(string) + "'"
		}
		if totalValue, ok := l["total_value"]; !ok || totalValue == nil {
			sqlTotalValue = "null"
		} else {
			sqlTotalValue = strconv.FormatInt(int64(100*totalValue.(float64)), 10)
		}
		value = "('" + l["name"].(string) + "'," + sqlDescript + "," +
			strconv.FormatInt(int64(100*l["value"].(float64)), 10) + "," + sqlTotalValue + ")"
		values = append(values, value)
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	queries := []string{`DROP TABLE IF EXISTS temp_plan_line`,
		`CREATE TABLE temp_plan_line (name varchar(255), descript text, 
		value bigint, total_value bigint)`,
		`INSERT INTO temp_plan_line (name,descript,value,total_value) VALUES ` +
			strings.Join(values, ",")}
	for _, qry := range queries {
		if _, err = tx.Exec(qry); err != nil {
			tx.Rollback()
			return err
		}
	}
	_, err = tx.Exec(`INSERT INTO plan_line (plan_id, name, descript, value, total_value) 
	SELECT $1 AS plan_id, * FROM temp_plan_line t 
	WHERE (t.name) NOT IN (SELECT name FROM plan_line WHERE plan_id=$1)`, planID)
	if err != nil {
		tx.Rollback()
		return err
	}
	if len(bKeys) > 0 {
		var planLineID int64
		var sPlID, sRatio string
		for _, l := range p.PlanLines {
			if err = tx.QueryRow(`SELECT id FROM plan_line WHERE name=$1 AND plan_id=$2`,
				l["name"].(string), planID).Scan(&planLineID); err != nil {
				tx.Rollback()
				return err
			}
			if _, err = tx.Exec(`DELETE FROM plan_line_ratios WHERE plan_line_id=$1`,
				planLineID); err != nil {
				tx.Rollback()
				return err
			}
			sPlID = strconv.FormatInt(planLineID, 10)
			values = nil
			for _, k := range bKeys {
				ratio, ok := l[k]
				if !ok || ratio == nil {
					continue
				}
				sRatio = strconv.FormatFloat(ratio.(float64), 'f', -1, 64)
				values = append(values, "("+sPlID+","+k+","+sRatio+")")
			}
			if len(values) == 0 {
				continue
			}
			if _, err = tx.Exec(`INSERT INTO plan_line_ratios (plan_line_id,beneficiary_id,
				ratio) VALUES` + strings.Join(values, ",")); err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	if _, err = tx.Exec(`DROP TABLE IF EXISTS temp_plan_line`); err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}
