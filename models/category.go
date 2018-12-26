package models

import (
	"database/sql"
	"errors"
)

// Category model
type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Categories embeddes an array of categories for json export.
type Categories struct {
	Categories []Category `json:"Category"`
}

// Validate checks if fields are well formed.
func (c *Category) Validate() error {
	if c.Name == "" || len(c.Name) > 50 {
		return errors.New("Name invalide")
	}
	return nil
}

// GetAll fetches all catégories from database.
func (c *Categories) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id, name FROM category`)
	if err != nil {
		return err
	}
	var r Category
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Name); err != nil {
			return err
		}
		c.Categories = append(c.Categories, r)
	}
	err = rows.Err()
	if len(c.Categories) == 0 {
		c.Categories = []Category{}
	}
	return err
}

// Create insert a new category into database.
func (c *Category) Create(db *sql.DB) (err error) {
	err = db.QueryRow("INSERT INTO category (name) VALUES($1) RETURNING id",
		c.Name).Scan(&c.ID)
	return err
}

// Update modify a category in database.
func (c *Category) Update(db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE category SET name=$1 WHERE id=$2`,
		c.Name, c.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Catégorie introuvable")
	}
	return err
}

// Delete removes a category from database.
func (c *Category) Delete(db *sql.DB) (err error) {
	res, err := db.Exec("DELETE FROM category WHERE id = $1", c.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Catégorie introuvable")
	}
	return nil
}
