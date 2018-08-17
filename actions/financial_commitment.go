package actions

import (
	"errors"
	"net/http"
	"time"

	"github.com/Iledant/iris_propera/actions/queries"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// reqFcIds is used to decode sent ids for attaching financial commitment.
type reqFcIDs struct {
	IDs []int64 `json:"fcIdList"`
}

// unlinkFc is used to get financial commitment not linked to a physical operation or plan line
type unlinkFc struct {
	ID          int       `json:"id" gorm:"column:id"`
	Value       int64     `json:"value" gorm:"column:value"`
	IrisCode    string    `json:"iris_code" gorm:"column:iris_code"`
	Name        string    `json:"name" gorm:"column:name"`
	Date        time.Time `json:"date" gorm:"column:date"`
	Beneficiary string    `json:"beneficiary" gorm:"column:beneficiary"`
}

type unlinkFcs struct {
	FinancialCommitments []unlinkFc `json:"FinancialCommitment"`
}

// unlinkFcs embeddes an array of unlinkFc and informations about pagination of the query.
type pagUnlinkFcs struct {
	Data        unlinkFcs `json:"data"`
	CurrentPage int64     `json:"current_page"`
	LastPage    int64     `json:"last_page"`
}

// monthFC is used to get financial amount of a month
type monthFC struct {
	Month int64 `json:"month"`
	Value int64 `json:"value"`
}

// monthFCC embeddes an array of monthFC for the JSON response.
type monthFCC struct {
	FinancialCommitmentPerMonth []monthFC `json:"FinancialCommitmentPerMonth"`
}

// getPageOffset returns the correct offset and page according to total number of rows.
func getPageOffset(page int64, count int64) (offset int64, newPage int64, lastPage int64) {
	if count == 0 {
		return 0, 0, 1
	}

	offset = (page - 1) * 15
	newPage = 1

	if offset < 0 {
		offset = 0
	}

	if offset >= count {
		offset = (count - 1) - ((count - 1) % 15)
		newPage = offset/15 + 1
	}

	lastPage = (count-1)/15 + 1

	return offset, newPage, lastPage
}

// getOpUnlinkedFcs handles get unlinked financial commitments to physical operations
func getUnlinkedFcs(ctx iris.Context, lType string, search string, minDate time.Time, page int64) {
	db, countQry, selQry, cCount := ctx.Values().Get("db").(*gorm.DB), "", "", struct{ Count int64 }{}

	if lType == "PhysicalOp" {
		countQry = queries.SQLCountOpUnlinkedFcs
		selQry = queries.SQLGetOpUnlinkedFcs
	} else {
		countQry = queries.SQLCountPlUnlinkedFcs
		selQry = queries.SQLGetPlUnlinkedFcs
	}

	tx := db.Begin()

	if err := tx.Raw(countQry, minDate, search, search, search).Scan(&cCount).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	offset, page, lastPage := getPageOffset(page, cCount.Count)

	rows, err := tx.Raw(selQry, minDate, search, search, search, offset).Rows()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}
	defer rows.Close()

	ii, i := unlinkFcs{}, unlinkFc{}
	for rows.Next() {
		db.ScanRows(rows, &i)
		ii.FinancialCommitments = append(ii.FinancialCommitments, i)
	}
	tx.Commit()

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(pagUnlinkFcs{CurrentPage: page, LastPage: lastPage, Data: ii})
}

// parseParams fetch params for linked or unlinked requests.
func parseParams(ctx iris.Context) (page int64, search string, minDate time.Time, lType string, err error) {
	page, err = ctx.URLParamInt64("page")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		return 0, "", time.Time{}, "", errors.New("erreur page:" + err.Error())
	}

	search, lType = ctx.URLParam("search"), ctx.URLParam("LinkType")

	minYear, err := ctx.URLParamInt("MinYear")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		return 0, "", time.Time{}, "", errors.New("erreur sur MinYear :" + err.Error())
	}

	if lType != "PhysicalOp" && lType != "PlanLine" {
		ctx.StatusCode(http.StatusBadRequest)
		return 0, "", time.Time{}, "", errors.New("mauvais paramètre LinkType")
	}

	search, minDate = "%"+search+"%", time.Date(minYear, 1, 1, 0, 0, 0, 0, time.UTC)

	return page, search, minDate, lType, nil
}

// GetUnlinkedFcs handles the request to get all financial commitments not linked to a physical operation or a plan line.
// It uses a Laravel paginated request and has parameters for searching
func GetUnlinkedFcs(ctx iris.Context) {
	page, search, minDate, lType, err := parseParams(ctx)
	if err != nil {
		ctx.JSON(jsonError{"Engagements non liés : " + err.Error()})
		return
	}

	getUnlinkedFcs(ctx, lType, search, minDate, page)
}

//GetMonthFC handles the request to get the amount of financial commitments each montant of a given year.
func GetMonthFC(ctx iris.Context) {
	year, err := ctx.URLParamInt("year")
	if err != nil || year == 0 {
		year = time.Now().Year()
	}

	db := ctx.Values().Get("db").(*gorm.DB)

	rows, err := db.Raw(queries.SQLMonthFCs, year).Rows()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}
	defer rows.Close()

	mm, m := monthFCC{}, monthFC{}
	for rows.Next() {
		db.ScanRows(rows, &m)
		mm.FinancialCommitmentPerMonth = append(mm.FinancialCommitmentPerMonth, m)
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(mm)
}

// validate check if all financial commitment IDs exist.
func (fcIDs reqFcIDs) validate(ctx iris.Context, db *gorm.DB, errPrefix string) bool {
	if len(fcIDs.IDs) > 0 {
		var count struct{ Count int }
		err := db.Raw("SELECT count(id) FROM financial_commitment WHERE id IN (?)", fcIDs.IDs).Scan(&count).Error

		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{err.Error()})
			return false
		}

		if count.Count != len(fcIDs.IDs) {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{errPrefix + "mauvais identificateur d'engagement"})
			return false
		}
	}

	return true
}

// getFcIds fetch an array of financial commitments and check if all exists and set context accordingly
func getFcIds(ctx iris.Context, db *gorm.DB, errPrefix string) (reqFcIDs, bool) {
	fcIDs := reqFcIDs{}

	if err := ctx.ReadJSON(&fcIDs); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{errPrefix + err.Error()})
		return fcIDs, false
	}

	return fcIDs, fcIDs.validate(ctx, db, errPrefix)
}

// LinkFcToOp handles the request to link an array of financial commitments to an physical operation.
func LinkFcToOp(ctx iris.Context) {
	opID, err := ctx.Params().GetInt("opID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	op, db := models.PhysicalOp{}, ctx.Values().Get("db").(*gorm.DB)
	if err = db.First(&op, opID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Rattachement d'engagement : opération introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
	}

	fcIDs, ok := getFcIds(ctx, db, "Rattachement d'engagement : ")
	if !ok {
		return
	}
	tx := db.Begin()

	for _, id := range fcIDs.IDs {
		if err := tx.Exec("UPDATE financial_commitment SET physical_op_id = ? WHERE id = ?", opID, id).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonMessage{err.Error()})
			tx.Rollback()
			return
		}
	}

	tx.Commit()

	getUnlinkedFcs(ctx, "PhysicalOp", "%", time.Time{}, 1)
}

// LinkFcToPl handles the request to link an array of financial commitments to a plan line.
func LinkFcToPl(ctx iris.Context) {
	plID, err := ctx.Params().GetInt("plID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	pl, db := models.PlanLine{}, ctx.Values().Get("db").(*gorm.DB)
	if err = db.First(&pl, plID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Rattachement d'engagement : ligne de plan introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
	}

	fcIDs, ok := getFcIds(ctx, db, "Rattachement d'engagement : ")
	if !ok {
		return
	}

	tx := db.Begin()

	for _, id := range fcIDs.IDs {
		if err := tx.Exec("UPDATE financial_commitment SET plan_line_id = ? WHERE id = ?", plID, id).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonMessage{err.Error()})
			tx.Rollback()
			return
		}
	}

	tx.Commit()

	getUnlinkedFcs(ctx, "PlanLine", "%", time.Time{}, 1)
}

// opLinkFc is used to query financial commitment linked to a physical operation.
type opLinkFc struct {
	FcID          int       `json:"fcID" gorm:"column:fc_id"`
	FcValue       int64     `json:"fcValue" gorm:"column:fc_value"`
	FcName        string    `json:"fcName" gorm:"column:fc_name"`
	IrisCode      string    `json:"iris_code" gorm:"column:iris_code"`
	FcDate        time.Time `json:"fcDate" gorm:"column:fc_date"`
	OpNumber      string    `json:"opNumber" gorm:"column:op_number"`
	OpName        string    `json:"opName" gorm:"column:op_name"`
	FcBeneficiary string    `json:"fcBeneficiary" gorm:"column:fc_beneficiary"`
}

// opLinkFcs embeddes an array of opLinkFc.
type opLinkFcs struct {
	FinancialCommitments []opLinkFc `json:"FinancialCommitment"`
}

// unlinkFcs embeddes opLinkFcs and informations about pagination of the query.
type pagOpLinkFcs struct {
	Data        opLinkFcs `json:"data"`
	CurrentPage int64     `json:"current_page"`
	LastPage    int64     `json:"last_page"`
}

// getOpLinkedFcs return the financial commitments linked to an operation matching with search pattern
// Use laravel like pagination
func getOpLinkedFcs(ctx iris.Context, search string, minDate time.Time, page int64) {
	db, cCount := ctx.Values().Get("db").(*gorm.DB), struct{ Count int64 }{}

	tx := db.Begin()

	if err := tx.Raw(queries.SQLCountOpLinkedFcs, minDate, search, search, search, search).Scan(&cCount).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	offset, page, lastPage := getPageOffset(page, cCount.Count)

	rows, err := tx.Raw(queries.SQLGetOpLinkedFcs, minDate, search, search, search, search, offset).Rows()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}
	defer rows.Close()

	ii, i := opLinkFcs{}, opLinkFc{}
	for rows.Next() {
		db.ScanRows(rows, &i)
		ii.FinancialCommitments = append(ii.FinancialCommitments, i)
	}
	tx.Commit()

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(pagOpLinkFcs{CurrentPage: page, LastPage: lastPage, Data: ii})
}

// plLinkFc is used to query financial commitment linked to a plan line.
type plLinkFc struct {
	FcID          int       `json:"fcID" gorm:"column:fc_id"`
	FcValue       int64     `json:"fcValue" gorm:"column:fc_value"`
	FcName        string    `json:"fcName" gorm:"column:fc_name"`
	IrisCode      string    `json:"iris_code" gorm:"column:iris_code"`
	FcDate        time.Time `json:"fcDate" gorm:"column:fc_date"`
	PlName        string    `json:"plName" gorm:"column:pl_name"`
	FcBeneficiary string    `json:"fcBeneficiary" gorm:"column:fc_beneficiary"`
}

// plLinkFcs embeddes an array of plLinkFc.
type plLinkFcs struct {
	FinancialCommitments []plLinkFc `json:"FinancialCommitment"`
}

// unlinkFcs embeddes a plLinkFcs and informations about pagination of the query.
type pagPlLinkFcs struct {
	Data        plLinkFcs `json:"data"`
	CurrentPage int64     `json:"current_page"`
	LastPage    int64     `json:"last_page"`
}

// getPlLinkedFcs return the financial commitments linked to an operation matching with search pattern
// Use laravel like pagination
func getPlLinkedFcs(ctx iris.Context, search string, minDate time.Time, page int64) {
	db, cCount := ctx.Values().Get("db").(*gorm.DB), struct{ Count int64 }{}

	tx := db.Begin()

	if err := tx.Raw(queries.SQLCountPlLinkedFcs, minDate, search, search, search).Scan(&cCount).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	offset, page, lastPage := getPageOffset(page, cCount.Count)

	rows, err := tx.Raw(queries.SQLGetPlLinkedFcs, minDate, search, search, search, offset).Rows()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}
	defer rows.Close()

	ii, i := plLinkFcs{}, plLinkFc{}
	for rows.Next() {
		db.ScanRows(rows, &i)
		ii.FinancialCommitments = append(ii.FinancialCommitments, i)
	}
	tx.Commit()

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(pagPlLinkFcs{CurrentPage: page, LastPage: lastPage, Data: ii})
}

// GetLinkedFcs handles the request to get all financial commitments linked to a physical operation or a plan line.
// It uses a Laravel paginated request and has parameters for searching
func GetLinkedFcs(ctx iris.Context) {
	page, search, minDate, lType, err := parseParams(ctx)
	if err != nil {
		ctx.JSON(jsonError{"Engagements non liés : " + err.Error()})
		return
	}

	if lType == "PhysicalOp" {
		getOpLinkedFcs(ctx, search, minDate, page)
	} else {
		getPlLinkedFcs(ctx, search, minDate, page)
	}
}

// getFcsOpReq embedded the list of financial commits linked to a physical operation.
type getFcsOpReq struct {
	Fcs []models.FinancialCommitment `json:"FinancialCommitment"`
}

// GetOpFcs handles the request to get all financial commitments linked to a physical operation.
func GetOpFcs(ctx iris.Context) {
	opID, err := ctx.Params().GetInt("opID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	op, db := models.PhysicalOp{}, ctx.Values().Get("db").(*gorm.DB)
	if err = db.First(&op, opID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Liste des engagements : opération introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
	}

	rows, err := db.Raw("SELECT * from financial_commitment WHERE physical_op_id = ?", opID).Rows()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}
	defer rows.Close()

	ii, i := getFcsOpReq{}, models.FinancialCommitment{}
	for rows.Next() {
		db.ScanRows(rows, &i)
		ii.Fcs = append(ii.Fcs, i)
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(ii)
}

type unlinkFcsReq struct {
	LinkType string `json:"linkType"`
	reqFcIDs
}

// UnlinkFcs handles the requests to unset link between a financial commitment and a physical operation or plan line.
func UnlinkFcs(ctx iris.Context) {
	req, db := unlinkFcsReq{}, ctx.Values().Get("db").(*gorm.DB)
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if !req.reqFcIDs.validate(ctx, db, "Détachement d'engagement : ") {
		return
	}

	tx, qry := db.Begin(), ""

	if req.LinkType == "PhysicalOp" {
		qry = `UPDATE financial_commitment SET physical_op_id = NULL where id = ?`
	} else {
		qry = `UPDATE financial_commitment SET plan_line_id = NULL where id = ?`
	}

	for _, id := range req.IDs {
		if err := tx.Exec(qry, id).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{err.Error()})
			tx.Rollback()
			return
		}
	}
	tx.Commit()

	if req.LinkType == "PhysicalOp" {
		getOpLinkedFcs(ctx, "%", time.Time{}, 1)
	} else {
		getPlLinkedFcs(ctx, "%", time.Time{}, 1)
	}
}

//sentFc is used to decode uploaded financial commitments
type sentFc struct {
	Chapter         string          `json:"chapter" gorm:"column:chapter"`
	Action          string          `json:"action" gorm:"column:action"`
	IrisCode        string          `json:"iris_code" gorm:"column:iris_code"`
	CoriolisYear    string          `json:"coriolis_year" gorm:"column:coriolis_year"`
	CoriolisEgtCode string          `json:"coriolis_egt_code" gorm:"column:coriolis_egt_code"`
	CoriolisEgtNum  string          `json:"coriolis_egt_num" gorm:"column:coriolis_egt_num"`
	CoriolisEgtLine string          `json:"coriolis_egt_line" gorm:"column:coriolis_egt_line"`
	Name            string          `json:"name" gorm:"column:name"`
	Beneficiary     string          `json:"beneficiary" gorm:"column:beneficiary"`
	BeneficiaryCode int             `json:"beneficiary_code" gorm:"column:beneficiary_code"`
	Date            time.Time       `json:"date" gorm:"column:date"`
	Value           int64           `json:"value" gorm:"column:value"`
	LapseDate       models.NullTime `json:"lapse_date" gorm:"column:lapse_date"`
}

func (sentFc) TableName() string {
	return "temp_commitment"
}

// sentFcc embeddes a array of sentFcs
type sentFcc struct {
	Fcs []sentFc `json:"FinancialCommitment"`
}

// BatchFcs handles the post request with an array of financial commitments (IRIS import).
func BatchFcs(ctx iris.Context) {
	db, fcs := ctx.Values().Get("db").(*gorm.DB), sentFcc{}

	if err := ctx.ReadJSON(&fcs); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	tx := db.Begin()
	if err := tx.Exec("DELETE from temp_commitment").Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	for _, fc := range fcs.Fcs {
		if err := tx.Exec(queries.SQLInsertTempCommitment, fc.Chapter, fc.Action, fc.IrisCode, fc.CoriolisYear,
			fc.CoriolisEgtCode, fc.CoriolisEgtNum, fc.CoriolisEgtLine, fc.Name, fc.Beneficiary, fc.BeneficiaryCode,
			fc.Date, 100*fc.Value, fc.LapseDate).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{err.Error()})
			tx.Rollback()
			return
		}
	}

	queries := []string{queries.SQLUpdateUploadFcs, queries.SQLInsertUploadFcs,
		queries.SQLInsertNewBeneficiary, queries.SQLZeroDuplicatedFcs,
		queries.SQLUpdateFcActionField}
	for _, qry := range queries {
		if err := tx.Exec(qry).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{err.Error()})
			tx.Rollback()
			return
		}
	}

	if err := tx.Exec("DELETE from temp_commitment").Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	if err := tx.Exec("UPDATE import_logs SET last_date = ? WHERE category = 'FinancialCommitments'", time.Now()).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}
	tx.Commit()
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Engagements importés et mis à jour"})
}

// sentOpFc is used to decode uploaded link between physical operation and financial commitment.
type sentOpFc struct {
	OpNumber        string `json:"op_number" gorm:"column:op_number"`
	CoriolisYear    string `json:"coriolis_year" gorm:"column:coriolis_year"`
	CoriolisEgtCode string `json:"coriolis_egt_code" gorm:"column:coriolis_egt_code"`
	CoriolisEgtNum  string `json:"coriolis_egt_num" gorm:"column:coriolis_egt_num"`
	CoriolisEgtLine string `json:"coriolis_egt_line" gorm:"column:coriolis_egt_line"`
}

func (sentOpFc) TableName() string {
	return "temp_attachment"
}

// sentOpFcc embeddes an array of sentOpFc
type sentOpFcc struct {
	OpFcs []sentOpFc `json:"Attachment"`
}

// BatchOpFcs handle the post request to link of an array of physical operations with financial commitments.
func BatchOpFcs(ctx iris.Context) {
	db, opFcs := ctx.Values().Get("db").(*gorm.DB), sentOpFcc{}

	if err := ctx.ReadJSON(&opFcs); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	tx := db.Begin()
	if err := tx.Exec("DELETE from temp_attachment").Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	for _, opFc := range opFcs.OpFcs {
		if err := tx.Create(&opFc).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{err.Error()})
			tx.Rollback()
			return
		}
	}

	if err := tx.Exec(queries.SQLBatchOpFc).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	if err := tx.Exec("DELETE from temp_attachment").Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}
	tx.Commit()
	ctx.StatusCode(http.StatusOK)
	ctx.JSON("Rattachements importés et réalisés")
}
