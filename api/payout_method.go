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

func createPayoutMethod(ctx echo.Context) error {
	req, err := validators.ValidateCreatePayoutMethod(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.PayoutMethodDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()

	m := &models.PayoutMethod{
		ID:          utils.NewUUID(),
		Name:        req.Name,
		IsPublished: req.IsPublished,
		Inputs:      req.Inputs,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	au := data.NewMarketplaceRepository()
	if err := au.CreatePayoutMethod(db, m); err != nil {
		msg, ok := errors.IsDuplicateKeyError(err)
		if ok {
			resp.Title = msg
			resp.Status = http.StatusConflict
			resp.Code = errors.PayoutMethodAlreadyExists
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

func updatePayoutMethod(ctx echo.Context) error {
	pomID := ctx.Param("pom_id")

	req, err := validators.ValidateReqUpdatePayoutMethod(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.PayoutMethodDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()

	au := data.NewMarketplaceRepository()
	pom, err := au.GetPayoutMethod(db, pomID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Payout method not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.PayoutMethodNotFound
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
		pom.Name = *req.Name
	}
	if req.IsPublished != nil {
		pom.IsPublished = *req.IsPublished
	}
	if req.Inputs != nil {
		pom.Inputs = *req.Inputs
	}

	pom.UpdatedAt = time.Now().UTC()

	if err := au.UpdatePayoutMethod(db, pom); err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = pom
	return resp.ServerJSON(ctx)
}

func deletePayoutMethod(ctx echo.Context) error {
	pomID := ctx.Param("pom_id")

	resp := core.Response{}

	db := app.DB()
	au := data.NewMarketplaceRepository()
	if err := au.DeletePayoutMethod(db, pomID); err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusNoContent
	return resp.ServerJSON(ctx)
}

func listPayoutMethodsForUser(ctx echo.Context) error {
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
	poms, err := au.ListPayoutMethodForUser(db, int(from), int(limit))
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = poms
	return resp.ServerJSON(ctx)
}

func listPayoutMethods(ctx echo.Context) error {
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
	poms, err := au.ListPayoutMethods(db, int(from), int(limit))
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = poms
	return resp.ServerJSON(ctx)
}

func getPayoutMethodForUser(ctx echo.Context) error {
	pomID := ctx.Param("pom_id")

	resp := core.Response{}

	db := app.DB()

	au := data.NewMarketplaceRepository()
	pom, err := au.GetPayoutMethodForUser(db, pomID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Payout method not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.PayoutMethodNotFound
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
	resp.Data = pom
	return resp.ServerJSON(ctx)
}

func getPayoutMethod(ctx echo.Context) error {
	pomID := ctx.Param("pom_id")

	resp := core.Response{}

	db := app.DB()

	au := data.NewMarketplaceRepository()
	pom, err := au.GetPayoutMethod(db, pomID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Payout method not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.PayoutMethodNotFound
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
	resp.Data = pom
	return resp.ServerJSON(ctx)
}
