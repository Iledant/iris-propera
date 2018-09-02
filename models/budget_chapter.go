package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// BudgetChapter model
type BudgetChapter struct {
	ID   int    `json:"id" gorm:"column:id"`
	Code int    `json:"code" gorm:"column:code"`
	Name string `json:"name" gorm:"column:name"`
}

// TableName ensures table name for budget_chapter
func (BudgetChapter) TableName() string {
	return "budget_chapter"
}

// GetByID fetch a physical operation by ID or return error using ctx to set status code and return json error code
func (b *BudgetChapter) GetByID(ctx iris.Context, db *gorm.DB, prefix string, ID int64) error {
	if err := db.First(b, ID).Error; err != nil {
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
