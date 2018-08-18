package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/actions/queries"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// prrResp embeddes an array of paymentratio
type prrResp struct {
	Prr []models.PaymentRatio `json:"PaymentRatio"`
}

// GetRatios handles the get request to fetch all payment ratios.
func GetRatios(ctx iris.Context) {
	db, prr := ctx.Values().Get("db").(*gorm.DB), prrResp{}

	if err := db.Find(&prr.Prr).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(prr)
}

// GetPtRatios handles the get request to fetch ratios linked to a payment type.
func GetPtRatios(ctx iris.Context) {
	db, pt := ctx.Values().Get("db").(*gorm.DB), models.PaymentType{}

	ptID, err := ctx.Params().GetInt("ptID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des ratios d'une chronique, erreur de paramètre : " + err.Error()})
		return
	}

	if err = db.First(&pt, ptID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Liste des ratios : chronique introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	prr := prrResp{}
	if err = db.Where("payment_types_id = ?", ptID).Find(&prr.Prr).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(prr)
}

// DeleteRatios handles the delete request for a payment ratio.
func DeleteRatios(ctx iris.Context) {
	ptID, err := ctx.Params().GetInt("ptID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}
	db, pt := ctx.Values().Get("db").(*gorm.DB), models.PaymentType{}

	if err := db.First(&pt, ptID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Suppression de ratios : introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression de ratios : " + err.Error()})
		return
	}

	if err := db.Exec(queries.SQLDeleteRatios, ptID).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression de ratios : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Ratios supprimés"})
}

// sentPrReq is used to decode a sent payment ratio.
type sentPrReq struct {
	Ratio float64 `json:"ratio"`
	Index int64   `json:"index"`
}

// sentPrrReq is used to decode the payment ratios payload.
type sentPrrReq struct {
	Prr []sentPrReq `json:"PaymentRatio"`
}

// SetPtRatios handle the post request for setting all ratios of an payment type.
func SetPtRatios(ctx iris.Context) {
	db, pt := ctx.Values().Get("db").(*gorm.DB), models.PaymentType{}

	ptID, err := ctx.Params().GetInt("ptID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Ratios d'une chronique, erreur de paramètre : " + err.Error()})
		return
	}

	if err = db.First(&pt, ptID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Ratios : chronique introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	prr := sentPrrReq{}
	if err := ctx.ReadJSON(&prr); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	tx := db.Begin()

	if err := tx.Exec(queries.SQLDeleteRatios, ptID).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	qry := `INSERT into payment_ratios (payment_types_id, ratio, index) VALUES (?,?,?)`
	for _, pr := range prr.Prr {
		if err := tx.Exec(qry, ptID, pr.Ratio, pr.Index).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{err.Error()})
			tx.Rollback()
			return
		}
	}

	resp := prrResp{}
	if err = tx.Where("payment_types_id = ?", ptID).Find(&resp.Prr).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}
	tx.Commit()

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// yearRatio is used to scan and encode an year ratio
type yearRatio struct {
	Index int64   `json:"index"`
	Ratio float64 `json:"ratio"`
}

// yearRatios embeddes an array of yearRatio for the response
type yearRatios struct {
	Ratios []yearRatio `json:"Ratios"`
}

// GetYearRatios handles the get request to fetch the payment ratios linked to the financial commitments of a given year.
func GetYearRatios(ctx iris.Context) {
	year := ctx.URLParam("Year")

	if year == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Ratios annuels : année manquante"})
		return
	}

	db, resp, yr := ctx.Values().Get("db").(*gorm.DB), yearRatios{}, yearRatio{}

	rows, err := db.Raw(queries.SQLGetYearRatio, year, year).Rows()

	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		db.ScanRows(rows, &yr)
		resp.Ratios = append(resp.Ratios, yr)
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
