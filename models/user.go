package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// User model
type User struct {
	ID            int        `json:"id"`
	Created       NullTime   `json:"created_at" gorm:"column:created_at"`
	Updated       NullTime   `json:"updated_at" gorm:"column:updated_at"`
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	Password      string     `json:"-"`
	Role          string     `json:"role"`
	RememberToken NullString `json:"-"`
	Active        bool       `json:"active"`
}

const (
	// AdminRole defines value of role row in users table for an admin
	AdminRole = "ADMIN"
	// ObserverRole defines value of role row in users table for an observer
	ObserverRole = "OBSERVER"
	// UserRole defines value of role row in users table for an usual user
	UserRole = "USER"
)

// TableName ensures table name for users
func (User) TableName() string {
	return "users"
}

// BeforeCreate is a hook to automatically set Created et Updated when creating a user
func (u *User) BeforeCreate(scope *gorm.Scope) error {
	t := NullTime{Time: time.Now(), Valid: true}
	scope.SetColumn("created_at", t)
	scope.SetColumn("updated_at", t)
	return nil
}

// BeforeUpdate is a hook to automatically set Updated when updating a user
func (u *User) BeforeUpdate(scope *gorm.Scope) error {
	t := NullTime{Time: time.Now(), Valid: true}
	scope.SetColumn("updated_at", t)
	return nil
}
