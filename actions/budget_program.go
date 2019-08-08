package actions

import (
	"database/sql"
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/kataras/iris"
)

// bpResp embeddes response for an single budget program.
type bpResp struct {
	BudgetProgram models.BudgetProgram `json:"BudgetProgram"`
}

// GetChapterBudgetPrograms handles request get budget programs of a chapter.
func GetChapterBudgetPrograms(ctx iris.Context) {
	chpID, err := ctx.Params().GetInt64("chpID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Programmes d'un chapitre, paramètre : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	var resp models.BudgetPrograms
	if err = resp.GetAllChapterLinked(chpID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Programmes d'un chapitre, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetAllBudgetPrograms handles request get all budget programs.
func GetAllBudgetPrograms(ctx iris.Context) {
	var resp models.BudgetPrograms
	db := ctx.Values().Get("db").(*sql.DB)
	if err := resp.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des programmes budgétaires, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// CreateBudgetProgram handles request post request to create a new program.
func CreateBudgetProgram(ctx iris.Context) {
	chpID, err := ctx.Params().GetInt64("chpID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'un programme, paramètre : " + err.Error()})
		return
	}
	var req models.BudgetProgram
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'un programme, décodage : " + err.Error()})
		return
	}
	if err = req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'un programme : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	req.ChapterID = chpID
	if err = req.Create(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'un programme, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(bpResp{req})
}

// ModifyBudgetProgram handles request put requestion to modify a program.
func ModifyBudgetProgram(ctx iris.Context) {
	bpID, err := ctx.Params().GetInt64("bpID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}
	chpID, err := ctx.Params().GetInt64("chpID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	var req models.BudgetProgram
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un programme, décodage : " + err.Error()})
		return
	}
	if err = req.Validate(); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un programme : " + err.Error()})
		return
	}
	req.ID = bpID
	req.ChapterID = chpID
	if err = req.Update(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un programme, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(bpResp{req})
}

// DeleteBudgetProgram handles the request to delete an budget program.
func DeleteBudgetProgram(ctx iris.Context) {
	bpID, err := ctx.Params().GetInt64("bpID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}
	bp, db := models.BudgetProgram{ID: bpID}, ctx.Values().Get("db").(*sql.DB)
	if err = bp.Delete(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'un programme, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Programme supprimé"})
}

// BatchBudgetProgram handles the post request to import a batch of budget programs.
func BatchBudgetProgram(ctx iris.Context) {
	var req models.BudgetProgramBatch
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Batch de programmes budgétaires, décodage : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := req.Save(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch de programmes budgétaires, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Batch importé"})
}
