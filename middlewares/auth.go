package middlewares

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/config"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/data"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/utils"
	"net/http"
	"strings"
)

func JWTAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			resp := core.Response{}

			claims, token, err := extractAndValidateToken(ctx)
			if err != nil {
				resp.Status = http.StatusUnauthorized
				resp.Code = errors.InvalidAuthorizationToken
				resp.Title = "Unauthorized request"
				resp.Errors = err
				return resp.ServerJSON(ctx)
			}

			db := app.DB()
			userDao := data.NewUserRepository()

			u, err := userDao.Get(db, claims.UserID)
			if err != nil {
				if errors.IsRecordNotFoundError(err) {
					resp.Status = http.StatusUnauthorized
					resp.Code = errors.UserNotFound
					resp.Title = "User not found"
					resp.Errors = err
					return resp.ServerJSON(ctx)
				}

				resp.Status = http.StatusInternalServerError
				resp.Code = errors.DatabaseQueryFailed
				resp.Title = "Database query failed"
				resp.Errors = err
				return resp.ServerJSON(ctx)
			}

			_, permission, err := userDao.GetPermission(db, token.Raw)
			if err != nil {
				if errors.IsRecordNotFoundError(err) {
					resp.Status = http.StatusUnauthorized
					resp.Code = errors.InvalidAuthorizationToken
					resp.Title = "Token not found"
					resp.Errors = err
					return resp.ServerJSON(ctx)
				}

				resp.Status = http.StatusInternalServerError
				resp.Code = errors.DatabaseQueryFailed
				resp.Title = "Database query failed"
				resp.Errors = err
				return resp.ServerJSON(ctx)
			}

			ctx.Set(utils.UserID, claims.UserID)
			ctx.Set(utils.Scope, utils.UserScope(claims.Audience))
			ctx.Set(utils.UserPermission, *permission)
			ctx.Set(utils.UserStatus, u.Status)
			return next(ctx)
		}
	}
}

func extractTokenFromHeader(ctx echo.Context) string {
	tokenWithBearer := ctx.Request().Header.Get("Authorization")
	token := strings.Replace(tokenWithBearer, "Bearer", "", -1)
	return strings.TrimSpace(token)
}

func extractTokenFromUrl(ctx echo.Context) string {
	token := ctx.QueryParam("Authorization")
	return strings.TrimSpace(token)
}

func extractToken(ctx echo.Context) string {
	token := extractTokenFromHeader(ctx)
	if token != "" {
		return token
	}
	token = extractTokenFromUrl(ctx)
	if token != "" {
		return token
	}
	return ""
}

func extractAndValidateToken(ctx echo.Context) (*utils.Claims, *jwt.Token, error) {
	token := extractToken(ctx)
	if token == "" {
		return nil, nil, errors.NewError("Authorization token not found")
	}
	claims := utils.Claims{}
	jwtToken, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (i interface{}, err error) {
		return []byte(config.App().JWTKey), nil
	})
	if err != nil {
		return nil, nil, err
	}
	if !jwtToken.Valid {
		return nil, nil, errors.NewError("Token is invalid")
	}
	return &claims, jwtToken, nil
}
