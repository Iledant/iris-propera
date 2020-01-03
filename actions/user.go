package actions

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/Iledant/iris-propera/models"
	"github.com/kataras/iris"
)

// returnedToken is used to send a unique JSON object for login
type returnedToken struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

type userResp struct {
	User models.User `json:"User"`
}

// sentUser is used for creating or updating user
type sentUser struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Active   bool   `json:"active"`
}

// credentials is used to decode user login payload
type credentials struct {
	Email    *string
	Password *string
}

// Login handles user login using credentials and return token if success.
func Login(ctx iris.Context) {
	c := credentials{}
	if err := ctx.ReadJSON(&c); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Décodage login : " + err.Error()})
	}
	if c.Email == nil || *c.Email == "" || c.Password == nil || *c.Password == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Champ manquant ou incorrect"})
		return
	}
	db, user := ctx.Values().Get("db").(*sql.DB), models.User{}
	if err := user.GetByEmail(*c.Email, db); err != nil {
		if err == models.ErrBadCredential {
			ctx.StatusCode(http.StatusBadRequest)
		} else {
			ctx.StatusCode(http.StatusInternalServerError)
		}
		ctx.JSON(jsonError{err.Error()})
		return
	}
	if err := user.ValidatePwd(*c.Password); err != nil {
		if err == models.ErrBadCredential {
			ctx.StatusCode(http.StatusBadRequest)
		} else {
			ctx.StatusCode(http.StatusInternalServerError)
		}
		ctx.JSON(jsonError{err.Error()})
		return
	}
	token, err := setToken(&user)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(returnedToken{token, user})
}

// Logout handles users logout and destroy his token.
func Logout(ctx iris.Context) {
	userID := ctx.Values().Get("uID").(int)
	delToken(userID)
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Utilisateur déconnecté"})
}

// GetUsers handles the GET request for all users and send back only secure fields.
func GetUsers(ctx iris.Context) {
	var users models.Users
	db := ctx.Values().Get("db").(*sql.DB)
	if err := users.GetAll(db); err != nil {
		ctx.JSON(jsonMessage{"Liste des utilisateurs : " + err.Error()})
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	ctx.JSON(users)
	ctx.StatusCode(http.StatusOK)
}

// CreateUser handles the creation by admin of a new user and returns the created user.
func CreateUser(ctx iris.Context) {
	var req sentUser
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonMessage{"Création d'utilisateur : " + err.Error()})
		return
	}
	if req.Name == "" || req.Email == "" || req.Password == "" ||
		(req.Role != models.UserRole && req.Role != models.AdminRole && req.Role != models.ObserverRole) {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonMessage{"Création d'utilisateur : Champ manquant ou incorrect"})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	user := models.User{Name: req.Name, Email: req.Email, Active: req.Active, Role: req.Role, Password: req.Password}
	if err := user.Exists(db); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création d'utilisateur : " + err.Error()})
		return
	}
	if err := user.CryptPwd(); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'utilisateur, cryptage : " + err.Error()})
		return
	}
	if err := user.Create(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création d'utilisateur, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(userResp{user})
}

// UpdateUser handles the updating by admin of an existing user and sent back modified user.
func UpdateUser(ctx iris.Context) {
	userID, err := ctx.Params().GetInt("userID")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Modification d'utilisateur, paramètre : " + err.Error()})
		return
	}
	db, user := ctx.Values().Get("db").(*sql.DB), models.User{ID: userID}
	if err = user.GetByID(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'utilisateur, requête get : " + err.Error()})
		return
	}
	var req sentUser
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'utilisateur, décodage : " + err.Error()})
		return
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	user.Active = req.Active
	if req.Role != "" {
		if req.Role != models.AdminRole && req.Role != models.UserRole && req.Role != models.ObserverRole {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{"Modification d'utilisateur, rôle incorrect"})
			return
		}
		user.Role = req.Role
	}
	if req.Password != "" {
		user.Password = req.Password
		if err = user.CryptPwd(); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Modification d'utilisateur, mot de passe : " + err.Error()})
			return
		}
	}
	if err = user.Update(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification d'utilisateur, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(userResp{user})
}

// DeleteUser handles the deleting by admin of an existing user.
func DeleteUser(ctx iris.Context) {
	userID, err := ctx.Params().GetInt("userID")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Suppression d'utilisateur, paramètre : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	user := models.User{ID: userID}
	if err = user.Delete(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression d'utilisateur, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Utilisateur supprimé"})
}

type signUpReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignUp handles the request of a new user and creates an inactive account.
func SignUp(ctx iris.Context) {
	var req signUpReq
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Inscription d'utilisateur, décodage : " + err.Error()})
		return
	}
	if req.Name == "" || req.Email == "" || req.Password == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Inscription d'utilisateur : nom, email ou mot de passe manquant"})
		return
	}
	var user models.User
	user.Name = req.Name
	user.Email = req.Email
	user.Password = req.Password
	user.Role = models.UserRole
	user.Active = false
	db := ctx.Values().Get("db").(*sql.DB)
	if err := user.Exists(db); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Inscription d'utilisateur, exists : " + err.Error()})
		return
	}
	if err := user.CryptPwd(); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Inscription d'utilisateur, password : " + err.Error()})
		return
	}
	if err := user.Create(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Inscription d'utilisateur, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(jsonMessage{"Utilisateur créé, en attente d'activation"})
}

// ChangeUserPwd handles the request of a user to change his password.
func ChangeUserPwd(ctx iris.Context) {
	currentPwd, newPwd := ctx.URLParam("current_password"), ctx.URLParam("password")
	if currentPwd == "" || newPwd == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Changement de mot de passe : Ancien et nouveau mots de passe requis"})
		return
	}
	userID := ctx.Values().Get("uID").(int)
	db, user := ctx.Values().Get("db").(*sql.DB), models.User{ID: userID}
	if err := user.GetByID(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Changement de mot de passe, get : " + err.Error()})
		return
	}
	if err := user.ValidatePwd(currentPwd); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Changement de mot de passe : Erreur de mot de passe"})
		return
	}
	user.Password = newPwd
	if err := user.CryptPwd(); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Changement de mot de passe, password : " + err.Error()})
		return
	}
	if err := user.Update(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Changement de mot de passe, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Mot de passe changé"})
}

// getUserRoleAndID fetch user role and ID with the token
func getUserID(ctx iris.Context) (uID int64, err error) {
	u := ctx.Values().Get("uID")
	role := ctx.Values().Get("role")
	if u == nil || role == nil {
		return 0, errors.New("Utilisateur non enregistré")
	}
	uID = int64(u.(int))
	if role.(string) == models.AdminRole {
		uID = 0
	}
	return uID, nil
}
