package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// PhysicalOp is the model for physical operations. Number is unique.
type PhysicalOp struct {
	ID             int64      `json:"id" gorm:"column:id"`
	Number         string     `json:"number" gorm:"column:number"`
	Name           string     `json:"name" gorm:"column:name"`
	Descript       NullString `json:"descript" gorm:"column:descript"`
	Isr            bool       `json:"isr" gorm:"column:isr"`
	Value          NullInt64  `json:"value" gorm:"column:value"`
	ValueDate      NullTime   `json:"valuedate" gorm:"column:valuedate"`
	Length         NullInt64  `json:"length" gorm:"column:length"`
	TRI            NullInt64  `json:"tri" gorm:"column:tri"`
	VAN            NullInt64  `json:"van" gorm:"column:van"`
	BudgetActionID NullInt64  `json:"budget_action_id" gorm:"column:budget_action_id"`
	PaymentTypeID  NullInt64  `json:"payment_types_id" gorm:"column:payment_types_id"`
	PlanLineID     NullInt64  `json:"plan_line_id" gorm:"column:plan_line_id"`
	StepID         NullInt64  `json:"step_id" gorm:"column:step_id"`
	CategoryID     NullInt64  `json:"category_id" gorm:"column:category_id"`
}

// TableName ensures the correct table name for physical operations.
func (PhysicalOp) TableName() string {
	return "physical_op"
}

// GetByID fetch a physical operation by ID or return error using ctx to set status code and return json error code
func (p *PhysicalOp) GetByID(ctx iris.Context, db *gorm.DB, prefix string, ID int64) error {
	if err := db.Find(p, ID).Error; err != nil {
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
