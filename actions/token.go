package actions

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Iledant/iris_propera/models"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/kataras/iris"
)

// UserClaims hold token user fields to avoid fetching database
type UserClaims struct {
	Role   string
	Active bool
	ID     int
}

// customClaims add role and active to token to avoid fetching database
type customClaims struct {
	Role   string `json:"rol"`
	Active bool   `json:"act"`
	jwt.StandardClaims
}

var (
	signingKey   = []byte(os.Getenv("JWT_SIGNING_KEY"))
	expireDelay  = time.Hour * 3
	refreshDelay = time.Hour * 15 * 24
	iss          = "https://www.propera.net/api/login"
	tokens       = map[int]customClaims{}
	// ErrNoToken happens when header have no or bad authorization bearer
	ErrNoToken = errors.New("Token absent")
	// ErrBadToken happends when bearer token can't be verified
	// or isn't already stored after login or refresh
	ErrBadToken = errors.New("Token invalide")
)

// getTokenString store claims and return JWT token string
func getTokenString(claims *customClaims) (string, error) {
	mutex := &sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(signingKey)

	if err != nil {
		return "", err
	}

	userID, err := strconv.Atoi(claims.Subject)

	if err != nil {
		return "", err
	}
	tokens[userID] = *claims

	return tokenString, nil
}

// setToken creates or update a token for a given user
func setToken(u *models.User) (string, error) {
	t := time.Now()
	claims := customClaims{u.Role, u.Active,
		jwt.StandardClaims{Subject: strconv.Itoa(u.ID), ExpiresAt: t.Add(expireDelay).Unix(),
			IssuedAt: t.Unix(), Issuer: iss}}

	return getTokenString(&claims)
}

// delToken remove user ID from list of stored tokens
func delToken(userID int) {
	mutex := &sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()
	delete(tokens, userID)
}

// refreshToken replace an existing expired token and add it to the response header
func refreshToken(ctx iris.Context, u *customClaims) error {
	t := time.Now()
	u.ExpiresAt = t.Add(expireDelay).Unix()
	u.IssuedAt = t.Unix()

	tokenString, err := getTokenString(u)

	if err != nil {
		return err
	}

	ctx.Header("Authorization", "Bearer "+tokenString)

	return nil
}

// bearerToUser gets user claims (ID, role, active) from token in request header
// and send refreshed token if first time expired
func bearerToUser(ctx iris.Context) (u *customClaims, err error) {

	bearer := ctx.GetHeader("Authorization")
	if len(bearer) < 8 {
		return nil, ErrNoToken
	}

	tokenString := strings.TrimPrefix(bearer, "Bearer ")

	if tokenString == "" {
		return nil, ErrNoToken
	}

	parser := jwt.Parser{ValidMethods: nil, UseJSONNumber: true, SkipClaimsValidation: true}
	token, err := parser.ParseWithClaims(tokenString, &customClaims{},
		func(token *jwt.Token) (interface{}, error) { return []byte(signingKey), nil })

	if err != nil || !token.Valid {
		return nil, ErrBadToken
	}

	claims := token.Claims.(*customClaims)

	// Check if previously disconnected or refreshed
	userID, _ := strconv.Atoi(claims.Subject)
	mutex := &sync.Mutex{}
	mutex.Lock()
	stored, ok := tokens[userID]
	mutex.Unlock()

	if !ok || stored != *claims {
		return nil, ErrBadToken
	}

	// Refresh if expired
	if t := time.Now().Unix(); t > claims.ExpiresAt {
		err = refreshToken(ctx, claims)
	}
	ctx.Values().Set("claims", claims)
	return claims, err
}

// userClaims converts a customClaims to UserClaims
func (c *customClaims) userClaims() *UserClaims {
	u := &UserClaims{Active: c.Active, Role: c.Role}

	u.ID, _ = strconv.Atoi(c.Subject)

	return u
}

// isActive check an existing token in header and, if succeed, parse returning user active field
func isActive(ctx iris.Context) (bool, error) {
	u, err := bearerToUser(ctx)

	if err != nil {
		return false, err
	}

	return u.Active, nil
}

// isAdmin check an existing token in header and, if succeed, parse check if user active and admin
func isAdmin(ctx iris.Context) (bool, error) {
	u, err := bearerToUser(ctx)

	if err != nil {
		return false, err
	}

	return u.Active && u.Role == models.AdminRole, nil
}

// isObserver check an existing token in header and, if succeed, parse check if user active and observer
func isObserver(ctx iris.Context) (bool, error) {
	u, err := bearerToUser(ctx)

	if err != nil {
		return false, err
	}

	return u.Active && u.Role == models.ObserverRole, nil
}

// AdminMiddleware checks if there's a token and if it belongs to admin user otherwise prompt error
func AdminMiddleware(ctx iris.Context) {
	admin, err := isAdmin(ctx)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{Error: err.Error()})
		ctx.StopExecution()
		return
	}
	if !admin {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(jsonError{Error: "Droits administrateur requis"})
		ctx.StopExecution()
		return
	}
	ctx.Next()
}

// ActiveMiddleware checks if there's a valid token and user is active otherwise prompt error
func ActiveMiddleware(ctx iris.Context) {
	active, err := isActive(ctx)

	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{Error: err.Error()})
		ctx.StopExecution()
		return
	}
	if !active {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(jsonError{Error: "Connexion requise"})
		ctx.StopExecution()
		return
	}
	ctx.Next()
}
