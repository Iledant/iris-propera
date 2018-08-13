package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/actions/queries"
	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

type baaResp struct {
	BudgetAction []models.BudgetAction `json:"BudgetAction"`
}

type baResp struct {
	BudgetAction models.BudgetAction `json:"BudgetAction"`
}

type baReq struct {
	Code     *string `json:"code"`
	Name     *string `json:"name"`
	SectorID *int    `json:"sector_id"`
}

// GetProgramBudgetActions handles request get budget actions of a program.
func GetProgramBudgetActions(ctx iris.Context) {
	prgID, err := ctx.Params().GetInt("prgID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	db := ctx.Values().Get("db").(*gorm.DB)
	rows, err := db.Raw("SELECT * FROM budget_action WHERE program_id = ?", prgID).Rows()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}
	defer rows.Close()

	arr, item := baaResp{}, models.BudgetAction{}
	for rows.Next() {
		rows.Scan(&item)
		arr.BudgetAction = append(arr.BudgetAction, item)
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(arr)
}

// GetAllBudgetActions handles request get all budget actions.
func GetAllBudgetActions(ctx iris.Context) {
	baa := []models.BudgetAction{}

	db := ctx.Values().Get("db").(*gorm.DB)
	if err := db.Find(&baa).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(baaResp{baa})
}

// CreateBudgetAction handles request post request to create a new action.
func CreateBudgetAction(ctx iris.Context) {
	prgID, err := ctx.Params().GetInt("prgID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ba := baReq{}
	if err = ctx.ReadJSON(&ba); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if ba.Code == nil || *ba.Code == "" || ba.Name == nil || *ba.Name == "" || ba.SectorID == nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'action budgétaire, champ manquant ou incorrect"})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)

	if db.Raw("SELECT id FROM budget_sector WHERE id = ?", *ba.SectorID).RecordNotFound() {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'action budgétaire, index secteur incorrect"})
		return
	}

	newBa := models.BudgetAction{Code: *ba.Code, Name: *ba.Name, ProgramID: prgID, SectorID: *ba.SectorID}

	if err = db.Create(&newBa).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(baResp{newBa})
}

// baSent is used for decoding one line of an array in the budget actions arrays
type baSent struct{ Code, Name, Sector string }

// baaSent is used
type baaSent struct {
	BudgetAction []baSent `json:"BudgetAction"`
}

// BatchBudgetActions handles request post an array of actions.
func BatchBudgetActions(ctx iris.Context) {
	var baa baaSent
	if err := ctx.ReadJSON(&baa); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	db := ctx.Values().Get("db").(*gorm.DB)
	tx := db.Begin()

	if err := tx.Exec(queries.SQLDropTempActionTable).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	if err := tx.Exec(queries.SQLCreateTempActionTable).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	for _, ba := range baa.BudgetAction {
		if len(ba.Code) < 7 {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Erreur lors de l'import, code trop court :" + ba.Code})
			tx.Rollback()
			return
		}
		cc, cf, cn, ac := ba.Code[0:1], ba.Code[1:3], ba.Code[3:6], ba.Code[6:]

		if err := tx.Exec(queries.SQLInsertTempAction, cc, cf, cn, ac, ba.Name, ba.Sector).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{err.Error()})
			tx.Rollback()
			return
		}
	}

	if err := tx.Exec(queries.SQLUpdateBudgetAction).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	if err := tx.Exec(queries.SQLInsertBudgetAction).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	tx.Exec(queries.SQLDropTempActionTable)
	tx.Commit()

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Actions mises à jour"})
}

// ModifyBudgetAction handles request put requestion to modify an action.
func ModifyBudgetAction(ctx iris.Context) {
	baID, err := ctx.Params().GetInt("baID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ba, db := models.BudgetAction{}, ctx.Values().Get("db").(*gorm.DB)

	if err = db.First(&ba, baID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonMessage{"Modification d'action : introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	req := baReq{}
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if req.Code != nil && *req.Code != "" && len(*req.Code) < 4 {
		ba.Code = *req.Code
	}

	if req.Name != nil && *req.Name != "" && len(*req.Name) < 100 {
		ba.Name = *req.Name
	}

	if err = db.Save(&ba).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(baResp{ba})
}

// DeleteBudgetAction handles the request to delete an budget action.
func DeleteBudgetAction(ctx iris.Context) {
	baID, err := ctx.Params().GetInt("baID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ba, db := models.BudgetAction{ID: baID}, ctx.Values().Get("db").(*gorm.DB)

	if err = db.First(&ba, baID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(jsonError{"Suppression d'action : introuvable"})
		return
	}

	if err = db.Delete(&ba).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Action supprimée"})
}
