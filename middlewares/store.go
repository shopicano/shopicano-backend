package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/data"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/log"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/shopicano/shopicano-backend/utils"
	"net/http"
)

func HasStore() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			resp := core.Response{}

			db := app.DB()

			su := data.NewStoreRepository()
			store, err := su.GetStoreUserProfile(db, ctx.Get(utils.UserID).(string))
			if err != nil {
				log.Log().Errorln(err)
				if errors.IsRecordNotFoundError(err) {
					resp.Status = http.StatusNotFound
					resp.Code = errors.StoreNotFound
					resp.Title = "Store not found"
					return resp.ServerJSON(ctx)
				}

				resp.Status = http.StatusInternalServerError
				resp.Code = errors.DatabaseQueryFailed
				resp.Title = "Database query failed"
				resp.Errors = err
				return resp.ServerJSON(ctx)
			}

			storeID := ctx.Param("store_id")
			if storeID != "" && storeID != store.StoreID {
				resp.Status = http.StatusForbidden
				resp.Code = errors.UnauthorizedStoreAccess
				resp.Title = "Unauthorized request"
				return resp.ServerJSON(ctx)
			}

			log.Log().Infoln(store.StaffPermission)
			log.Log().Infoln(store.StoreStatus)

			ctx.Set(utils.StoreID, store.StoreID)
			ctx.Set(utils.StorePermission, store.StaffPermission)
			ctx.Set(utils.StoreStatus, store.StoreStatus)
			return next(ctx)
		}
	}
}

func IsStoreActive() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			resp := core.Response{}

			if utils.GetStoreStatus(ctx) != models.StoreActive {
				resp.Status = http.StatusForbidden
				resp.Code = errors.StoreNotActive
				resp.Title = "Store not active"
				return resp.ServerJSON(ctx)
			}
			return next(ctx)
		}
	}
}

func IsStoreManager() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			resp := core.Response{}

			log.Log().Infoln(utils.GetStorePermission(ctx))

			if !(utils.GetStorePermission(ctx) == models.AdminPerm || utils.GetStorePermission(ctx) == models.ManagerPerm) {
				resp.Status = http.StatusForbidden
				resp.Code = errors.UnauthorizedStoreAccess
				resp.Title = "Unauthorized to access store as manager"
				return resp.ServerJSON(ctx)
			}
			return next(ctx)
		}
	}
}

func IsStoreAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			resp := core.Response{}

			if utils.GetStorePermission(ctx) != models.AdminPerm {
				resp.Status = http.StatusForbidden
				resp.Code = errors.UnauthorizedStoreAccess
				resp.Title = "Unauthorized to access store as admin"
				return resp.ServerJSON(ctx)
			}
			return next(ctx)
		}
	}
}
