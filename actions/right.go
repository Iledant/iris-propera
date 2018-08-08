package actions

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// right is used to deal with users rights (set, get, inherit) with IDs for users or physical operations.
type right struct {
	IDs []int `json:"Right"`
}

// getRight is used for the frontend page dedicated to users rights on physical operations
type getRight struct {
	right
	User       []models.User       `json:"User"`
	PhysicalOp []models.PhysicalOp `json:"PhysicalOp"`
}

// SetRight give a user rights on physical operations
func SetRight(ctx iris.Context) {
	userID, err := ctx.Params().GetInt("userID")

	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	db, user := ctx.Values().Get("db").(*gorm.DB), models.User{}

	if err = db.First(&user, userID).Error; err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Utilisateur introuvable"})
		return
	}

	rights := right{}
	if err = ctx.ReadJSON(&rights); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	// Check if all physical operations IDs exist otherwise break
	if len(rights.IDs) > 0 {
		var count struct{ Count int }
		err = db.Raw("SELECT count(id) FROM physical_op WHERE id IN (?)", rights.IDs).Scan(&count).Error

		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{err.Error()})
			return
		}

		if count.Count != len(rights.IDs) {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Mauvais identificateur d'opération"})
			return
		}
	}

	tx := db.Begin()

	// Clear rights of the user
	if err = tx.Exec("DELETE from rights WHERE users_id = ?", userID).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		tx.Rollback()
		return
	}

	// Set new rights
	if len(rights.IDs) > 0 {
		tuples, tuple := []string{}, ""
		for _, pID := range rights.IDs {
			tuple = "(" + strconv.Itoa(userID) + "," + strconv.Itoa(pID) + ")"
			tuples = append(tuples, tuple)
		}
		query := "INSERT INTO rights (users_id, physical_op_id) VALUES" + strings.Join(tuples, ",")
		if err = tx.Exec(query).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{err.Error()})
			tx.Rollback()
			return
		}
	}
	tx.Commit()

	// Send back updated rights
	updatedRight, err := getUserRight(userID, db)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(*updatedRight)
}

// GetRight get rights of an users on physical operations and send back rights, list of users and physical operations list
func GetRight(ctx iris.Context) {
	userID, err := ctx.Params().GetInt("userID")

	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	db, user := ctx.Values().Get("db").(*gorm.DB), models.User{}

	if err = db.Find(&user, userID).Error; err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Utilisateur introuvable"})
		return
	}

	rights, err := getUserRight(userID, db)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	users := []models.User{}
	if err = db.Where("role = 'USER'").Find(&users).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	physicalOps := []models.PhysicalOp{}
	if err = db.Find(&physicalOps).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	response := getRight{*rights, users, physicalOps}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(response)
}

// InheritRight add rights of users on physical operations to the given user
func InheritRight(ctx iris.Context) {
	userID, err := ctx.Params().GetInt("userID")

	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	db, user := ctx.Values().Get("db").(*gorm.DB), models.User{}

	if err = db.Find(&user, userID).Error; err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Utilisateur introuvable"})
		return
	}

	rights := right{}
	if err = ctx.ReadJSON(&rights); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	// Check if all users op ID exists and update rights otherwise break
	if len(rights.IDs) > 0 {
		var count struct{ Count int }
		if err = db.Raw("SELECT count(id) FROM users WHERE id IN (?)", rights.IDs).Scan(&count).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{err.Error()})
			return
		}

		if count.Count != len(rights.IDs) {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Mauvais identificateur d'utilisateur"})
			return
		}

		qry :=
			`INSERT INTO rights (users_id, physical_op_id) SELECT ?, * FROM 
		  (SELECT DISTINCT physical_op_id FROM rights WHERE users_id IN (?) ) ids 
			 WHERE ids.physical_op_id NOT IN (SELECT physical_op_id FROM rights WHERE users_id = ?)`
		if err = db.Exec(qry, userID, rights.IDs, userID).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{err.Error()})
			return
		}
	}

	// Send back updated rights
	updatedRight, err := getUserRight(userID, db)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(*updatedRight)
}

// getUserRight return the rights in the database
func getUserRight(userID int, db *gorm.DB) (*right, error) {
	rows, err := db.Raw("SELECT physical_op_id FROM rights WHERE users_id = ?", userID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rights, r := right{IDs: []int{}}, 0
	for rows.Next() {
		rows.Scan(&r)
		rights.IDs = append(rights.IDs, r)
	}

	return &rights, nil
}
