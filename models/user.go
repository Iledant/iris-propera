package models

import (
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ErrBadCredential is the error of a user whose login ou password is not in the
// user database
var ErrBadCredential = errors.New("Erreur de login ou de mot de passe")

// User model
type User struct {
	ID            int        `json:"id"`
	Created       NullTime   `json:"created_at"`
	Updated       NullTime   `json:"updated_at"`
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	Password      string     `json:"-"`
	Role          string     `json:"role"`
	RememberToken NullString `json:"-"`
	Active        bool       `json:"active"`
}

// Users embeddes an array of User for json export.
type Users struct {
	Users []User `json:"User"`
}

const (
	// AdminRole defines value of role row in users table for an admin
	AdminRole = "ADMIN"
	// ObserverRole defines value of role row in users table for an observer
	ObserverRole = "OBSERVER"
	// UserRole defines value of role row in users table for an usual user
	UserRole = "USER"
)

// GetByID fetches a user from database using ID.
func (u *User) GetByID(db *sql.DB) (err error) {
	err = db.QueryRow(`SELECT id, created_at, updated_at, name, email, role, 
	password, active FROM users WHERE id = $1 LIMIT 1`, u.ID).Scan(&u.ID,
		&u.Created, &u.Updated, &u.Name, &u.Email, &u.Role, &u.Password, &u.Active)
	return err
}

// CryptPwd crypt not codded password field.
func (u *User) CryptPwd() (err error) {
	cryptPwd, err := bcrypt.GenerateFromPassword([]byte(u.Password), 10)
	if err != nil {
		return err
	}
	u.Password = string(cryptPwd)
	return nil
}

// ValidatePwd compared sent uncodded password with internal password.
func (u *User) ValidatePwd(pwd string) error {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(pwd))
	if err != nil && errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrBadCredential
	}
	return err
}

// GetAll fetches all users from database.
func (users *Users) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id, created_at, updated_at, name, email, role,
	active FROM users`)
	if err != nil {
		return err
	}
	var r User
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Created, &r.Updated, &r.Name, &r.Email,
			&r.Role, &r.Active); err != nil {
			return err
		}
		users.Users = append(users.Users, r)
	}
	err = rows.Err()
	if len(users.Users) == 0 {
		users.Users = []User{}
	}
	return err
}

//GetRole fetches all users according to a role.
func (users *Users) GetRole(role string, db *sql.DB) (err error) {
	rows, err := db.Query(`SELECT id, created_at, updated_at, name, email, role,
	active FROM users WHERE role = $1`, role)
	if err != nil {
		return err
	}
	var r User
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Created, &r.Updated, &r.Name, &r.Email,
			&r.Role, &r.Active); err != nil {
			return err
		}
		users.Users = append(users.Users, r)
	}
	err = rows.Err()
	if len(users.Users) == 0 {
		users.Users = []User{}
	}
	return err
}

// GetByEmail fetches an user by email.
func (u *User) GetByEmail(email string, db *sql.DB) (err error) {
	err = db.QueryRow(`SELECT id, created_at, updated_at, name, email, role, 
	password, active FROM users WHERE email = $1 LIMIT 1`, email).Scan(&u.ID,
		&u.Created, &u.Updated, &u.Name, &u.Email, &u.Role, &u.Password, &u.Active)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrBadCredential
	}
	return err
}

// Exists checks if name or email is already in database.
func (u *User) Exists(db *sql.DB) error {
	var count int64
	if err := db.QueryRow(`SELECT count(1) FROM users WHERE email=$1 OR name=$2`,
		u.Email, u.Name).Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return errors.New("Utilisateur existant")
	}
	return nil
}

// Create insert a new user into database updating time fields.
func (u *User) Create(db *sql.DB) (err error) {
	now := time.Now()
	err = db.QueryRow(`INSERT INTO users (created_at, updated_at, name, email, 
		password, role, active) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id`, now, now,
		u.Name, u.Email, u.Password, u.Role, u.Active).Scan(&u.ID)
	return err
}

// Update modifies a user into database.
func (u *User) Update(db *sql.DB) (err error) {
	now := time.Now()
	res, err := db.Exec(`UPDATE users SET updated_at=$1, name=$2, email=$3, 
	password=$4, role=$5, active=$6 WHERE id=$7 `, now, u.Name, u.Email, u.Password,
		u.Role, u.Active, u.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Utilisateur introuvable")
	}
	return err
}

// Delete removes a user from database.
func (u *User) Delete(db *sql.DB) (err error) {
	res, err := db.Exec("DELETE FROM users WHERE id = $1", u.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Utilisateur introuvable")
	}
	return nil
}
