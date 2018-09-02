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
	opID, err := ctx.Params().GetInt64("opID")

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

	if op.GetByID(ctx, db, "Modification d'opération", opID) != nil {
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
func getOpWithPlanAction(db *gorm.DB, opID int64) (opWithPlanAction, error) {
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

// opPrevFc is used to decode dedicated financial commitments for previsions request.
type opPrevFc struct {
	ID          int64           `json:"id"`
	Date        time.Time       `json:"date"`
	IrisCode    string          `json:"iris_code"`
	Name        string          `json:"name"`
	Beneficiary string          `json:"beneficiary"`
	Value       int64           `json:"value"`
	LapseDate   models.NullTime `json:"lapse_date"`
	Available   int64           `json:"available"`
}

// opPrevPayment is used to decode dedicated payment for previsions request.
type opPrevPayment struct {
	Date        time.Time `json:"date"`
	Value       int64     `json:"value"`
	Beneficiary string    `json:"beneficiary"`
	IrisCode    string    `json:"iris_code"`
}

// opPrevFcPerBeneficiary is used te decode dedicated payment per beneficiary for previsions request.
type opPrevFcPerBeneficiary struct {
	Beneficiary string `json:"beneficiary"`
	Value       int64  `json:"value"`
}

// getPrevisionsResp embeddes all datas for the physical operation's previsions
type getPrevisionsResp struct {
	PrevCommitment                    []models.PrevCommitment  `json:"PrevCommitment"`
	PrevPayment                       []models.PrevPayment     `json:"PrevPayment"`
	FinancialCommitment               []opPrevFc               `json:"FinancialCommitment"`
	PendingCommitment                 models.NullInt64         `json:"PendingCommitment"`
	Payment                           []opPrevPayment          `json:"Payment"`
	PaymentPerBeneficiary             []opPrevFcPerBeneficiary `json:"PaymentPerBeneficiary"`
	FinancialCommitmentPerBeneficiary []opPrevFcPerBeneficiary `json:"FinancialCommitmentPerBeneficiary"`
	ImportLog                         []models.ImportLog       `json:"ImportLog"`
}

// GetOpPrevisions handles the get request to fetch commitments and payments prevision for a physical operation.
func GetOpPrevisions(ctx iris.Context) {
	opID, err := ctx.Params().GetInt64("opID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, erreur décodage identificateur : " + err.Error()})
		return
	}
	year, err := ctx.URLParamInt64("year")
	if err != nil {
		year = int64(time.Now().Year())
	}
	op, db := models.PhysicalOp{}, ctx.Values().Get("db").(*gorm.DB)
	if err = db.First(&op, opID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Prevision d'opération, opération introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, erreur select : " + err.Error()})
		return
	}

	resp, db := getPrevisionsResp{}, ctx.Values().Get("db").(*gorm.DB)
	if err = db.Where("year >= ?", year).Where("physical_op_id = ?", op.ID).Find(&resp.PrevCommitment).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, requête prévision engagements : " + err.Error()})
		return
	}
	if err = db.Where("year >= ?", year).Where("physical_op_id = ?", op.ID).Find(&resp.PrevPayment).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, requête prévision paiements : " + err.Error()})
		return
	}
	rows, err := db.Raw(`SELECT f.id, f.date, f.iris_code, f.name AS name, b.name AS beneficiary, f.value, 
		f.lapse_date, f.value - COALESCE(SUM(p.value - p.cancelled_value),0) AS available
		FROM financial_commitment f
		JOIN beneficiary b ON b.code = f.beneficiary_code
		LEFT JOIN payment p ON p.financial_commitment_id = f.id
		WHERE f.physical_op_id = ? GROUP BY 1,2,3,5,6,7 ORDER BY 2`, op.ID).Rows()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, requête engagements : " + err.Error()})
		return
	}
	defer rows.Close()
	fc := opPrevFc{}
	for rows.Next() {
		db.ScanRows(rows, &fc)
		resp.FinancialCommitment = append(resp.FinancialCommitment, fc)
	}
	err = db.Raw(`SELECT SUM(proposed_value) AS value FROM pending_commitments 
	WHERE physical_op_id = ? AND EXTRACT(YEAR from commission_date)=?`, op.ID, year).Scan(&resp.PendingCommitment).Error
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, requête pending : " + err.Error()})
		return
	}
	rows, err = db.Raw(`SELECT p.date, (p.value - p.cancelled_value) AS value, b.name AS beneficiary, 
	f.iris_code FROM payment p 
	JOIN financial_commitment f ON p.financial_commitment_id = f.id 
	JOIN beneficiary b ON b.code = f.beneficiary_code 
	WHERE p.financial_commitment_id IN 
	(SELECT f.id FROM financial_commitment f WHERE f.physical_op_id = ?)`, op.ID).Rows()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, requête payment : " + err.Error()})
		return
	}
	defer rows.Close()
	payment := opPrevPayment{}
	for rows.Next() {
		db.ScanRows(rows, &payment)
		resp.Payment = append(resp.Payment, payment)
	}
	rows, err = db.Raw(`SELECT b.name AS beneficiary, SUM(p.value - p.cancelled_value) AS value
	FROM payment p, financial_commitment f, beneficiary b
	WHERE p.financial_commitment_id = f.id AND b.code = f.beneficiary_code AND
	p.financial_commitment_id IN (SELECT f.id FROM financial_commitment f WHERE f.physical_op_id = ?)
	GROUP BY b.name`, op.ID).Rows()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, requête payment par bénéficiaire : " + err.Error()})
		return
	}
	defer rows.Close()
	fcPerBen := opPrevFcPerBeneficiary{}
	for rows.Next() {
		db.ScanRows(rows, &fcPerBen)
		resp.PaymentPerBeneficiary = append(resp.PaymentPerBeneficiary, fcPerBen)
	}
	rows, err = db.Raw(`SELECT b.name AS beneficiary, SUM(f.value) AS value FROM financial_commitment f  
	JOIN beneficiary b ON b.code=f.beneficiary_code WHERE f.physical_op_id = ? GROUP BY b.name`, op.ID).Rows()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, requête engagement par bénéficiaire : " + err.Error()})
		return
	}
	defer rows.Close()
	for rows.Next() {
		db.ScanRows(rows, &fcPerBen)
		resp.FinancialCommitmentPerBeneficiary = append(resp.FinancialCommitmentPerBeneficiary, fcPerBen)
	}
	if err = db.Find(&resp.ImportLog).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Prévision d'opération, requête import logs : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// setOpPc is used to decode a row of sent financial commitment for prevision.
type setOpPc struct {
	Year       int64              `json:"year"`
	Value      int64              `json:"value"`
	Descript   models.NullString  `json:"descript"`
	TotalValue models.NullInt64   `json:"total_value"`
	StateRatio models.NullFloat64 `json:"state_ratio"`
}

// setOpPayment is used to decode a row of sent payment for prevision.
type setOpPayment struct {
	Year     int64             `json:"year"`
	Value    int64             `json:"value"`
	Descript models.NullString `json:"descript"`
}

// setOpReq is used to decode sent data to the post request setting previsons of a physical operation.
type setOpReq struct {
	PrevCommitment []setOpPc      `json:"PrevCommitment"`
	PrevPayment    []setOpPayment `json:"PrevPayment"`
}

// setOpPrevResp embeddes arrays of financial commitments and payments previsions
type setOpPrevResp struct {
	PrevCommitment []models.PrevCommitment `json:"PrevCommitment"`
	PrevPayment    []models.PrevPayment    `json:"PrevPayment"`
}

// SetOpPrevisions handles the post request to set financial commitments and payments previsions
func SetOpPrevisions(ctx iris.Context) {
	opID, err := ctx.Params().GetInt64("opID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation prévision d'opération, erreur décodage identificateur : " + err.Error()})
		return
	}
	req := setOpReq{}
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation prévision d'opération, erreur décodage payload : " + err.Error()})
		return
	}
	op, db := models.PhysicalOp{}, ctx.Values().Get("db").(*gorm.DB)
	if err = db.First(&op, opID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Fixation prévision d'opération, opération introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation prévision d'opération, erreur select : " + err.Error()})
		return
	}

	tx := db.Begin()
	if err = tx.Exec("delete from prev_commitment where physical_op_id = ?", opID).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation prévision d'opération, erreur delete : " + err.Error()})
		tx.Rollback()
		return
	}
	if err = tx.Exec("delete from prev_payment where physical_op_id = ?", opID).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation prévision d'opération, erreur delete : " + err.Error()})
		tx.Rollback()
		return
	}
	for _, pc := range req.PrevCommitment {
		if err = tx.Exec("insert into prev_commitment (year, value, descript, total_value, state_ratio, physical_op_id) values (?, ?, ?, ?, ?, ?)",
			pc.Year, pc.Value, pc.Descript, pc.TotalValue, pc.StateRatio, opID).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Fixation prévision d'opération, erreur insert prev_commitment : " + err.Error()})
			tx.Rollback()
			return
		}
	}
	for _, p := range req.PrevPayment {
		if err = tx.Exec("insert into prev_payment (year, value, descript, physical_op_id) values (?, ?, ?, ?)",
			p.Year, p.Value, p.Descript, opID).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Fixation prévision d'opération, erreur insert prev_payment : " + err.Error()})
			tx.Rollback()
			return
		}
	}
	tx.Commit()

	resp := setOpPrevResp{}
	if err = db.Where("physical_op_id = ?", opID).Find(&resp.PrevCommitment).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation prévision d'opération, requête get prévision engagements : " + err.Error()})
		return
	}
	if err = db.Where("physical_op_id = ?", opID).Find(&resp.PrevPayment).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation prévision d'opération, requête get prévision paiements : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
