package actions

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Iledant/iris_propera/actions/queries"
	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

type opWithPlanAction struct {
	ID                 int               `json:"id" gorm:"column:id"`
	Number             string            `json:"number" gorm:"column:number"`
	Name               string            `json:"name" gorm:"column:name"`
	Descript           models.NullString `json:"descript" gorm:"column:descript"`
	ISR                bool              `json:"isr" gorm:"column:isr"`
	Value              models.NullInt64  `json:"value" gorm:"column:value"`
	ValueDate          models.NullTime   `json:"valuedate" gorm:"column:valuedate"`
	Length             models.NullInt64  `json:"length" gorm:"column:length"`
	TRI                models.NullInt64  `json:"tri" gorm:"column:tri"`
	VAN                models.NullInt64  `json:"van" gorm:"column:van"`
	BudgetActionID     models.NullInt64  `json:"budget_action_id" gorm:"column:budget_action_id"`
	PaymentTypeID      models.NullInt64  `json:"payment_types_id" gorm:"column:payment_types_id"`
	PlanLineID         models.NullInt64  `json:"plan_line_id" gorm:"column:plan_line_id"`
	StepID             models.NullInt64  `json:"step_id" gorm:"column:step_id"`
	CategoryID         models.NullInt64  `json:"category_id" gorm:"column:category_id"`
	PlanName           models.NullString `json:"plan_name" gorm:"column:plan_name"`
	PlanID             models.NullInt64  `json:"plan_id" gorm:"column:plan_id"`
	PlanLineName       models.NullString `json:"plan_line_name" gorm:"column:plan_line_name"`
	PlanLineValue      models.NullInt64  `json:"plan_line_value" gorm:"column:plan_line_value"`
	PlanLineTotalValue models.NullInt64  `json:"plan_line_total_value" gorm:"column:plan_line_total_value"`
	BudgetActionName   models.NullString `json:"budget_action_name" gorm:"column:budget_action_name"`
	StepName           models.NullString `json:"step_name" gorm:"column:step_name"`
	CategoryName       models.NullString `json:"category_name" gorm:"column:category_name"`
}

// GetPhysicalOps handles physical operations get request.It returns all operations with plan name and action name
// for admin and observer all operations are returned, for users only operations on which the user have rights
func GetPhysicalOps(ctx iris.Context) {
	u, err := bearerToUser(ctx)

	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	qry := queries.SQLGetOpWithPlanAction

	if u.Role == models.UserRole {
		qry = strings.Replace(qry, "physical_op op", "(SELECT * FROM physical_op WHERE physical_op.id IN (SELECT physical_op_id FROM rights WHERE users_id = "+u.Subject+")) op ", -1)
	}

	db := ctx.Values().Get("db").(*gorm.DB)

	rows, err := db.Raw(qry).Rows()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}
	defer rows.Close()

	ops := struct {
		PhysicalOp []opWithPlanAction `json:"PhysicalOp"`
	}{PhysicalOp: []opWithPlanAction{}}
	var op opWithPlanAction
	for rows.Next() {
		db.ScanRows(rows, &op)
		ops.PhysicalOp = append(ops.PhysicalOp, op)
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(ops)
}

// CreatePhysicalOp handles physical operation create request.
func CreatePhysicalOp(ctx iris.Context) {
	sentOp := models.PhysicalOp{}

	if err := ctx.ReadJSON(&sentOp); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if len(sentOp.Number) != 7 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Mauvais format de numéro d'opération"})
		return
	}

	db := ctx.Values().Get("db").(*gorm.DB)
	count, err := opNumberCnt(db, sentOp.Number)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if count > 0 {
		opNumHeader, lastOpNum := sentOp.Number[0:4]+"%", struct{ Number string }{}
		if err := db.Raw("SELECT number FROM physical_op WHERE number ILIKE ? ORDER BY number DESC LIMIT 1", opNumHeader).Scan(&lastOpNum).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{err.Error()})
			return
		}
		newOpNum, err := strconv.Atoi(lastOpNum.Number[4:])
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{err.Error()})
			return
		}
		newOpNum++
		sentOp.Number = fmt.Sprintf("%s%03d", sentOp.Number[0:4], newOpNum)
	}

	if sentOp.Name == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Nom de l'opération absent"})
		return

	}

	if err := db.Create(&sentOp).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	fullOp, err := getOpWithPlanAction(db, sentOp.ID)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(struct {
		PhysicalOp opWithPlanAction `json:"PhysicalOp"`
	}{fullOp})
}

// DeletePhysicalOp handles physical operation delete request.
func DeletePhysicalOp(ctx iris.Context) {
	opID, err := ctx.Params().GetInt("opID")

	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	db := ctx.Values().Get("db").(*gorm.DB)
	op := models.PhysicalOp{}

	if err = db.First(&op, opID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(jsonError{"Opération introuvable"})
		return
	}

	if err = db.Delete(&op).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Opération supprimée"})
}

// UpdatePhysicalOp handles physical operation put request.
func UpdatePhysicalOp(ctx iris.Context) {
	opID, err := ctx.Params().GetInt("opID")

	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	db, sentOp, op := ctx.Values().Get("db").(*gorm.DB), models.PhysicalOp{}, models.PhysicalOp{}

	if err := ctx.ReadJSON(&sentOp); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if err = db.First(&op, opID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(jsonError{"Opération introuvable"})
		return
	}

	claims := ctx.Values().Get("claims").(*customClaims)
	isAdmin := (claims.Role == models.AdminRole)

	if !isAdmin {
		userID, err := strconv.Atoi(claims.Subject)
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{err.Error()})
			return
		}
		count := struct{ Count int }{}
		err = db.Raw("SELECT count(id) FROM rights WHERE users_id = ? AND physical_op_id = ?", userID, opID).Scan(&count).Error
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{err.Error()})
			return
		}
		if count.Count == 0 {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"L'utilisateur n'a pas de droits sur l'opération"})
			return
		}
	}

	if sentOp.Number != "" && isAdmin {
		o := models.PhysicalOp{}

		if err := db.Where("number = ?", sentOp.Number).First(&o).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{err.Error()})
			return
		}

		if opID != o.ID {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Numéro d'opération existant"})
			return
		}
		op.Number = sentOp.Number
	}

	if sentOp.Descript.String != "" {
		op.Descript = sentOp.Descript
	}

	if sentOp.Name != "" && isAdmin {
		op.Name = sentOp.Name
	}

	op.Isr = sentOp.Isr
	op.Value = sentOp.Value
	op.ValueDate = sentOp.ValueDate
	op.Length = sentOp.Length
	op.VAN = sentOp.VAN
	op.TRI = sentOp.TRI

	if isAdmin {
		op.BudgetActionID = sentOp.BudgetActionID
		op.CategoryID = sentOp.CategoryID
		op.PaymentTypeID = sentOp.PaymentTypeID
		op.PlanLineID = sentOp.PlanLineID
		op.StepID = sentOp.StepID
	}

	if err = db.Save(&op).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	fullOp, err := getOpWithPlanAction(db, opID)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(struct {
		PhysicalOp opWithPlanAction `json:"PhysicalOp"`
	}{fullOp})
}

// batchOp is used for decoding the array sent by the request with use of pointers to distinguish between column sent or not
type batchOp struct {
	Number        *string
	Name          *string
	Descript      *string
	Isr           *bool
	Value         *int64
	Valuedate     *time.Time
	Length        *int64
	Step          *string
	Category      *string
	Tri           *int64
	Van           *int64
	Action        *string
	PaymentTypeID *int64 `json:"payment_types_id" gorm:"column:payment_types_id"`
	PlanLineID    *int64 `json:"plan_line_id" gorm:"column:plan_line_id"`
}

func (batchOp) TableName() string {
	return "temp_physical_op"
}

// batchRequest is tne embedded struct used for decoding the request
type batchRequest struct {
	PhysicalOps []batchOp `json:"PhysicalOp"`
}

// BatchPhysicalOps handles the request sending an array of physical operations.
func BatchPhysicalOps(ctx iris.Context) {
	var req batchRequest

	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	db := ctx.Values().Get("db").(*gorm.DB)
	tx := db.Begin()

	if err := tx.Exec("DROP TABLE IF EXISTS temp_physical_op;").Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	if err := tx.Exec(queries.SQLCreateTempPhysOp).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	for _, o := range req.PhysicalOps {
		if err := tx.Create(&o).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Erreur d'insertion :" + err.Error()})
			tx.Rollback()
			return
		}
	}

	if err := tx.Exec(queries.SQLUpdateTempPhysOp).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	if err := tx.Exec(queries.SQLAddTempPhysOp).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	tx.Exec("DROP TABLE IF EXISTS temp_physical_op;")

	tx.Commit()

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Terminé"})
}

// getOpWithPlanAction return the physical operation with complementary fields
func getOpWithPlanAction(db *gorm.DB, opID int) (opWithPlanAction, error) {
	fullOp := opWithPlanAction{}
	qry := queries.SQLGetOpWithPlanAction + " WHERE op.id = ?"
	err := db.Raw(qry, opID).Scan(&fullOp).Error

	return fullOp, err
}

// opNumberCnt returns count of physical operation with the specified number.
func opNumberCnt(db *gorm.DB, number string) (int, error) {
	count := struct{ Count int }{}
	if err := db.Raw("SELECT count(id) FROM physical_op WHERE number = ?", number).Scan(&count).Error; err != nil {
		return 0, err
	}

	return count.Count, nil
}
