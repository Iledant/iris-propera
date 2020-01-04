package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"strconv"
	"time"
)

var b = time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)

type jsonError struct {
	Erreur string `json:"Erreur"`
}

// ExcelDate is used for batch imports to decode an integer and transform it
// into a SQL date using toSQL function. 0 value is coded to "null" by toSQL
type ExcelDate int64

// NullExcelDate is used for batch import to decde an nullable integer and
// transfort it into en nullable SQL date using toSQL function
type NullExcelDate struct {
	Valid bool
	Date  int64
}

// ToDate calculates the date according to the int64 ExcelDate
func (e ExcelDate) ToDate() time.Time {
	return b.Add(time.Duration(e*24) * time.Hour)
}

// ToDate calculates the null date according to the int64 ExcelDate if valid
func (ne NullExcelDate) ToDate() NullTime {
	if !ne.Valid {
		return NullTime{Valid: false}
	}
	return NullTime{
		Valid: true,
		Time:  b.Add(time.Duration(ne.Date*24) * time.Hour)}
}

// NullTime is used for null time commands
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

// UnmarshalJSON implements the unmarshal interface
func (ne *NullExcelDate) UnmarshalJSON(b []byte) error {
	if len(b) == 4 && b[0] == 110 && b[1] == 117 && b[2] == 108 && b[3] == 108 {
		ne.Valid = false
		return nil
	}
	err := json.Unmarshal(b, &ne.Date)
	ne.Valid = (err == nil)
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
	var n sql.NullBool
	if err := n.Scan(value); err != nil {
		return err
	}
	nb.Bool, nb.Valid = n.Bool, n.Valid
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
	var n sql.NullInt64
	if err := n.Scan(value); err != nil {
		return err
	}
	ni.Int64, ni.Valid = n.Int64, n.Valid
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
	var n sql.NullString
	if err := n.Scan(value); err != nil {
		return err
	}
	ns.String, ns.Valid = n.String, n.Valid
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
	var n sql.NullFloat64
	if err := n.Scan(value); err != nil {
		return err
	}
	nf.Float64, nf.Valid = n.Float64, n.Valid
	return nil
}

// Value implements the driver Valuer interface
func (nf NullFloat64) Value() (driver.Value, error) {
	if !nf.Valid {
		return nil, nil
	}
	return nf.Float64, nil
}

// toSQL convert i to a string used for INSERT VALUES statement.
func toSQL(i interface{}) string {
	switch v := i.(type) {
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return "$$" + v + "$$"
	case time.Time:
		return "'" + v.Format("2006-01-02") + "'"
	case ExcelDate:
		if v == 0 {
			return "null"
		}
		return "'" + b.Add(time.Duration(v*24)*time.Hour).Format("2006-01-02") + "'"
	case int:
		return strconv.Itoa(v)
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	case NullInt64:
		if !v.Valid {
			return "null"
		}
		return strconv.FormatInt(v.Int64, 10)
	case NullString:
		if !v.Valid {
			return "null"
		}
		return "$$" + v.String + "$$"
	case NullFloat64:
		if !v.Valid {
			return "null"
		}
		return strconv.FormatFloat(v.Float64, 'f', -1, 64)
	case NullBool:
		if !v.Valid {
			return "null"
		}
		if v.Bool {
			return "TRUE"
		}
		return "FALSE"
	case NullTime:
		if !v.Valid {
			return "null"
		}
		return "'" + v.Time.Format("2006-01-02") + "'"
	case NullExcelDate:
		if !v.Valid {
			return "null"
		}
		return "'" + b.Add(time.Duration(v.Date*24)*time.Hour).Format("2006-01-02") + "'"
	}
	return ""
}
