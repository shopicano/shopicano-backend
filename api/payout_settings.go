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
	"time"
)

func createOrUpdatePayoutSettings(ctx echo.Context) error {
	req, err := validators.ValidateCreateOrUpdatePayoutSettings(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.PayoutSettingsDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB().Begin()
	au := data.NewMarketplaceRepository()

	isNew := false
	m, err := au.GetPayoutSettings(db, utils.GetStoreID(ctx))
	if err != nil {
		if !errors.IsRecordNotFoundError(err) {
			db.Rollback()

			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		isNew = true
		m = &models.PayoutSettings{}
		m.ID = utils.NewUUID()
		m.StoreID = utils.GetStoreID(ctx)
		m.CreatedAt = time.Now().UTC()
	}

	m.CountryID = req.CountryID
	m.AccountTypeID = req.AccountTypeID
	m.BusinessName = req.BusinessName
	m.BusinessAddressID = req.BusinessAddressID
	m.VatNumber = req.VatNumber
	m.PayoutMethodID = req.PayoutMethodID
	m.PayoutMethodDetails = req.PayoutMethodDetails
	m.PayoutMinimumThreshold = req.PayoutMinimumThreshold
	m.UpdatedAt = time.Now().UTC()

	if isNew {
		err := au.CreatePayoutSettings(db, m)
		if err != nil {
			db.Rollback()

			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}
	} else {
		err := au.UpdatePayoutSettings(db, m)
		if err != nil {
			db.Rollback()

			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}
	}

	if err := db.Commit().Error; err != nil {
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

func getPayoutSettings(ctx echo.Context) error {
	resp := core.Response{}

	db := app.DB()

	au := data.NewMarketplaceRepository()
	m, err := au.GetPayoutSettingsDetails(db, utils.GetStoreID(ctx))
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Payout settings not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.PayoutSettingsNotFound
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
	resp.Data = m
	return resp.ServerJSON(ctx)
}
