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

func createBusinessAccountType(ctx echo.Context) error {
	req, err := validators.ValidateCreateBusinessAccountType(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.BusinessAccountTypeDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()

	m := &models.BusinessAccountType{
		ID:          utils.NewUUID(),
		Name:        req.Name,
		IsPublished: req.IsPublished,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	au := data.NewMarketplaceRepository()
	if err := au.CreateBusinessAccountType(db, m); err != nil {
		msg, ok := errors.IsDuplicateKeyError(err)
		if ok {
			resp.Title = msg
			resp.Status = http.StatusConflict
			resp.Code = errors.BusinessAccountTypeAlreadyExists
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

func updateBusinessAccountType(ctx echo.Context) error {
	batID := ctx.Param("bat_id")

	req, err := validators.ValidateUpdateBusinessAccountType(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.BusinessAccountTypeDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()

	au := data.NewMarketplaceRepository()
	bat, err := au.GetBusinessAccountType(db, batID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Business account type not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.BusinessAccountTypeNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if req.Name != nil {
		bat.Name = *req.Name
	}
	if req.IsPublished != nil {
		bat.IsPublished = *req.IsPublished
	}

	bat.UpdatedAt = time.Now().UTC()

	if err := au.UpdateBusinessAccountType(db, bat); err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = bat
	return resp.ServerJSON(ctx)
}

func deleteBusinessAccountType(ctx echo.Context) error {
	batID := ctx.Param("bat_id")

	resp := core.Response{}

	db := app.DB()
	au := data.NewMarketplaceRepository()
	if err := au.DeleteBusinessAccountType(db, batID); err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusNoContent
	return resp.ServerJSON(ctx)
}

func listBusinessAccountTypesForUser(ctx echo.Context) error {
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

	from := (page - 1) * limit

	resp := core.Response{}

	db := app.DB()

	au := data.NewMarketplaceRepository()
	bats, err := au.ListBusinessAccountTypesForUser(db, int(from), int(limit))
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = bats
	return resp.ServerJSON(ctx)
}

func listBusinessAccountTypes(ctx echo.Context) error {
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

	from := (page - 1) * limit

	resp := core.Response{}

	db := app.DB()

	au := data.NewMarketplaceRepository()
	bats, err := au.ListBusinessAccountTypes(db, int(from), int(limit))
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = bats
	return resp.ServerJSON(ctx)
}

func getBusinessAccountTypeForUser(ctx echo.Context) error {
	batID := ctx.Param("bat_id")

	resp := core.Response{}

	db := app.DB()

	au := data.NewMarketplaceRepository()
	bat, err := au.GetBusinessAccountTypeForUser(db, batID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Business account type not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.BusinessAccountTypeNotFound
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
	resp.Data = bat
	return resp.ServerJSON(ctx)
}

func getBusinessAccountType(ctx echo.Context) error {
	batID := ctx.Param("bat_id")

	resp := core.Response{}

	db := app.DB()

	au := data.NewMarketplaceRepository()
	bat, err := au.GetBusinessAccountType(db, batID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Business account type not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.BusinessAccountTypeNotFound
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
	resp.Data = bat
	return resp.ServerJSON(ctx)
}
