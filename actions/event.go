package actions

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/Iledant/iris_propera/actions/queries"
	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// evvResp embeddes response for an array of events.
type evvResp struct {
	Event []models.Event `json:"Event"`
}

// evResp embeddes response for an single event.
type evResp struct {
	Event models.Event `json:"Event"`
}

// evReq is used for creation and modification of a event.
type evReq struct {
	PhysicalOpID *int       `json:"physical_op_id"`
	Name         *string    `json:"name"`
	Date         *time.Time `json:"date"`
	IsCertain    *bool      `json:"iscertain"`
	Descript     *string    `json:"descript"`
}

// GetEvents handles request get all events.
func GetEvents(ctx iris.Context) {
	opID, err := ctx.Params().GetInt("opID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	op, db := models.PhysicalOp{}, ctx.Values().Get("db").(*gorm.DB)
	if err = db.First(&op, opID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonMessage{"Liste des événements : opération introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	evv := []models.Event{}
	if err := db.Where("physical_op_id = ?", opID).Find(&evv).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(evvResp{evv})
}

// CreateEvent handles request post request to create a new event.
func CreateEvent(ctx iris.Context) {
	opID, err := ctx.Params().GetInt("opID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	op, db := models.PhysicalOp{}, ctx.Values().Get("db").(*gorm.DB)
	if err = db.First(&op, opID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonMessage{"Création d'événement : opération introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ev := evReq{}
	if err := ctx.ReadJSON(&ev); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if ev.Name == nil || len(*ev.Name) == 0 || len(*ev.Name) > 255 ||
		ev.Date == nil || ev.IsCertain == nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'événement, champ manquant ou incorrect"})
		return
	}

	newEv := models.Event{PhysicalOpID: opID, Name: *ev.Name, Date: *ev.Date, IsCertain: *ev.IsCertain}
	if ev.Descript == nil {
		newEv.Descript.Valid = false
	} else {
		newEv.Descript.Valid = true
		newEv.Descript.String = *ev.Descript
	}

	if err := db.Create(&newEv).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(evResp{newEv})
}

// ModifyEvent handles request put requestion to modify a event.
func ModifyEvent(ctx iris.Context) {
	evID, err := ctx.Params().GetInt("evID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ev, db := models.Event{}, ctx.Values().Get("db").(*gorm.DB)
	if err = db.First(&ev, evID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonMessage{"Modification d'événement : introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	req := evReq{}
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if req.Name != nil && len(*req.Name) > 0 && len(*req.Name) < 255 {
		ev.Name = *req.Name
	}

	if req.Date != nil {
		ev.Date = *req.Date
	}

	if req.IsCertain != nil {
		ev.IsCertain = *req.IsCertain
	}

	if req.Descript != nil {
		ev.Descript.Valid = true
		ev.Descript.String = *req.Descript
	}

	if err = db.Save(&ev).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(evResp{ev})
}

// DeleteEvent handles the request to delete an event.
func DeleteEvent(ctx iris.Context) {
	evID, err := ctx.Params().GetInt("evID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ev, db := models.Event{ID: evID}, ctx.Values().Get("db").(*gorm.DB)

	if err = db.First(&ev, evID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(jsonError{"Suppression d'événement : introuvable"})
		return
	}

	if err = db.Delete(&ev).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Événement supprimé"})
}

// nxtMonthEvt is used to fetch results from dedicated query
type nxtMonthEvt struct {
	ID        int       `json:"id" gorm:"column:id"`
	Date      time.Time `json:"date" gorm:"column:date"`
	Event     string    `json:"event" gorm:"column:event"`
	Operation string    `json:"operation" gorm:"column:operation"`
}

type nmeResp struct {
	Event []nxtMonthEvt `json:"Event"`
}

// GetNextMonthEvent handles request for first page according to roles admins have all operations, users only theirs.
func GetNextMonthEvent(ctx iris.Context) {
	u, err := bearerToUser(ctx)

	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	uID, err := strconv.Atoi(u.Subject)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	var rows *sql.Rows
	nmm, nm, db := []nxtMonthEvt{}, nxtMonthEvt{}, ctx.Values().Get("db").(*gorm.DB)
	if u.Role == models.AdminRole {
		rows, err = db.Raw(queries.SQLGetAdminNextMonthEvents).Rows()
	} else {
		rows, err = db.Raw(queries.SQLGetUserNextMonthEvents, uID).Rows()
	}

	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		db.ScanRows(rows, &nm)
		nmm = append(nmm, nm)
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(nmeResp{nmm})
}
