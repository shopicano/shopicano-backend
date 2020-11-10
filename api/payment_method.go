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

func createPaymentMethod(ctx echo.Context) error {
	req, err := validators.ValidateCreatePaymentMethod(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.PaymentMethodCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	m := &models.PaymentMethod{
		ID:               utils.NewUUID(),
		Name:             req.Name,
		IsFlat:           req.IsFlat,
		IsOfflinePayment: req.IsOfflinePayment,
		MaxProcessingFee: req.MaxProcessingFee,
		MinProcessingFee: req.MinProcessingFee,
		ProcessingFee:    req.ProcessingFee,
		IsPublished:      req.IsPublished,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	db := app.DB()

	au := data.NewMarketplaceRepository()
	if err := au.CreatePaymentMethod(db, m); err != nil {
		msg, ok := errors.IsDuplicateKeyError(err)
		if ok {
			resp.Title = msg
			resp.Status = http.StatusConflict
			resp.Code = errors.PaymentMethodAlreadyExists
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

func updatePaymentMethod(ctx echo.Context) error {
	ID := ctx.Param("id")

	req, err := validators.ValidateCreatePaymentMethod(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.PaymentMethodCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()

	au := data.NewMarketplaceRepository()
	m, err := au.GetPaymentMethod(db, ID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Payment method not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.PaymentMethodNotFound
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
	m.IsFlat = req.IsFlat
	m.MaxProcessingFee = req.MaxProcessingFee
	m.MinProcessingFee = req.MinProcessingFee
	m.ProcessingFee = req.ProcessingFee
	m.IsOfflinePayment = req.IsOfflinePayment
	m.UpdatedAt = time.Now().UTC()

	if err := au.UpdatePaymentMethod(db, m); err != nil {
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

func deletePaymentMethod(ctx echo.Context) error {
	ID := ctx.Param("id")

	resp := core.Response{}

	db := app.DB()

	au := data.NewMarketplaceRepository()
	if err := au.DeletePaymentMethod(db, ID); err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Payment method not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.PaymentMethodNotFound
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

func getPaymentMethod(ctx echo.Context) error {
	ID := ctx.Param("id")

	resp := core.Response{}

	db := app.DB()

	au := data.NewMarketplaceRepository()
	pm, err := au.GetPaymentMethod(db, ID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Payment method not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.PaymentMethodNotFound
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
	resp.Data = pm
	return resp.ServerJSON(ctx)
}

func getPaymentMethodForUser(ctx echo.Context) error {
	ID := ctx.Param("id")

	resp := core.Response{}

	db := app.DB()

	au := data.NewMarketplaceRepository()
	pm, err := au.GetPaymentMethodForUser(db, ID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Payment method not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.PaymentMethodNotFound
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
	resp.Data = pm
	return resp.ServerJSON(ctx)
}

func listPaymentMethodsAsAdmin(ctx echo.Context) error {
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
	v, err = au.ListPaymentMethods(db, int(from), int(limit))

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
