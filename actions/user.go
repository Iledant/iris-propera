package actions

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
	"golang.org/x/crypto/bcrypt"
)

// returnedToken is used to send a unique JSON object for login
type returnedToken struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

// sentUser is used for creating or updating user
type sentUser struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Active   bool   `json:"active"`
}

// Login handles user login using credentials and return token if success.
func Login(ctx iris.Context) {
	email, password := ctx.URLParam("email"), ctx.URLParam("password")
	// Check parameters
	if email == "" || password == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Champ manquant ou incorrect"})
		return
	}

	db, user := ctx.Values().Get("db").(*gorm.DB), models.User{}

	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusNotFound)
			ctx.JSON(jsonError{"Erreur de login ou mot de passe"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(jsonError{"Erreur de login ou mot de passe"})
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
	u, err := bearerToUser(ctx)

	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Erreur de token"})
		return
	}

	userID, _ := strconv.Atoi(u.Subject)
	delToken(userID)

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Utilisateur déconnecté"})
}

// GetUsers handles the GET request for all users and send back only secure fields.
func GetUsers(ctx iris.Context) {
	users := []models.User{}
	db := ctx.Values().Get("db").(*gorm.DB)

	if err := db.Find(&users).Error; err != nil {
		ctx.JSON(jsonMessage{err.Error()})
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}

	ctx.JSON(struct {
		User []models.User `json:"user"`
	}{users})
	ctx.StatusCode(http.StatusOK)
}

// CreateUser handles the creation by admin of a new user and returns the created user.
func CreateUser(ctx iris.Context) {
	sent := sentUser{}

	if err := ctx.ReadJSON(&sent); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonMessage{err.Error()})
		return
	}

	// Check parameters
	if sent.Name == "" || sent.Email == "" || sent.Password == "" ||
		(sent.Role != models.UserRole && sent.Role != models.AdminRole && sent.Role != models.ObserverRole) {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonMessage{"Champ manquant ou incorrect"})
		return
	}

	newUser, db := sent.toUser(), ctx.Values().Get("db").(*gorm.DB)

	if err := usrExists(&newUser, db); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if err := setUserPwd(&newUser, sent.Password); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if err := db.Create(&newUser).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	response := struct {
		User models.User `json:"user"`
	}{newUser}
	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(response)
}

// UpdateUser handles the updating by admin of an existing user and sent back modified user.
func UpdateUser(ctx iris.Context) {
	userID, err := ctx.Params().GetInt("userID")

	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	db, user := ctx.Values().Get("db").(*gorm.DB), models.User{ID: userID}

	if err = db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusNotFound)
			ctx.JSON(jsonError{"Utilisateur introuvable"})
			return
		}
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	sent := sentUser{}
	if err = ctx.ReadJSON(&sent); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonMessage{err.Error()})
		return
	}

	if sent.Email != "" {
		user.Email = sent.Email
	}

	if sent.Name != "" {
		user.Name = sent.Name
	}

	user.Active = sent.Active

	if sent.Role != "" {
		if sent.Role != models.AdminRole && sent.Role != models.UserRole && sent.Role != models.ObserverRole {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(jsonError{Error: "Rôle différent de ADMIN, USER et OBSERVER"})
			return
		}
		user.Role = sent.Role
	}

	if sent.Password != "" {
		if err = setUserPwd(&user, sent.Password); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{err.Error()})
			return
		}
	}

	if err = db.Save(&user).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	response := struct {
		User models.User `json:"user"`
	}{user}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(response)
}

// DeleteUser handles the deleting by admin of an existing user.
func DeleteUser(ctx iris.Context) {
	userID, err := ctx.Params().GetInt("userID")

	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	db := ctx.Values().Get("db").(*gorm.DB)
	user := models.User{}

	if err = db.First(&user, userID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(jsonError{"Utilisateur introuvable"})
		return
	}

	if err = db.Delete(&user).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Utilisateur supprimé"})
}

// SignUp handles the request of a new user and creates an inactive account.
func SignUp(ctx iris.Context) {
	name, email, password := ctx.URLParam("name"), ctx.URLParam("email"), ctx.URLParam("password")

	// Parameters validation
	if name == "" || email == "" || password == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Champ manquant ou incorrect"})
		return
	}

	db := ctx.Values().Get("db").(*gorm.DB)
	user := models.User{Name: name, Email: email, Role: models.UserRole, Active: false}

	if err := usrExists(&user, db); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if err := setUserPwd(&user, password); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	if err := db.Create(&user).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
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
		ctx.JSON(jsonError{"Ancien et nouveau mots de passe requis"})
		return
	}

	u, err := bearerToUser(ctx)

	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	db, user := ctx.Values().Get("db").(*gorm.DB), models.User{}
	userID, _ := strconv.Atoi(u.Subject)

	if err = db.Find(&user, userID).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPwd)); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Erreur de mot de passe"})
		return
	}

	if err = setUserPwd(&user, newPwd); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{Error: err.Error()})
		return
	}

	if err = db.Model(&user).Update("password", user.Password).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{err.Error()})
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Mot de passe changé"})
}

// usrExists verifies if a name or email already in the users table.
func usrExists(user *models.User, db *gorm.DB) error {
	var count int
	err := db.Where("name = ? OR email = ?", user.Name, user.Email).Find(&user).Count(&count).Error

	if err == gorm.ErrRecordNotFound {
		return nil
	}

	if count > 0 {
		err = errors.New("Utilisateur existant")
	}

	return err
}

// setUserPwd sets and crypts password of a user.
func setUserPwd(u *models.User, pwd string) error {
	cryptPwd, err := bcrypt.GenerateFromPassword([]byte(pwd), 10)

	if err != nil {
		return err
	}

	u.Password = string(cryptPwd)

	return nil
}

// toUser convert to modes.USer
func (u sentUser) toUser() models.User {
	return models.User{Name: u.Name, Email: u.Email, Active: u.Active, Role: u.Role}
}
