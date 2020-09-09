package middlewares

import (
	"git.cloudbro.net/michaelfigg/yallawebsites/core"
	"git.cloudbro.net/michaelfigg/yallawebsites/errors"
	"git.cloudbro.net/michaelfigg/yallawebsites/models"
	"git.cloudbro.net/michaelfigg/yallawebsites/utils"
	"github.com/labstack/echo/v4"
	"net/http"
)

func IsPlatformAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		resp := core.Response{}

		if utils.GetUserPermission(ctx) != models.AdminPerm {
			resp.Status = http.StatusForbidden
			resp.Code = errors.UnauthorizedStoreAccess
			resp.Title = "Unauthorized to access platform as admin"
			return resp.ServerJSON(ctx)
		}
		return next(ctx)
	}
}

func IsPlatformManager(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		resp := core.Response{}

		if !(utils.GetUserPermission(ctx) == models.AdminPerm || utils.GetUserPermission(ctx) == models.ManagerPerm) {
			resp.Status = http.StatusForbidden
			resp.Code = errors.UnauthorizedStoreAccess
			resp.Title = "Unauthorized to access platform as manager"
			return resp.ServerJSON(ctx)
		}
		return next(ctx)
	}
}
