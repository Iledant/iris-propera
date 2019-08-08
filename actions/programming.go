package actions

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Iledant/iris_propera/models"
	"github.com/kataras/iris"
)

// GetProgrammings handle the get request to fetch the programming of a year.
func GetProgrammings(ctx iris.Context) {
	year, err := ctx.URLParamInt64("year")
	if err != nil {
		year = int64(time.Now().Year())
	}
	var resp models.Programmings
	db := ctx.Values().Get("db").(*sql.DB)
	if err := resp.GetAll(year, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Programmation annuelle, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// GetProgrammingsYear handles the get request to get years with available programmation
func GetProgrammingsYear(ctx iris.Context) {
	db := ctx.Values().Get("db").(*sql.DB)
	var resp models.ProgrammingsYears
	if err := resp.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Années de programmation, select : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// BatchProgrammings handles the post request containing a full programmation for the current year.
func BatchProgrammings(ctx iris.Context) {
	var req models.ProgrammingBatch
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch programmation, décodage : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := req.Save(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch programmation, requête : " + err.Error()})
		return
	}
	var resp models.Programmings
	if err := resp.GetAll(req.Year, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch programmation, requête programmation : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
