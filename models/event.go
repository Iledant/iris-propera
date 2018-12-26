package models

import (
	"database/sql"
	"errors"
	"time"
)

// Event model
type Event struct {
	ID           int64      `json:"id"`
	PhysicalOpID int64      `json:"physical_op_id"`
	Name         string     `json:"name"`
	Date         time.Time  `json:"date"`
	IsCertain    bool       `json:"iscertain"`
	Descript     NullString `json:"descript"`
}

// Events embeddes an array of documents for json export.
type Events struct {
	Events []Event `json:"Event"`
}

// Validate checks if fields are correctly formed.
func (e *Event) Validate() error {
	if e.PhysicalOpID == 0 || e.Name == "" || e.Date.IsZero() {
		return errors.New("PhysicalOpID, Name ou Date incorrect")
	}
	return nil
}

// GetOpAll fetches all events of a physical operation from database.
func (e *Events) GetOpAll(opID int64, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id,physical_op_id,name,date,iscertain,descript 
	FROM event WHERE physical_op_id=$1`, opID)
	if err != nil {
		return err
	}
	var r Event
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.PhysicalOpID, &r.Name, &r.Date,
			&r.IsCertain, &r.Descript); err != nil {
			return err
		}
		e.Events = append(e.Events, r)
	}
	err = rows.Err()
	return err
}

// Create insert a new event into database.
func (e *Event) Create(db *sql.DB) (err error) {
	err = db.QueryRow(`INSERT INTO event (physical_op_id,name,date,iscertain,descript)
	 VALUES($1,$2,$3,$4,$5) RETURNING id`,
		e.PhysicalOpID, e.Name, e.Date, e.IsCertain, e.Descript).Scan(&e.ID)
	return err
}

// Update modify an event in the database.
func (e *Event) Update(db *sql.DB) (err error) {
	res, err := db.Exec(`UPDATE event SET physical_op_id=$1, name=$2, date=$3, 
	iscertain=$4, descript=$5 WHERE id = $6`,
		e.PhysicalOpID, e.Name, e.Date, e.IsCertain, e.Descript, e.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Événement introuvable")
	}
	return err
}

// Delete removes en event from database.
func (e *Event) Delete(db *sql.DB) (err error) {
	res, err := db.Exec("DELETE FROM event WHERE id = $1", e.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Événement introuvable")
	}
	return nil
}

// NextMonthEvent embeddes datas for next month events request.
type NextMonthEvent struct {
	ID        int       `json:"id"`
	Date      time.Time `json:"date"`
	Event     string    `json:"event"`
	Operation string    `json:"operation"`
}

// NextMonthEvents embeddes an array of NextMonthEvent for json export.
type NextMonthEvents struct {
	NextMonthEvents []NextMonthEvent `json:"Event"`
}

// Get fetches next month events from database according to user ID, 0 is dedicated to admin.
func (n *NextMonthEvents) Get(uID int64, db *sql.DB) (err error) {
	var rows *sql.Rows
	query := `SELECT e.id, e.date, o.name AS operation, e.name AS event 
	  FROM event e, physical_op o 
		WHERE e.date < CURRENT_DATE + interval '1 month' AND e.date >= CURRENT_DATE 
			AND e.physical_op_id=o.id`
	if uID != 0 {
		query = query + ` AND o.id IN (SELECT rights.physical_op_id FROM rights 
			WHERE rights.users_id = $1)`
		rows, err = db.Query(query, uID)
	} else {
		rows, err = db.Query(query)
	}
	if err != nil {
		return err
	}
	var r NextMonthEvent
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Date, &r.Event, &r.Operation); err != nil {
			return err
		}
		n.NextMonthEvents = append(n.NextMonthEvents, r)
	}
	if len(n.NextMonthEvents) == 0 {
		n.NextMonthEvents = []NextMonthEvent{}
	}
	err = rows.Err()
	return err
}
