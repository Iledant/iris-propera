package actions

import (
	"net/http"

	"github.com/jinzhu/gorm"

	"github.com/Iledant/iris_propera/models"
	"github.com/kataras/iris"
)

// GetPreProgrammings handle the get request to get all preprogrammings and all linked datas.
// The scope is all physical operations for ADMIN role or controlled operations for USER
func GetPreProgrammings(ctx iris.Context) {
	year, err := ctx.URLParamInt64("year")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste de la préprogrammation, paramètre : " + err.Error()})
		return
	}
	uID, err := getUserID(ctx)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste de la préprogrammation, user : " + err.Error()})
		return
	}
	var resp models.FullPreProgrammings
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = resp.GetAll(uID, year, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste de la préprogrammation, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// BatchPreProgrammings sets the pre programmings replacing existing one
func BatchPreProgrammings(ctx iris.Context) {
	var req models.PreProgrammingBatch
	err := ctx.ReadJSON(&req)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch préprogrammation, décodage : " + err.Error()})
		return
	}
	userID, err := getUserID(ctx)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch préprogrammation, user : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err = req.Save(userID, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch préprogrammation, requête : " + err.Error()})
		return
	}
	var resp models.FullPreProgrammings
	if err = resp.GetAll(userID, req.Year, db.DB()); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch préprogrammation, requête get : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
