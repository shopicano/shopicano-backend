package utils

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/config"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/models"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

const (
	PasswordCost = 11

	StoreID         = "store_id"
	StorePermission = "store_permission"
	StoreStatus     = "store_status"
	UserID          = "user_id"
	UserPermission  = "user_permission"
	UserStatus      = "user_status"
	Scope           = "user_scope"
)

const (
	Platform   UserScope = "platform"
	BackStore  UserScope = "back_store"
	FrontStore UserScope = "front_store"
)

type UserScope string

type Claims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

func GeneratePassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), PasswordCost)
	return string(bytes), err
}

func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func ParseBearerToken(ctx echo.Context) (string, error) {
	bearer := ctx.Request().Header.Get("Authorization")
	bearerWithToken := strings.Split(bearer, " ")

	if len(bearerWithToken) != 2 {
		return "", errors.NewError("Bearer token not found")
	}
	return bearerWithToken[1], nil
}

func IsStoreStaff(ctx echo.Context) bool {
	perm := ctx.Get(StorePermission)
	return ctx.Get(StoreID) != nil && perm != nil && (perm.(models.Permission) == models.ManagerPerm || perm.(models.Permission) == models.AdminPerm)
}

func IsPlatformAdmin(ctx echo.Context) bool {
	perm := ctx.Get(UserPermission)
	return perm != nil && (perm.(models.Permission) == models.ManagerPerm || perm.(models.Permission) == models.AdminPerm)
}

func GetStoreID(ctx echo.Context) string {
	return ctx.Get(StoreID).(string)
}

func GetUserID(ctx echo.Context) string {
	return ctx.Get(UserID).(string)
}

func GetUserStatus(ctx echo.Context) models.UserStatus {
	return ctx.Get(UserStatus).(models.UserStatus)
}

func GetStoreStatus(ctx echo.Context) models.StoreStatus {
	return ctx.Get(StoreStatus).(models.StoreStatus)
}

func GetUserPermission(ctx echo.Context) models.Permission {
	return ctx.Get(UserPermission).(models.Permission)
}

func GetStorePermission(ctx echo.Context) models.Permission {
	return ctx.Get(StorePermission).(models.Permission)
}

func BuildJWTToken(userID string, scope UserScope) (string, error) {
	claims := Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			Audience:  string(scope),
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.App().JWTKey))
}
