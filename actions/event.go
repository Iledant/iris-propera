package actions

import (
	"database/sql"
	"net/http"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

// evResp embeddes response for an single event.
type evResp struct {
	Event models.Event `json:"Event"`
}

// GetEvents handles request get all events.
func GetEvents(ctx iris.Context) {
	opID, err := ctx.Params().GetInt64("opID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des événements, paramètre : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	var resp models.Events
	if err = resp.GetOpAll(opID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des événements, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// CreateEvent handles request post request to create a new event.
func CreateEvent(ctx iris.Context) {
	opID, err := ctx.Params().GetInt64("opID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'un événement, paramètre : " + err.Error()})
		return
	}
	var req models.Event
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'un événement, décodage : " + err.Error()})
		return
	}
	req.PhysicalOpID = opID
	if err = req.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'un événement : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := req.Create(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'un événement, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(evResp{req})
}

// ModifyEvent handles request put requestion to modify a event.
func ModifyEvent(ctx iris.Context) {
	evID, err := ctx.Params().GetInt64("evID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un événement, paramètre evID : " + err.Error()})
		return
	}
	opID, err := ctx.Params().GetInt64("opID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un événement, paramètre opID : " + err.Error()})
		return
	}
	var req models.Event
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un événement, décodage : " + err.Error()})
		return
	}
	req.PhysicalOpID = opID
	req.ID = evID
	if err = req.Validate(); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un événement : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err = req.Update(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'un événement, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(evResp{req})
}

// DeleteEvent handles the request to delete an event.
func DeleteEvent(ctx iris.Context) {
	evID, err := ctx.Params().GetInt64("evID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'un événement, décodage : " + err.Error()})
		return
	}
	ev, db := models.Event{ID: evID}, ctx.Values().Get("db").(*sql.DB)
	if err = ev.Delete(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'un événement, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Événement supprimé"})
}

// GetNextMonthEvent handles request for first page according to roles admins have all operations, users only theirs.
func GetNextMonthEvent(ctx iris.Context) {
	uID, err := getUserID(ctx)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Événements du prochain mois, user : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	var resp models.NextMonthEvents
	if err = resp.Get(uID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Événements du prochain mois, requête : "})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}
