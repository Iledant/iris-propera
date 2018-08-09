package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"
)

// NullTime is used for gorm null time commands
type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

// Scan implements the Scanner interface
func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Valid = false
		return nil
	}
	nt.Time, nt.Valid = value.(time.Time), true
	return nil
}

// Value implements the driver Valuer interface
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

// MarshalJSON implents the marshall interface
func (nt NullTime) MarshalJSON() ([]byte, error) {
	if nt.Valid == false {
		return []byte("null"), nil
	}
	return nt.Time.MarshalJSON()
}

// UnmarshalJSON implents the unmarshal interface
func (nt *NullTime) UnmarshalJSON(b []byte) error {
	if len(b) == 4 && b[0] == 110 && b[1] == 117 && b[2] == 108 && b[3] == 108 {
		nt.Valid = false
		return nil
	}
	err := json.Unmarshal(b, &nt.Time)
	nt.Valid = (err == nil)
	return err
}

// NullBool uses alias in order to mashall and un marshall correctly
type NullBool sql.NullBool

// MarshalJSON implents the marshall interface
func (nb NullBool) MarshalJSON() ([]byte, error) {
	if nb.Valid == false {
		return []byte("null"), nil
	}

	return json.Marshal(nb.Bool)
}

// UnmarshalJSON implents the unmarshal interface
func (nb *NullBool) UnmarshalJSON(b []byte) error {
	if len(b) == 4 && b[0] == 110 && b[1] == 117 && b[2] == 108 && b[3] == 108 {
		nb.Valid = false
		return nil
	}
	err := json.Unmarshal(b, &nb.Bool)
	nb.Valid = (err == nil)
	return err
}

// Scan implements the Scanner interface
func (nb *NullBool) Scan(value interface{}) error {
	if value == nil {
		nb.Valid = false
		return nil
	}
	nb.Bool, nb.Valid = value.(bool), true
	return nil
}

// Value implements the driver Valuer interface
func (nb NullBool) Value() (driver.Value, error) {
	if !nb.Valid {
		return nil, nil
	}
	return nb.Bool, nil
}

// NullInt64 uses alias in order to mashall and un marshall correctly
type NullInt64 sql.NullInt64

// MarshalJSON implents the marshall interface
func (ni NullInt64) MarshalJSON() ([]byte, error) {
	if ni.Valid == false {
		return []byte("null"), nil
	}

	return json.Marshal(ni.Int64)
}

// UnmarshalJSON implents the unmarshal interface
func (ni *NullInt64) UnmarshalJSON(b []byte) error {
	if len(b) == 4 && b[0] == 110 && b[1] == 117 && b[2] == 108 && b[3] == 108 {
		ni.Valid = false
		return nil
	}
	err := json.Unmarshal(b, &ni.Int64)
	ni.Valid = (err == nil)
	return err
}

// Scan implements the Scanner interface
func (ni *NullInt64) Scan(value interface{}) error {
	if value == nil {
		ni.Valid = false
		return nil
	}
	ni.Int64, ni.Valid = value.(int64), true
	return nil
}

// Value implements the driver Valuer interface
func (ni NullInt64) Value() (driver.Value, error) {
	if !ni.Valid {
		return nil, nil
	}
	return ni.Int64, nil
}

// NullString uses alias in order to mashall and un marshall correctly
type NullString sql.NullString

// MarshalJSON implents the marshall interface
func (ns NullString) MarshalJSON() ([]byte, error) {
	if ns.Valid == false {
		return []byte("null"), nil
	}

	return json.Marshal(ns.String)
}

// UnmarshalJSON implents the unmarshal interface
func (ns *NullString) UnmarshalJSON(b []byte) error {
	if len(b) == 4 && b[0] == 110 && b[1] == 117 && b[2] == 108 && b[3] == 108 {
		ns.Valid = false
		return nil
	}
	err := json.Unmarshal(b, &ns.String)
	ns.Valid = (err == nil)
	return err
}

// Scan implements the Scanner interface
func (ns *NullString) Scan(value interface{}) error {
	if value == nil {
		ns.Valid = false
		return nil
	}
	ns.String, ns.Valid = value.(string), true
	return nil
}

// Value implements the driver Valuer interface
func (ns NullString) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.String, nil
}

// NullFloat64 uses alias in order to mashall and un marshall correctly
type NullFloat64 sql.NullFloat64

// MarshalJSON implents the marshall interface
func (nf NullFloat64) MarshalJSON() ([]byte, error) {
	if nf.Valid == false {
		return []byte("null"), nil
	}

	return json.Marshal(nf.Float64)
}

// UnmarshalJSON implents the unmarshal interface
func (nf *NullFloat64) UnmarshalJSON(b []byte) error {
	if len(b) == 4 && b[0] == 110 && b[1] == 117 && b[2] == 108 && b[3] == 108 {
		nf.Valid = false
		return nil
	}
	err := json.Unmarshal(b, &nf.Float64)
	nf.Valid = (err == nil)
	return err
}

// Scan implements the Scanner interface
func (nf *NullFloat64) Scan(value interface{}) error {
	if value == nil {
		nf.Valid = false
		return nil
	}
	nf.Float64, nf.Valid = value.(float64), true
	return nil
}

// Value implements the driver Valuer interface
func (nf NullFloat64) Value() (driver.Value, error) {
	if !nf.Valid {
		return nil, nil
	}
	return nf.Float64, nil
}
