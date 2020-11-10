package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/shopicano/shopicano-backend/utils"
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
