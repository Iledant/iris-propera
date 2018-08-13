package actions

import (
	"net/http"
	"time"

	"github.com/Iledant/iris_propera/actions/queries"
	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

type brrResponse struct {
	BudgetCredit []models.BudgetCredit `json:"BudgetCredits"`
}

type brResponse struct {
	BudgetCredit models.BudgetCredit `json:"BudgetCredits"`
}

// GetBudgetCredits handles request get all budget credits.
func GetBudgetCredits(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	brr := []models.BudgetCredit{}

	if err := db.Find(&brr).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(brrResponse{brr})
}

// GetLastBudgetCredits handles request for getting the most recent budget credits of current year.
func GetLastBudgetCredits(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	year := time.Now().Year()

	rows, err := db.Raw(queries.SQLGetMostRecentCredits, year).Rows()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}
	defer rows.Close()
	brr, br := []models.BudgetCredit{}, models.BudgetCredit{}

	for rows.Next() {
		db.ScanRows(rows, &br)
		brr = append(brr, br)
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(brrResponse{brr})
}

type sentBrReq struct {
	CommissionDate     models.NullTime  `json:"commission_date"`
	ChapterID          models.NullInt64 `json:"chapter_id"`
	PrimaryCommitment  *int64           `json:"primary_commitment"`
	FrozenCommitment   *int64           `json:"frozen_commitment"`
	ReservedCommitment *int64           `json:"reserved_commitment"`
}

// CreateBudgetCredit handles post request for creating a budget credit.
func CreateBudgetCredit(ctx iris.Context) {
	sbr := sentBrReq{}

	if err := ctx.ReadJSON(&sbr); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if sbr.CommissionDate.Valid == false || sbr.ChapterID.Valid == false || sbr.PrimaryCommitment == nil ||
		sbr.FrozenCommitment == nil || sbr.ReservedCommitment == nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création de crédits : champ manquant ou incorrect"})
		return
	}

	br := models.BudgetCredit{CommissionDate: sbr.CommissionDate, ChapterID: sbr.ChapterID,
		PrimaryCommitment: *sbr.PrimaryCommitment, FrozenCommitment: *sbr.FrozenCommitment,
		ReservedCommitment: *sbr.ReservedCommitment}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := db.Create(&br).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(brResponse{br})
}

// ModifyBudgetCredit handles put request for modifying budget credits.
func ModifyBudgetCredit(ctx iris.Context) {
	brID, err := ctx.Params().GetInt("brID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	sbr := sentBrReq{}
	if err := ctx.ReadJSON(&sbr); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	br := models.BudgetCredit{}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := db.Find(&br, brID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Modification des crédits: introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if sbr.ChapterID.Valid == true {
		br.ChapterID = sbr.ChapterID
	}

	if sbr.CommissionDate.Valid == true {
		br.CommissionDate = sbr.CommissionDate
	}

	if sbr.PrimaryCommitment != nil {
		br.PrimaryCommitment = *sbr.PrimaryCommitment
	}

	if sbr.ReservedCommitment != nil {
		br.ReservedCommitment = *sbr.ReservedCommitment
	}

	if sbr.FrozenCommitment != nil {
		br.FrozenCommitment = *sbr.FrozenCommitment
	}

	if err = db.Save(&br).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(brResponse{br})
}

// DeleteBudgetCredit handles delete request for budget credits.
func DeleteBudgetCredit(ctx iris.Context) {
	brID, err := ctx.Params().GetInt("brID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	br := models.BudgetCredit{}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := db.Find(&br, brID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Suppression de crédits: introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if err = db.Delete(&br).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Crédits supprimés"})
}

// batchBr is used for batch import of budget credits
type batchBr struct {
	CommissionDate     *time.Time `json:"commission_date"`
	Chapter            *int       `json:"chapter"`
	PrimaryCommitment  *int64     `json:"primary_commitment"`
	ReservedCommitment *int64     `json:"reserved_commitment"`
	FrozenCommitment   *int64     `json:"frozen_commitment"`
}

// batchBrr is used to embed batch credits imports
type batchBrr struct {
	BudgetCredits []batchBr `json:"BudgetCredits"`
}

// BatchBudgetCredits handles the post array request for budget credits
func BatchBudgetCredits(ctx iris.Context) {
	var brr batchBrr

	if err := ctx.ReadJSON(&brr); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Erreur de lecture du batch crédits : " + err.Error()})
		return
	}

	db := ctx.Values().Get("db").(*gorm.DB)
	tx := db.Begin()

	if err := tx.Exec(queries.SQLDropTempCreditsTable).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch crédits erreur de suppression de la table temporaire : " + err.Error()})
		tx.Rollback()
		return
	}

	if err := tx.Exec(queries.SQLCreateTempCreditsTable).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch crédits erreur de création de la table temporaire : " + err.Error()})
		tx.Rollback()
		return
	}

	for _, br := range brr.BudgetCredits {
		if br.CommissionDate == nil || br.Chapter == nil || br.PrimaryCommitment == nil ||
			br.ReservedCommitment == nil || br.FrozenCommitment == nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Batch crédits, champs manquants"})
			tx.Rollback()
			return
		}

		if err := tx.Exec(queries.SQLInsertTempCredits, *br.CommissionDate, *br.Chapter,
			*br.PrimaryCommitment, *br.ReservedCommitment, *br.FrozenCommitment).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Batch crédits erreur d'import : " + err.Error()})
			tx.Rollback()
			return
		}
	}

	if err := tx.Exec(queries.SQLUpdateBatchCredits).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch crédits erreur d'insertion : " + err.Error()})
		tx.Rollback()
		return
	}

	if err := tx.Exec(queries.SQLDropTempCreditsTable).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch crédits erreur de suppression de la table temporaire : " + err.Error()})
		tx.Rollback()
		return
	}

	tx.Commit()

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Credits importés"})
}
