package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// bppResp embeddes response for an array of budget programs.
type bppResp struct {
	BudgetProgram []models.BudgetProgram `json:"BudgetProgram"`
}

// bppResp embeddes response for an single budget program.
type bpResp struct {
	BudgetProgram models.BudgetProgram `json:"BudgetProgram"`
}

// bpReq is used for creation and modification of a budget program.
type bpReq struct {
	CodeContract    *string `json:"code_contract"`
	CodeFunction    *string `json:"code_function"`
	CodeNumber      *string `json:"code_number"`
	CodeSubfunction *string `json:"code_subfunction"`
	Name            *string `json:"name"`
}

// GetChapterBudgetPrograms handles request get budget programs of a chapter.
func GetChapterBudgetPrograms(ctx iris.Context) {
	chpID, err := ctx.Params().GetInt("chpID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	db := ctx.Values().Get("db").(*gorm.DB)
	rows, err := db.Raw("SELECT * FROM budget_program WHERE chapter_id = ?", chpID).Rows()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}
	defer rows.Close()

	arr, item := bppResp{}, models.BudgetProgram{}
	for rows.Next() {
		rows.Scan(&item)
		arr.BudgetProgram = append(arr.BudgetProgram, item)
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(arr)
}

// GetAllBudgetPrograms handles request get all budget programs.
func GetAllBudgetPrograms(ctx iris.Context) {
	bpp := []models.BudgetProgram{}

	db := ctx.Values().Get("db").(*gorm.DB)
	if err := db.Find(&bpp).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(bppResp{bpp})
}

// CreateBudgetProgram handles request post request to create a new program.
func CreateBudgetProgram(ctx iris.Context) {
	chpID, err := ctx.Params().GetInt("chpID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	bp := bpReq{}
	if err = ctx.ReadJSON(&bp); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if bp.CodeContract == nil || len(*bp.CodeContract) != 1 || bp.CodeFunction == nil ||
		*bp.CodeFunction == "" || len(*bp.CodeFunction) > 2 || bp.CodeNumber == nil || *bp.CodeNumber == "" ||
		len(*bp.CodeNumber) > 3 || (bp.CodeSubfunction != nil && len(*bp.CodeSubfunction) != 1) {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création de programme budgétaire, champ manquant ou incorrect"})
		return
	}
	db, chp := ctx.Values().Get("db").(*gorm.DB), models.BudgetChapter{}

	if err = db.First(&chp, chpID).Error; err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création de programme budgétaire, index chapitre incorrect"})
		return
	}

	newBp := models.BudgetProgram{CodeContract: *bp.CodeContract, CodeFunction: *bp.CodeFunction, CodeNumber: *bp.CodeNumber, ChapterID: chpID}
	if bp.CodeSubfunction == nil {
		newBp.CodeSubfunction.Valid = false
	} else {
		newBp.CodeSubfunction.Valid = true
		newBp.CodeSubfunction.String = *bp.CodeSubfunction
	}

	if err = db.Create(&newBp).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(bpResp{newBp})
}

// ModifyBudgetProgram handles request put requestion to modify a program.
func ModifyBudgetProgram(ctx iris.Context) {
	bpID, err := ctx.Params().GetInt("bpID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	bp, db := models.BudgetProgram{}, ctx.Values().Get("db").(*gorm.DB)

	if err = db.First(&bp, bpID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonMessage{"Modification de programme : introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	req := bpReq{}
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if req.CodeContract != nil && len(*req.CodeContract) == 1 {
		bp.CodeContract = *req.CodeContract
	}

	if req.CodeFunction != nil && len(*req.CodeFunction) < 3 {
		bp.CodeFunction = *req.CodeFunction
	}

	if req.CodeNumber != nil && len(*req.CodeNumber) < 4 {
		bp.CodeNumber = *req.CodeNumber
	}

	if req.CodeSubfunction != nil && len(*req.CodeSubfunction) < 2 {
		bp.CodeSubfunction.Valid = true
		bp.CodeSubfunction.String = *req.CodeSubfunction
	}

	if err = db.Save(&bp).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(bpResp{bp})
}

// DeleteBudgetProgram handles the request to delete an budget program.
func DeleteBudgetProgram(ctx iris.Context) {
	bpID, err := ctx.Params().GetInt("bpID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	bp, db := models.BudgetProgram{ID: bpID}, ctx.Values().Get("db").(*gorm.DB)

	if err = db.First(&bp, bpID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(jsonError{"Suppression de programme : introuvable"})
		return
	}

	if err = db.Delete(&bp).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Programme supprimé"})
}
