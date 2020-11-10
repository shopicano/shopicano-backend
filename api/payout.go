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

func createPayoutEntry(ctx echo.Context) error {
	req, err := validators.ValidateCreatePayoutEntry(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.PayoutEntryDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB().Begin()
	au := data.NewMarketplaceRepository()
	su := data.NewStoreRepository()

	sv, err := su.GetStoreFinanceSummary(db, utils.GetStoreID(ctx))
	if err != nil {
		if !errors.IsRecordNotFoundError(err) {
			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		sv = &models.StoreFinanceSummaryView{
			StoreID:         utils.GetStoreID(ctx),
			TotalCommission: 0,
			TotalEarnings:   0,
			TotalIncome:     0,
		}
	}

	pv, err := su.GetStorePayoutSummary(db, utils.GetStoreID(ctx))
	if err != nil {
		if !errors.IsRecordNotFoundError(err) {
			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		pv = &models.StorePayoutSummaryView{
			StoreID:        utils.GetStoreID(ctx),
			TotalEarnings:  sv.TotalEarnings,
			TotalAvailable: sv.TotalEarnings,
			TotalPaid:      0,
		}
	}

	if pv.TotalAvailable-req.Amount < 0 {
		resp.Title = "Payout amount is invalid"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.PayoutAmountInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	psd, err := au.GetPayoutSettingsDetails(db, utils.GetStoreID(ctx))
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

	m := &models.PayoutSend{
		ID:                     utils.NewUUID(),
		StoreID:                utils.GetStoreID(ctx),
		InitiatedByUserID:      utils.GetUserID(ctx),
		Amount:                 req.Amount,
		Note:                   req.Note,
		Status:                 models.PayoutSendStatusPending,
		IsMarketplaceInitiated: false,
		FailureReason:          "",
		Highlights:             "",
		PayoutMethodID:         psd.PayoutMethodID,
		PayoutMethodDetails:    psd.PayoutMethodDetails,
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
	}

	if err := au.CreatePayoutEntry(db, m); err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if err := db.Commit().Error; err != nil {
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

func createPayoutEntryByMarketplace(ctx echo.Context) error {
	req, err := validators.ValidateCreatePayoutEntry(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.PayoutEntryDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	storeID := ctx.Param("store_id")

	db := app.DB().Begin()
	au := data.NewMarketplaceRepository()
	su := data.NewStoreRepository()

	sv, err := su.GetStoreFinanceSummary(db, storeID)
	if err != nil {
		if !errors.IsRecordNotFoundError(err) {
			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		sv = &models.StoreFinanceSummaryView{
			StoreID:         storeID,
			TotalCommission: 0,
			TotalEarnings:   0,
			TotalIncome:     0,
		}
	}

	pv, err := su.GetStorePayoutSummary(db, storeID)
	if err != nil {
		if !errors.IsRecordNotFoundError(err) {
			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		pv = &models.StorePayoutSummaryView{
			StoreID:        storeID,
			TotalEarnings:  sv.TotalEarnings,
			TotalAvailable: sv.TotalEarnings,
			TotalPaid:      0,
		}
	}

	if pv.TotalAvailable-req.Amount < 0 {
		resp.Title = "Payout amount is invalid"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.PayoutAmountInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	psd, err := au.GetPayoutSettingsDetails(db, storeID)
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

	m := &models.PayoutSend{
		ID:                     utils.NewUUID(),
		StoreID:                storeID,
		InitiatedByUserID:      utils.GetUserID(ctx),
		Amount:                 req.Amount,
		Note:                   req.Note,
		Status:                 models.PayoutSendStatusConfirmed,
		IsMarketplaceInitiated: true,
		FailureReason:          "",
		Highlights:             "",
		PayoutMethodID:         psd.PayoutMethodID,
		PayoutMethodDetails:    psd.PayoutMethodDetails,
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
	}

	if err := au.CreatePayoutEntry(db, m); err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if err := db.Commit().Error; err != nil {
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

func getStorePayoutSummary(ctx echo.Context) error {
	resp := core.Response{}

	db := app.DB()
	su := data.NewStoreRepository()

	sv, err := su.GetStoreFinanceSummary(db, utils.GetStoreID(ctx))
	if err != nil {
		if !errors.IsRecordNotFoundError(err) {
			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		sv = &models.StoreFinanceSummaryView{
			StoreID:         utils.GetStoreID(ctx),
			TotalCommission: 0,
			TotalEarnings:   0,
			TotalIncome:     0,
		}
	}

	pv, err := su.GetStorePayoutSummary(db, utils.GetStoreID(ctx))
	if err != nil {
		if !errors.IsRecordNotFoundError(err) {
			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		pv = &models.StorePayoutSummaryView{
			StoreID:        utils.GetStoreID(ctx),
			TotalEarnings:  sv.TotalEarnings,
			TotalAvailable: sv.TotalEarnings,
			TotalPaid:      0,
			TotalRequested: 0,
		}
	}

	resp.Status = http.StatusOK
	resp.Data = map[string]interface{}{
		"total_income":     sv.TotalIncome,
		"total_earnings":   sv.TotalEarnings,
		"total_commission": sv.TotalCommission,
		"total_requested":  pv.TotalRequested,
		"total_available":  pv.TotalAvailable,
		"total_paid":       pv.TotalPaid,
	}
	return resp.ServerJSON(ctx)
}

func getStorePayoutSummaryByMarketplace(ctx echo.Context) error {
	resp := core.Response{}

	storeID := ctx.Param("store_id")

	db := app.DB()
	su := data.NewStoreRepository()

	sv, err := su.GetStoreFinanceSummary(db, storeID)
	if err != nil {
		if !errors.IsRecordNotFoundError(err) {
			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		sv = &models.StoreFinanceSummaryView{
			StoreID:         utils.GetStoreID(ctx),
			TotalCommission: 0,
			TotalEarnings:   0,
			TotalIncome:     0,
		}
	}

	pv, err := su.GetStorePayoutSummary(db, storeID)
	if err != nil {
		if !errors.IsRecordNotFoundError(err) {
			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		pv = &models.StorePayoutSummaryView{
			StoreID:        utils.GetStoreID(ctx),
			TotalEarnings:  sv.TotalEarnings,
			TotalAvailable: sv.TotalEarnings,
			TotalPaid:      0,
			TotalRequested: 0,
		}
	}

	resp.Status = http.StatusOK
	resp.Data = map[string]interface{}{
		"total_income":     sv.TotalIncome,
		"total_earnings":   sv.TotalEarnings,
		"total_commission": sv.TotalCommission,
		"total_requested":  pv.TotalRequested,
		"total_available":  pv.TotalAvailable,
		"total_paid":       pv.TotalPaid,
	}
	return resp.ServerJSON(ctx)
}

func listPayoutEntries(ctx echo.Context) error {
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

	entries, err := au.ListPayoutEntries(db, utils.GetStoreID(ctx), int(from), int(limit))
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = entries
	return resp.ServerJSON(ctx)
}

func listPayoutEntriesByMarketplace(ctx echo.Context) error {
	storeID := ctx.Param("store_id")
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

	entries, err := au.ListPayoutEntries(db, storeID, int(from), int(limit))
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = entries
	return resp.ServerJSON(ctx)
}

func getPayoutEntry(ctx echo.Context) error {
	entryID := ctx.Param("entry_id")

	resp := core.Response{}

	db := app.DB()
	au := data.NewMarketplaceRepository()

	entry, err := au.GetPayoutEntryDetails(db, utils.GetStoreID(ctx), entryID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Payout entry not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.PayoutEntryNotFound
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
	resp.Data = entry
	return resp.ServerJSON(ctx)
}

func getPayoutEntryByMarketplace(ctx echo.Context) error {
	entryID := ctx.Param("entry_id")
	storeID := ctx.Param("store_id")

	resp := core.Response{}

	db := app.DB()
	au := data.NewMarketplaceRepository()

	entry, err := au.GetPayoutEntryDetails(db, storeID, entryID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Payout entry not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.PayoutEntryNotFound
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
	resp.Data = entry
	return resp.ServerJSON(ctx)
}

func updatePayoutEntryByMarketplace(ctx echo.Context) error {
	entryID := ctx.Param("entry_id")
	storeID := ctx.Param("store_id")

	resp := core.Response{}

	pld, err := validators.ValidateUpdatePayoutEntry(ctx)
	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.PayoutEntryDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()
	au := data.NewMarketplaceRepository()

	entry, err := au.GetPayoutEntry(db, storeID, entryID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Payout entry not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.PayoutEntryNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if pld.Status != nil {
		entry.Status = *pld.Status
	}
	if pld.FailureReason != nil {
		entry.FailureReason = *pld.FailureReason
	}
	if pld.Highlights != nil {
		entry.Highlights = *pld.Highlights
	}
	if pld.Amount != nil {
		entry.Amount = *pld.Amount
	}

	err = au.UpdatePayoutEntry(db, entry)
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = entry
	return resp.ServerJSON(ctx)
}
