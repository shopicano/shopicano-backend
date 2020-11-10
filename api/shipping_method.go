package api

import (
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/data"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/shopicano/shopicano-backend/utils"
	"github.com/shopicano/shopicano-backend/validators"
	"net/http"
	"strconv"
	"time"
)

func createShippingMethod(ctx echo.Context) error {
	req, err := validators.ValidateCreateShippingMethod(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.ShippingMethodCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()

	m := &models.ShippingMethod{
		ID:                      utils.NewUUID(),
		Name:                    req.Name,
		IsPublished:             req.IsPublished,
		ApproximateDeliveryTime: req.ApproximateDeliveryTime,
		DeliveryCharge:          req.DeliveryCharge,
		WeightUnit:              req.WeightUnit,
		IsFlat:                  req.IsFlat,
		CreatedAt:               time.Now().UTC(),
		UpdatedAt:               time.Now().UTC(),
	}

	au := data.NewMarketplaceRepository()
	if err := au.CreateShippingMethod(db, m); err != nil {
		msg, ok := errors.IsDuplicateKeyError(err)
		if ok {
			resp.Title = msg
			resp.Status = http.StatusConflict
			resp.Code = errors.ShippingMethodAlreadyExists
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusCreated
	resp.Data = m
	return resp.ServerJSON(ctx)
}

func updateShippingMethod(ctx echo.Context) error {
	ID := ctx.Param("id")

	req, err := validators.ValidateCreateShippingMethod(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.ShippingMethodCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()

	au := data.NewMarketplaceRepository()
	m, err := au.GetShippingMethod(db, ID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Shipping method not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.ShippingMethodNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	m.Name = req.Name
	m.IsPublished = req.IsPublished
	m.ApproximateDeliveryTime = req.ApproximateDeliveryTime
	m.IsFlat = req.IsFlat
	m.DeliveryCharge = req.DeliveryCharge
	m.WeightUnit = req.WeightUnit
	m.UpdatedAt = time.Now().UTC()

	if err := au.UpdateShippingMethod(db, m); err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = m
	return resp.ServerJSON(ctx)
}

func deleteShippingMethod(ctx echo.Context) error {
	ID := ctx.Param("id")

	resp := core.Response{}

	db := app.DB()

	au := data.NewMarketplaceRepository()
	if err := au.DeleteShippingMethod(db, ID); err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Shipping method not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.ShippingMethodNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusNoContent
	return resp.ServerJSON(ctx)
}

func getShippingMethod(ctx echo.Context) error {
	ID := ctx.Param("id")

	resp := core.Response{}

	db := app.DB()

	au := data.NewMarketplaceRepository()
	sm, err := au.GetShippingMethod(db, ID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Shipping method not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.ShippingMethodNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = sm
	return resp.ServerJSON(ctx)
}

func getShippingMethodForUser(ctx echo.Context) error {
	ID := ctx.Param("id")

	resp := core.Response{}

	db := app.DB()

	au := data.NewMarketplaceRepository()
	sm, err := au.GetShippingMethod(db, ID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Shipping method not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.ShippingMethodNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = sm
	return resp.ServerJSON(ctx)
}

func listShippingMethodsAsAdmin(ctx echo.Context) error {
	pageQ := ctx.Request().URL.Query().Get("page")
	limitQ := ctx.Request().URL.Query().Get("limit")

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

	from := (page - 1) * limit
	au := data.NewMarketplaceRepository()

	var v interface{}
	v, err = au.ListShippingMethods(db, int(from), int(limit))

	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = v
	return resp.ServerJSON(ctx)
}
