package models

import (
	"database/sql"
	"errors"
)

// Document model
type Document struct {
	ID           int64  `json:"id"`
	PhysicalOpID int64  `json:"physical_op_id"`
	Name         string `json:"name"`
	Link         string `json:"link"`
}

// Documents embeddes an array of documents for json exports.
type Documents struct {
	Documents []Document `json:"Document"`
}

// Validate checks if fields are correctly formed.
func (d *Document) Validate() error {
	if d.PhysicalOpID == 0 || d.Name == "" || d.Link == "" {
		return errors.New("PhysicalOpID, Name ou Link incorrect")
	}
	return nil
}

// GetOpAll fetches all documents linked to a physical operation from database.
func (d *Documents) GetOpAll(pID int64, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id,physical_op_id,name,link FROM documents
	 WHERE physical_op_id = $1`, pID)
	if err != nil {
		return err
	}
	var r Document
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.PhysicalOpID, &r.Name, &r.Link); err != nil {
			return err
		}
		d.Documents = append(d.Documents, r)
	}
	err = rows.Err()
	return err
}

// Create insert a new document into database.
func (d *Document) Create(db *sql.DB) (err error) {
	err = db.QueryRow(`INSERT INTO documents (physical_op_id,name,link)
	 VALUES($1,$2,$3) RETURNING id`,
		d.PhysicalOpID, d.Name, d.Link).Scan(&d.ID)
	return err
}

// Update modifies a document in the database.
func (d *Document) Update(db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE documents SET physical_op_id=$1, name=$2,
	 link=$3 WHERE id=$4`,
		d.PhysicalOpID, d.Name, d.Link, d.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Document introuvable")
	}
	return err
}

// Delete removes a document from database.
func (d *Document) Delete(db *sql.DB) (err error) {
	res, err := db.Exec("DELETE FROM documents WHERE id=$1", d.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Document introuvable")
	}
	return nil
}
