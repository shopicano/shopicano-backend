package api

import (
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/errors"
	gateway "github.com/shopicano/shopicano-backend/payment-gateways"
	"net/http"
)

func RegisterPaymentRoutes(publicEndpoints, platformEndpoints *echo.Group) {
	paymentsPublicPath := publicEndpoints.Group("/payments")

	paymentsPublicPath.GET("/configs/", getPaymentGatewayConfig)
	paymentsPublicPath.GET("/confirm/", processPayOrderFor2Checkout)
}

func getPaymentGatewayConfig(ctx echo.Context) error {
	resp := core.Response{}

	config, err := gateway.GetActivePaymentGateway().GetConfig()
	if err != nil {
		resp.Title = "Failed to get payment gateway client config"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = config
	return resp.ServerJSON(ctx)
}
