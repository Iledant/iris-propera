package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

type bccResponse struct {
	BudgetChapter []models.BudgetChapter `json:"BudgetChapter"`
}

// GetBudgetChapters handles request get all budget chapters.
func GetBudgetChapters(ctx iris.Context) {
	db := ctx.Values().Get("db").(*gorm.DB)
	bcc := []models.BudgetChapter{}

	if err := db.Find(&bcc).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(bccResponse{bcc})
}

type bcResponse struct {
	BudgetChapter models.BudgetChapter `json:"BudgetChapter"`
}

type sentBcReq struct {
	Code int
	Name string
}

// CreateBudgetChapter handles post request for creating a budget chapter.
func CreateBudgetChapter(ctx iris.Context) {
	sentBc := sentBcReq{}

	if err := ctx.ReadJSON(&sentBc); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if sentBc.Name == "" || len(sentBc.Name) > 100 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'un chapitre : mauvais format des paramètres"})
		return
	}

	bc := models.BudgetChapter{Code: sentBc.Code, Name: sentBc.Name}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := db.Create(&bc).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(bcResponse{bc})
}

// ModifyBudgetChapter handles put request for modifying a budget chapter.
func ModifyBudgetChapter(ctx iris.Context) {
	bcID, err := ctx.Params().GetInt("bcID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	sentBc := sentBcReq{}
	if err := ctx.ReadJSON(&sentBc); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	bc := models.BudgetChapter{}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := db.Find(&bc, bcID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Modification d'un chapitre: introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if sentBc.Name != "" {
		bc.Name = sentBc.Name
	}

	if sentBc.Code != 0 {
		bc.Code = sentBc.Code
	}

	if err = db.Save(&bc).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(bcResponse{bc})
}

// DeleteBudgetChapter handles delete request for a budget chapter.
func DeleteBudgetChapter(ctx iris.Context) {
	bcID, err := ctx.Params().GetInt("bcID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	bc := models.BudgetChapter{}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := db.Find(&bc, bcID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Suppression d'un chapitre: introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if err = db.Delete(&bc).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Chapitre supprimé"})
}
