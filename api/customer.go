package api

import (
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/data"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/middlewares"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/shopicano/shopicano-backend/utils"
	"net/http"
	"strconv"
)

func RegisterCustomerRoutes(publicEndpoints, platformEndpoints *echo.Group) {
	customersPlatformPath := platformEndpoints.Group("/customers")

	func(g echo.Group) {
		g.Use(middlewares.HasStore())
		g.Use(middlewares.IsStoreActive())
		g.Use(middlewares.IsStoreManager())
		g.GET("/", listCustomers)
	}(*customersPlatformPath)
}

func listCustomers(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)

	pageQ := ctx.Request().URL.Query().Get("page")
	limitQ := ctx.Request().URL.Query().Get("limit")
	query := ctx.Request().URL.Query().Get("query")

	page, err := strconv.ParseInt(pageQ, 10, 64)
	if err != nil {
		page = 1
	}
	limit, err := strconv.ParseInt(limitQ, 10, 64)
	if err != nil {
		limit = 10
	}

	resp := core.Response{}

	db := app.DB()

	offset := (page - 1) * limit

	cr := data.NewCustomerRepository()

	var customers []models.Customer

	if query == "" {
		customers, err = cr.List(db, storeID, int(offset), int(limit))
	} else {
		customers, err = cr.Search(db, query, storeID, int(offset), int(limit))
	}

	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = customers
	return resp.ServerJSON(ctx)
}
