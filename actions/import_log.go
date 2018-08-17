package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// illResp embeddes array of import logs for JSON response.
type illResp struct {
	ImportLogs []models.ImportLog `json:"ImportLog"`
}

// GetImportLogs handles to get request for log informations
func GetImportLogs(ctx iris.Context) {
	db, ill := ctx.Values().Get("db").(*gorm.DB), illResp{}

	if err := db.Find(&ill.ImportLogs).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(ill)
}
