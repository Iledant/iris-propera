package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

type baResp struct {
	BudgetAction models.BudgetAction `json:"BudgetAction"`
}

// GetProgramBudgetActions handles request get budget actions of a program.
func GetProgramBudgetActions(ctx iris.Context) {
	prgID, err := ctx.Params().GetInt("prgID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Actions budgétaires d'un programme, paramètre : " + err.Error()})
		return
	}
	var resp models.BudgetActions
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.GetAllPrgID(prgID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Actions budgétaires d'un programme, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetAllBudgetActions handles request get all budget actions.
func GetAllBudgetActions(ctx iris.Context) {
	fullCode := ctx.URLParam("FullCodeAction")
	if fullCode == "true" {
		var resp models.FullCodeBudgetActions
		db := ctx.Values().Get("db").(*gorm.DB)
		if err := resp.GetAll(db.DB()); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Liste des actions budgétaires, requête : " + err.Error()})
			return
		}
		ctx.StatusCode(http.StatusOK)
		ctx.JSON(resp)
	} else {
		var resp models.BudgetActions
		db := ctx.Values().Get("db").(*gorm.DB)
		if err := resp.GetAll(db.DB()); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Liste des actions budgétaires, requête : " + err.Error()})
			return
		}
		ctx.StatusCode(http.StatusOK)
		ctx.JSON(resp)
	}
}

// CreateBudgetAction handles request post request to create a new action.
func CreateBudgetAction(ctx iris.Context) {
	prgID, err := ctx.Params().GetInt64("prgID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'action budgétaire, paramètre : " + err.Error()})
		return
	}
	var req models.BudgetAction
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'action budgétaire, décodage : " + err.Error()})
		return
	}
	if err = req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'action budgétaire : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	req.ProgramID = prgID
	if err = req.Create(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'action budgétaire, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(baResp{req})
}

// BatchBudgetActions handles request post an array of actions.
func BatchBudgetActions(ctx iris.Context) {
	var baa models.BudgetActionsBatch
	if err := ctx.ReadJSON(&baa); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := baa.Save(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch budget action, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Actions mises à jour"})
}

// ModifyBudgetAction handles request put requestion to modify an action.
func ModifyBudgetAction(ctx iris.Context) {
	prgID, err := ctx.Params().GetInt64("prgID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'action budgétaire, paramètre : " + err.Error()})
		return
	}
	baID, err := ctx.Params().GetInt64("baID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'action budgétaire, paramètre : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	var req models.BudgetAction
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'action budgétaire, décodage : " + err.Error()})
		return
	}
	if err = req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Modification d'action budgétaire : " + err.Error()})
		return
	}
	req.ID = baID
	req.ProgramID = prgID
	if err = req.Update(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'action budgétaire, update : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(baResp{req})
}

// DeleteBudgetAction handles the request to delete an budget action.
func DeleteBudgetAction(ctx iris.Context) {
	baID, err := ctx.Params().GetInt64("baID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'action budgétaire, paramètre : " + err.Error()})
		return
	}
	ba, db := models.BudgetAction{ID: baID}, ctx.Values().Get("db").(*gorm.DB)
	if err = ba.Delete(db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'action budgétaire, delete : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Action supprimée"})
}
