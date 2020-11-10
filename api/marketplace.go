package api

import (
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/data"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/middlewares"
	"github.com/shopicano/shopicano-backend/validators"
	"net/http"
)

func RegisterPlatformRoutes(publicEndpoints, platformEndpoints *echo.Group) {
	func(g echo.Group) {
		g.Use(middlewares.IsPlatformManager)
		g.POST("/shipping-methods/", createShippingMethod)
		g.PUT("/shipping-methods/:id/", updateShippingMethod)
		g.DELETE("/shipping-methods/:id/", deleteShippingMethod)
		g.GET("/shipping-methods/", listShippingMethodsAsAdmin)
		g.GET("/shipping-methods/:id/", getShippingMethod)

		g.POST("/payment-methods/", createPaymentMethod)
		g.PUT("/payment-methods/:id/", updatePaymentMethod)
		g.DELETE("/payment-methods/:id/", deletePaymentMethod)
		g.GET("/payment-methods/", listPaymentMethodsAsAdmin)
		g.GET("/payment-methods/:id/", getPaymentMethod)

		g.POST("/business-account-types/", createBusinessAccountType)
		g.PUT("/business-account-types/:bat_id/", updateBusinessAccountType)
		g.DELETE("/business-account-types/:bat_id/", deleteBusinessAccountType)
		g.GET("/business-account-types/", listBusinessAccountTypes)
		g.GET("/business-account-types/:bat_id/", getBusinessAccountType)

		g.POST("/payout-methods/", createPayoutMethod)
		g.PUT("/payout-methods/:pom_id/", updatePayoutMethod)
		g.DELETE("/payout-methods/:pom_id/", deletePayoutMethod)
		g.GET("/payout-methods/", listPayoutMethods)
		g.GET("/payout-methods/:pom_id/", getPayoutMethod)

		g.GET("/users/", listUsers)
	}(*platformEndpoints)

	func(g echo.Group) {
		g.Use(middlewares.IsPlatformAdmin)
		g.PATCH("/settings/", updateSettings)
	}(*platformEndpoints)

	func(g echo.Group) {
		g.Use(middlewares.JWTAuth())
		g.GET("/payment-methods/:id/", getPaymentMethodForUser)
		g.GET("/shipping-methods/:id/", getShippingMethodForUser)
	}(*publicEndpoints)

	func(g echo.Group) {
		g.Use(middlewares.JWTAuth())
		g.Use(middlewares.HasStore())
		g.Use(middlewares.IsStoreAdmin())
		g.GET("/business-account-types/", listBusinessAccountTypesForUser)
		g.GET("/business-account-types/:bat_id/", getBusinessAccountTypeForUser)

		g.GET("/payout-methods/", listPayoutMethodsForUser)
		g.GET("/payout-methods/:pom_id/", getPayoutMethodForUser)
	}(*publicEndpoints)
}

func updateSettings(ctx echo.Context) error {
	req, err := validators.ValidateUpdateSettings(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.SettingsUpdateDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()

	au := data.NewMarketplaceRepository()

	s, err := au.GetSettings(db)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Settings not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.SettingsNotFound
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
		s.Name = *req.Name
	}
	if req.TagLine != nil {
		s.TagLine = *req.TagLine
	}
	if req.Status != nil {
		s.Status = *req.Status
	}
	if req.CompanyAddressID != nil {
		s.CompanyAddressID = *req.CompanyAddressID
	}
	if req.IsStoreCreationEnabled != nil {
		s.IsStoreCreationEnabled = *req.IsStoreCreationEnabled
	}
	if req.IsSignUpEnabled != nil {
		s.IsSignUpEnabled = *req.IsSignUpEnabled
	}
	if req.DefaultCommissionRate != nil {
		s.DefaultCommissionRate = *req.DefaultCommissionRate
	}
	if req.EnabledAutoStoreConfirmation != nil {
		s.EnabledAutoStoreConfirmation = *req.EnabledAutoStoreConfirmation
	}
	if req.Website != nil {
		s.Website = *req.Website
	}

	err = au.UpdateSettings(db, s)
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = s
	return resp.ServerJSON(ctx)
}
