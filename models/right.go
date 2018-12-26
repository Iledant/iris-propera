package models

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"
)

// Right model for the right of a user on physical operations.
type Right struct {
	ID           int `json:"id"`
	PhysicalOpID int `json:"physical_op_id"`
	UserID       int `json:"users_id"`
}

// OpRights embeddes an array of physical operation IDs to set rights of a user.
type OpRights struct {
	OpIDs []int64 `json:"Right"`
}

// UsersIDs is used to set inherit rights on physical operation and embeddes an
// array id user IDs.
type UsersIDs struct {
	UsersIDs []int64 `json:"Right"`
}

// UserSet replaces rights in the database returning an error if user ID or
// physical operation ID doesn't exist.
func (o *OpRights) UserSet(uID int64, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err = tx.Exec("DELETE from rights WHERE users_id = $1", uID); err != nil {
		tx.Rollback()
		return err
	}
	if len(o.OpIDs) > 0 {
		var value string
		var values []string
		for _, opID := range o.OpIDs {
			value = "(" + toSQL(uID) + "," + toSQL(opID) + ")"
			values = append(values, value)
		}
		res, err := tx.Exec("INSERT INTO rights (users_id, physical_op_id) VALUES" +
			strings.Join(values, ","))
		if err != nil {
			tx.Rollback()
			return err
		}
		count, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			return err
		}
		if count != int64(len(o.OpIDs)) {
			tx.Rollback()
			err = errors.New("Erreur d'utilisateur ou d'identifiant d'opérations")
			return err
		}
	}
	err = tx.Commit()
	return err
}

// UserGet fetches user's rights form database.
func (o *OpRights) UserGet(uID int64, db *sql.DB) (err error) {
	rows, err := db.Query("SELECT physical_op_id FROM rights WHERE users_id = $1", uID)
	if err != nil {
		return err
	}
	defer rows.Close()
	var opID int64
	for rows.Next() {
		if err = rows.Scan(&opID); err != nil {
			return err
		}
		o.OpIDs = append(o.OpIDs, opID)
	}
	err = rows.Err()
	if len(o.OpIDs) == 0 {
		o.OpIDs = []int64{}
	}
	return err
}

// Inherit updates the user's right with those from sent users.
func (o *UsersIDs) Inherit(uID int64, db *sql.DB) (err error) {
	_, err = db.Exec(`INSERT INTO rights (users_id, physical_op_id) SELECT $1,* FROM 
	(SELECT DISTINCT physical_op_id FROM rights WHERE users_id=ANY($2) ) ids 
	 WHERE ids.physical_op_id NOT IN (SELECT physical_op_id FROM rights WHERE users_id=$1)`,
		uID, pq.Array(o.UsersIDs))
	return err
}
