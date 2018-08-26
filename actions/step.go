package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// stsResp embeddes response for an array of steps.
type stsResp struct {
	Step []models.Step `json:"Step"`
}

// stResp embeddes response for an single step.
type stResp struct {
	Step models.Step `json:"Step"`
}

// stReq is used for creation and modification of a step.
type stReq struct {
	Name *string `json:"name"`
}

// GetSteps handles request get all steps.
func GetSteps(ctx iris.Context) {
	sts := []models.Step{}

	db := ctx.Values().Get("db").(*gorm.DB)
	if err := db.Find(&sts).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des étapes, requête  : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(stsResp{sts})
}

// CreateStep handles request post request to create a new step.
func CreateStep(ctx iris.Context) {
	st := stReq{}
	if err := ctx.ReadJSON(&st); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'étape, décoage : " + err.Error()})
		return
	}

	if st.Name == nil || len(*st.Name) == 0 || len(*st.Name) > 50 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'étape, champ 'name' manquant ou incorrect"})
		return
	}

	db := ctx.Values().Get("db").(*gorm.DB)
	newSt := models.Step{Name: *st.Name}

	if err := db.Create(&newSt).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(stResp{newSt})
}

// ModifyStep handles request put requestion to modify a step.
func ModifyStep(ctx iris.Context) {
	stID, err := ctx.Params().GetInt("stID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'étape, décodage paramètre : " + err.Error()})
		return
	}

	st, db := models.Step{}, ctx.Values().Get("db").(*gorm.DB)

	if err = db.First(&st, stID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonMessage{"Modification d'étape : introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	req := stReq{}
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'étape, décodage payload :" + err.Error()})
		return
	}

	if req.Name != nil && len(*req.Name) > 0 && len(*req.Name) < 50 {
		st.Name = *req.Name
	}

	if err = db.Save(&st).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(stResp{st})
}

// DeleteStep handles the request to delete an step.
func DeleteStep(ctx iris.Context) {
	stID, err := ctx.Params().GetInt("stID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'étape, décodage paramètre : " + err.Error()})
		return
	}

	st, db := models.Step{ID: stID}, ctx.Values().Get("db").(*gorm.DB)

	if err = db.First(&st, stID).Error; err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Suppression d'étape : introuvable"})
		return
	}

	if err = db.Delete(&st).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'étape, delete : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Etape supprimée"})
}
