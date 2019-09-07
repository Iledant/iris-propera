package actions

import (
	"database/sql"
	"net/http"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

// getRight is used for the frontend page dedicated to users rights on physical operations
type getRightResp struct {
	models.OpRights
	models.Users
	models.PhysicalOps
}

// SetRight give a user rights on physical operations
func SetRight(ctx iris.Context) {
	userID, err := ctx.Params().GetInt64("userID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation des droits, paramètre : " + err.Error()})
		return
	}
	var rights models.OpRights
	if err = ctx.ReadJSON(&rights); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Fixation des droits, décodage : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err = rights.UserSet(userID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation des droits, requête : " + err.Error()})
		return
	}
	var updatedRights models.OpRights
	if err = updatedRights.UserGet(userID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Fixation des droits, requête get : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(updatedRights)
}

// GetRight get rights of a user on physical operations and send back rights, list of users and physical operations list
func GetRight(ctx iris.Context) {
	userID, err := ctx.Params().GetInt64("userID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Droits d'un utilisateur, paramètre : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	var resp getRightResp
	if err = resp.OpRights.UserGet(userID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Droits d'un utilisateur, requête rights : " + err.Error()})
		return
	}
	if err = resp.Users.GetRole(models.UserRole, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Droits d'un utilisateur, requête users : " + err.Error()})
		return
	}
	if err = resp.PhysicalOps.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Droits d'un utilisateur, requête opérations : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// InheritRight add rights of users on physical operations to the given user
func InheritRight(ctx iris.Context) {
	userID, err := ctx.Params().GetInt64("userID")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Héritage de droit, paramètre : " + err.Error()})
		return
	}
	var req models.UsersIDs
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Héritage de droit, décodage : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err = req.Inherit(userID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Héritage de droit, requête : " + err.Error()})
		return
	}
	var updatedRights models.OpRights
	if err = updatedRights.UserGet(userID, db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Héritage de droit, requête get : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(updatedRights)
}
