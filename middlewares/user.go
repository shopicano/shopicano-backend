package middlewares

import (
	"git.cloudbro.net/michaelfigg/yallawebsites/core"
	"git.cloudbro.net/michaelfigg/yallawebsites/errors"
	"git.cloudbro.net/michaelfigg/yallawebsites/models"
	"git.cloudbro.net/michaelfigg/yallawebsites/utils"
	"github.com/labstack/echo/v4"
	"net/http"
)

func IsUserActive(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		resp := core.Response{}

		if utils.GetUserStatus(ctx) != models.UserActive {
			resp.Status = http.StatusForbidden
			resp.Code = errors.UserNotActive
			resp.Title = "User isn't active"
			return resp.ServerJSON(ctx)
		}

		return next(ctx)
	}
}
