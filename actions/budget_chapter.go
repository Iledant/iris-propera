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
	db, bcc := ctx.Values().Get("db").(*gorm.DB), bccResponse{}

	if err := db.Find(&bcc.BudgetChapter).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(bcc)
}

type bcResponse struct {
	BudgetChapter models.BudgetChapter `json:"BudgetChapter"`
}

type sentBcReq struct {
	Code *int
	Name *string
}

// CreateBudgetChapter handles post request for creating a budget chapter.
func CreateBudgetChapter(ctx iris.Context) {
	req := sentBcReq{}

	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if req.Name == nil || *req.Name == "" || len(*req.Name) > 100 || req.Code == nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'un chapitre : mauvais format des paramètres"})
		return
	}

	bc := models.BudgetChapter{Code: *req.Code, Name: *req.Name}
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
	bcID, err := ctx.Params().GetInt64("bcID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un chapitre, décodage : " + err.Error()})
		return
	}

	req := sentBcReq{}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	bc, db := models.BudgetChapter{}, ctx.Values().Get("db").(*gorm.DB)
	if bc.GetByID(ctx, db, "Modification d'un chapitre", bcID) != nil {
		return
	}

	if req.Name != nil && *req.Name != "" && len(*req.Name) < 100 {
		bc.Name = *req.Name
	}
	if req.Code != nil {
		bc.Code = *req.Code
	}

	if err = db.Save(&bc).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un chapitre, save : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(bcResponse{bc})
}

// DeleteBudgetChapter handles delete request for a budget chapter.
func DeleteBudgetChapter(ctx iris.Context) {
	bcID, err := ctx.Params().GetInt64("bcID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'un chapitre, décodage : " + err.Error()})
		return
	}

	bc, db := models.BudgetChapter{}, ctx.Values().Get("db").(*gorm.DB)
	if bc.GetByID(ctx, db, "Suppression d'un chapitre", bcID) != nil {
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
