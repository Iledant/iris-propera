package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
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

// TableName ensures the correct table name for users.
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

// GetByID fetch a user by ID or return error using ctx to set status code and return json error code
func (u *User) GetByID(ctx iris.Context, db *gorm.DB, prefix string, ID int64) error {
	if err := db.Find(u, ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{Erreur: prefix + ", introuvable"})
			return err
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{prefix + " : " + err.Error()})
		return err
	}
	return nil
}
