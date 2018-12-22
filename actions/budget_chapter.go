package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// GetBudgetChapters handles request get all budget chapters.
func GetBudgetChapters(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	var resp models.BudgetChapters
	if err := resp.GetAll(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des chapitres budgétaires, requête: " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

type bcResp struct {
	BudgetChapter models.BudgetChapter `json:"BudgetChapter"`
}

type sentBcReq struct {
	Code *int
	Name *string
}

// CreateBudgetChapter handles post request for creating a budget chapter.
func CreateBudgetChapter(ctx iris.Context) {
	var req models.BudgetChapter
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de chapitre budgétaire, décodage : " + err.Error()})
		return
	}
	if err := req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création de chapitre budgétaire : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := req.Create(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de chapitre budgétaire, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(bcResp{req})
}

// ModifyBudgetChapter handles put request for modifying a budget chapter.
func ModifyBudgetChapter(ctx iris.Context) {
	bcID, err := ctx.Params().GetInt64("bcID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un chapitre, paramètre : " + err.Error()})
		return
	}
	var req models.BudgetChapter
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un chapitre, décodage : " + err.Error()})
		return
	}
	if err = req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Modification d'un chapitre : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	req.ID = bcID
	if err = req.Update(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un chapitre, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(bcResp{req})
}

// DeleteBudgetChapter handles delete request for a budget chapter.
func DeleteBudgetChapter(ctx iris.Context) {
	bcID, err := ctx.Params().GetInt64("bcID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'un chapitre, paramètre : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	b := models.BudgetChapter{ID: bcID}
	if err = b.Delete(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'un chapitre, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Chapitre supprimé"})
}
