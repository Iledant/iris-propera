package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// todayMsgResp embeddes the today message for JSON response
type todayMsgResp struct {
	TodayMessage models.TodayMessage `json:"TodayMessage"`
}

// GetTodayMessage handles the get request to fetch title and text
func GetTodayMessage(ctx iris.Context) {
	var tm models.TodayMessage
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := db.First(&tm, 1).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Today message requête : " + err.Error()})
			return
		}
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(todayMsgResp{tm})
}

// todayMsgReq is used to decode sent post request
type todayMsgReq struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// SetTodayMessage handles the set request to fetch title and text
func SetTodayMessage(ctx iris.Context) {
	var req todayMsgReq
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation today message, décodage : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*gorm.DB)
	if err := db.Exec("UPDATE today_messages SET title = ?, text = ? WHERE id = 1",
		req.Title, req.Text).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation today message, update : " + err.Error()})
		return
	}

	tm := models.TodayMessage{}
	if err := db.First(&tm, 1).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation today message, lecture base de données : " + err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(todayMsgResp{tm})
}
