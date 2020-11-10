package api

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/config"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/data"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/log"
	"github.com/shopicano/shopicano-backend/models"
	payment_gateways "github.com/shopicano/shopicano-backend/payment-gateways"
	"github.com/shopicano/shopicano-backend/queue"
	"github.com/shopicano/shopicano-backend/utils"
	"io/ioutil"
	"net/http"
	"time"
)

// payOrder is the IPN callback
func payOrder(ctx echo.Context) error {
	orderID := ctx.Param("order_id")

	resp := core.Response{}

	db := app.DB()

	ou := data.NewOrderRepository()
	m, err := ou.GetDetails(db, orderID)
	if err != nil {
		resp.Title = "Order not found"
		resp.Status = http.StatusNotFound
		resp.Code = errors.OrderNotFound
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if m.Status == models.OrderCancelled {
		resp.Title = "Order already cancelled"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.OrderAlreadyCancelled
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if m.PaymentStatus == models.PaymentCompleted {
		resp.Title = "Order already paid"
		resp.Status = http.StatusConflict
		resp.Code = errors.PaymentAlreadyProcessed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if m.PaymentStatus == models.PaymentReverted {
		resp.Title = "Order payment already reverted"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.OrderPaymentAlreadyReverted
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	switch m.PaymentGateway {
	case payment_gateways.BrainTreePaymentGatewayName:
		return processPayOrderForBrainTree(ctx, m)
	case payment_gateways.StripePaymentGatewayName:
		return processPayOrderForStripe(ctx, m)
	case payment_gateways.SSLCommerzPaymentGatewayName:
		return processPayOrderForSSL(ctx, m)
	case payment_gateways.PaddlePaymentGatewayName:
		return processPayOrderForPaddle(ctx, m)
	}
	return serveInvalidPaymentRequest(ctx)
}

type reqBrainTreeNonce struct {
	Nonce *string `json:"nonce"`
}

func processPayOrderForBrainTree(ctx echo.Context, o *models.OrderDetailsView) error {
	resp := core.Response{}

	db := app.DB().Begin()
	or := data.NewOrderRepository()

	body := reqBrainTreeNonce{}
	if err := ctx.Bind(&body); err != nil {
		db.Rollback()

		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.OrderPaymentDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	o.Nonce = body.Nonce

	pg, err := payment_gateways.GetPaymentGatewayByName(o.PaymentGateway)
	if err != nil {
		db.Rollback()

		resp.Title = "Invalid payment gateway"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.PaymentProcessingFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	res, err := pg.Pay(o)
	if err != nil {
		db.Rollback()

		resp.Title = "Failed to process payment"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.PaymentProcessingFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	o.TransactionID = &res.Result

	if err := pg.ValidateTransaction(o); err != nil {
		log.Log().Errorln(err)

		o.PaymentStatus = models.PaymentFailed
	} else {
		o.PaymentStatus = models.PaymentCompleted
	}

	if err := or.UpdatePaymentInfo(db, o); err != nil {
		db.Rollback()

		resp.Title = "Failed to update payment info"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	ol := models.OrderLog{
		ID:        utils.NewUUID(),
		OrderID:   o.ID,
		Action:    string(o.PaymentStatus),
		Details:   "Payment has been updated using BrainTree",
		CreatedAt: time.Now(),
	}
	if err := or.CreateLog(db, &ol); err != nil {
		db.Rollback()

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if o.PaymentStatus == models.PaymentCompleted {
		if err := queue.SendPaymentConfirmationEmail(o.ID); err != nil {
			db.Rollback()

			resp.Title = "Failed to enqueue task"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.FailedToEnqueueTask
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
	resp.Data = map[string]interface{}{
		"transaction_id": res.Result,
	}
	return resp.ServerJSON(ctx)
}

func processPayOrderForStripe(ctx echo.Context, o *models.OrderDetailsView) error {
	resp := core.Response{}

	db := app.DB().Begin()
	or := data.NewOrderRepository()

	pg, err := payment_gateways.GetPaymentGatewayByName(o.PaymentGateway)
	if err != nil {
		db.Rollback()

		resp.Title = "Invalid payment gateway"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.PaymentGatewayFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if err := pg.ValidateTransaction(o); err != nil {
		log.Log().Errorln(err)

		o.PaymentStatus = models.PaymentFailed
	} else {
		o.PaymentStatus = models.PaymentCompleted
	}

	if err := or.UpdatePaymentInfo(db, o); err != nil {
		db.Rollback()

		resp.Title = "Failed to update payment info"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	ol := models.OrderLog{
		ID:        utils.NewUUID(),
		OrderID:   o.ID,
		Action:    string(o.PaymentStatus),
		Details:   "Payment has been updated using Stripe",
		CreatedAt: time.Now(),
	}
	if err := or.CreateLog(db, &ol); err != nil {
		db.Rollback()

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if o.PaymentStatus == models.PaymentCompleted {
		if err := queue.SendPaymentConfirmationEmail(o.ID); err != nil {
			db.Rollback()

			resp.Title = "Failed to enqueue task"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.FailedToEnqueueTask
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

	orderPath := fmt.Sprintf(config.PathMappingCfg()["after_payment_completed"], o.ID)
	paymentCompletedCallback := fmt.Sprintf("%s%s", config.App().FrontStoreUrl, orderPath)
	return ctx.Redirect(http.StatusPermanentRedirect, paymentCompletedCallback)
}

func processPayOrderFor2Checkout(ctx echo.Context) error {
	orderID := ctx.QueryParam("merchant_order_id")

	b, _ := ioutil.ReadAll(ctx.Request().Body)
	log.Log().Infoln(string(b))

	resp := core.Response{}

	db := app.DB().Begin()

	ou := data.NewOrderRepository()
	m, err := ou.GetDetails(db, orderID)
	if err != nil {
		db.Rollback()

		resp.Title = "Order not found"
		resp.Status = http.StatusNotFound
		resp.Code = errors.OrderNotFound
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if m.PaymentStatus == models.PaymentCompleted {
		db.Rollback()

		resp.Title = "Order already paid"
		resp.Status = http.StatusConflict
		resp.Code = errors.PaymentAlreadyProcessed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if m.PaymentGateway != payment_gateways.TwoCheckoutPaymentGatewayName {
		db.Rollback()
		return serveInvalidPaymentRequest(ctx)
	}

	pg, err := payment_gateways.GetPaymentGatewayByName(m.PaymentGateway)
	if err != nil {
		db.Rollback()

		return serveInvalidPaymentRequest(ctx)
	}

	trx := ctx.QueryParam("invoice_id")
	m.TransactionID = &trx

	if err := pg.ValidateTransaction(m); err != nil {
		log.Log().Errorln(err)

		m.PaymentStatus = models.PaymentFailed
	} else {
		m.PaymentStatus = models.PaymentCompleted
	}

	or := data.NewOrderRepository()

	if err := or.UpdatePaymentInfo(db, m); err != nil {
		db.Rollback()

		resp.Title = "Failed to update payment info"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	ol := models.OrderLog{
		ID:        utils.NewUUID(),
		OrderID:   m.ID,
		Action:    string(m.PaymentStatus),
		Details:   "Payment has been updated using 2Checkout",
		CreatedAt: time.Now(),
	}
	if err := or.CreateLog(db, &ol); err != nil {
		db.Rollback()

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if m.PaymentStatus == models.PaymentCompleted {
		if err := queue.SendPaymentConfirmationEmail(m.ID); err != nil {
			db.Rollback()

			resp.Title = "Failed to enqueue task"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.FailedToEnqueueTask
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

	orderPath := fmt.Sprintf(config.PathMappingCfg()["after_payment_completed"], m.ID)
	paymentCompletedCallback := fmt.Sprintf("%s%s", config.App().FrontStoreUrl, orderPath)
	return ctx.Redirect(http.StatusPermanentRedirect, paymentCompletedCallback)
}

func processPayOrderForSSL(ctx echo.Context, m *models.OrderDetailsView) error {
	resp := core.Response{}

	db := app.DB().Begin()

	if m.PaymentGateway != payment_gateways.SSLCommerzPaymentGatewayName {
		db.Rollback()
		return serveInvalidPaymentRequest(ctx)
	}

	pg, err := payment_gateways.GetPaymentGatewayByName(m.PaymentGateway)
	if err != nil {
		db.Rollback()

		return serveInvalidPaymentRequest(ctx)
	}

	if err := pg.ValidateTransaction(m); err != nil {
		log.Log().Errorln(err)

		m.PaymentStatus = models.PaymentFailed
	} else {
		m.PaymentStatus = models.PaymentCompleted
	}

	or := data.NewOrderRepository()

	if err := or.UpdatePaymentInfo(db, m); err != nil {
		db.Rollback()

		resp.Title = "Failed to update payment info"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	ol := models.OrderLog{
		ID:        utils.NewUUID(),
		OrderID:   m.ID,
		Action:    string(m.PaymentStatus),
		Details:   "Payment has been updated using SSL",
		CreatedAt: time.Now(),
	}
	if err := or.CreateLog(db, &ol); err != nil {
		db.Rollback()

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if m.PaymentStatus == models.PaymentCompleted {
		if err := queue.SendPaymentConfirmationEmail(m.ID); err != nil {
			db.Rollback()

			resp.Title = "Failed to enqueue task"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.FailedToEnqueueTask
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

	orderPath := fmt.Sprintf(config.PathMappingCfg()["after_payment_completed"], m.ID)
	paymentCompletedCallback := fmt.Sprintf("%s%s", config.App().FrontStoreUrl, orderPath)
	return ctx.Redirect(http.StatusPermanentRedirect, paymentCompletedCallback)
}

func processPayOrderForPaddle(ctx echo.Context, m *models.OrderDetailsView) error {
	resp := core.Response{}

	db := app.DB().Begin()

	if m.PaymentGateway != payment_gateways.PaddlePaymentGatewayName {
		db.Rollback()
		return serveInvalidPaymentRequest(ctx)
	}

	pg, err := payment_gateways.GetPaymentGatewayByName(m.PaymentGateway)
	if err != nil {
		db.Rollback()

		return serveInvalidPaymentRequest(ctx)
	}

	if err := ctx.Request().ParseForm(); err != nil {
		db.Rollback()

		resp.Status = http.StatusBadRequest
		resp.Title = "Failed to parse request body"
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	transactionID := ctx.Request().FormValue("p_order_id")
	m.TransactionID = &transactionID

	if err := pg.ValidateTransaction(m); err != nil {
		log.Log().Errorln(err)

		m.PaymentStatus = models.PaymentFailed
	} else {
		m.PaymentStatus = models.PaymentCompleted
	}

	or := data.NewOrderRepository()

	if err := or.UpdatePaymentInfo(db, m); err != nil {
		db.Rollback()

		resp.Title = "Failed to update payment info"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	ol := models.OrderLog{
		ID:        utils.NewUUID(),
		OrderID:   m.ID,
		Action:    string(m.PaymentStatus),
		Details:   fmt.Sprintf("Payment has been updated using %s", pg.DisplayName()),
		CreatedAt: time.Now(),
	}
	if err := or.CreateLog(db, &ol); err != nil {
		db.Rollback()

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if m.PaymentStatus == models.PaymentCompleted {
		if err := queue.SendPaymentConfirmationEmail(m.ID); err != nil {
			db.Rollback()

			resp.Title = "Failed to enqueue task"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.FailedToEnqueueTask
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

	return ctx.JSON(http.StatusOK, nil)
}

// generatePayNonce create payment reference / nonce
func generatePayNonce(ctx echo.Context) error {
	orderID := ctx.Param("order_id")

	resp := core.Response{}

	db := app.DB()

	ou := data.NewOrderRepository()
	m, err := ou.GetDetails(db, orderID)
	if err != nil {
		resp.Title = "Order not found"
		resp.Status = http.StatusNotFound
		resp.Code = errors.OrderNotFound
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if m.PaymentStatus == models.PaymentCompleted {
		resp.Title = "Order already paid"
		resp.Status = http.StatusConflict
		resp.Code = errors.PaymentAlreadyProcessed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	switch m.PaymentGateway {
	case payment_gateways.StripePaymentGatewayName:
		return generateStripePayNonce(ctx, m)
	case payment_gateways.TwoCheckoutPaymentGatewayName:
		return generate2CheckoutPayUrl(ctx, m)
	case payment_gateways.SSLCommerzPaymentGatewayName:
		return generateSSLPayUrl(ctx, m)
	case payment_gateways.PaddlePaymentGatewayName:
		return generatePaddlePayUrl(ctx, m)
	}
	return serveInvalidPaymentRequest(ctx)
}

func generateStripePayNonce(ctx echo.Context, o *models.OrderDetailsView) error {
	resp := core.Response{}

	db := app.DB()
	or := data.NewOrderRepository()

	pg, err := payment_gateways.GetPaymentGatewayByName(o.PaymentGateway)
	if err != nil {
		resp.Title = "Invalid payment gateway"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.PaymentProcessingFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	res, err := pg.Pay(o)
	if err != nil {
		resp.Title = "Failed to process payment"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.PaymentProcessingFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	o.TransactionID = &res.Result
	o.Nonce = &res.Nonce

	if err := or.UpdatePaymentInfo(db, o); err != nil {
		resp.Title = "Failed to update payment info"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = map[string]interface{}{
		"nonce": res.Nonce,
	}
	return resp.ServerJSON(ctx)
}

func generate2CheckoutPayUrl(ctx echo.Context, o *models.OrderDetailsView) error {
	resp := core.Response{}

	pg, err := payment_gateways.GetPaymentGatewayByName(o.PaymentGateway)
	if err != nil {
		resp.Title = "Invalid payment gateway"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.PaymentProcessingFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	res, err := pg.Pay(o)
	if err != nil {
		resp.Title = "Failed to process payment"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.PaymentProcessingFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	url := res.Result
	resp.Status = http.StatusOK
	resp.Data = map[string]interface{}{
		"url": url,
	}
	return resp.ServerJSON(ctx)
}

func generateSSLPayUrl(ctx echo.Context, o *models.OrderDetailsView) error {
	resp := core.Response{}

	pg, err := payment_gateways.GetPaymentGatewayByName(o.PaymentGateway)
	if err != nil {
		resp.Title = "Invalid payment gateway"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.PaymentProcessingFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	res, err := pg.Pay(o)
	if err != nil {
		log.Log().Infoln(err)

		resp.Title = "Failed to process payment"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.PaymentProcessingFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()
	or := data.NewOrderRepository()

	o.TransactionID = &res.Result
	o.Nonce = &res.Nonce

	if err := or.UpdatePaymentInfo(db, o); err != nil {
		resp.Title = "Failed to update payment info"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	url := res.Nonce
	resp.Status = http.StatusOK
	resp.Data = map[string]interface{}{
		"url": url,
	}
	return resp.ServerJSON(ctx)
}

func generatePaddlePayUrl(ctx echo.Context, o *models.OrderDetailsView) error {
	resp := core.Response{}

	pg, err := payment_gateways.GetPaymentGatewayByName(o.PaymentGateway)
	if err != nil {
		resp.Title = "Invalid payment gateway"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.PaymentProcessingFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	res, err := pg.Pay(o)
	if err != nil {
		log.Log().Infoln(err)

		resp.Title = "Failed to process payment"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.PaymentProcessingFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	url := res.Result
	resp.Status = http.StatusOK
	resp.Data = map[string]interface{}{
		"url": url,
	}
	return resp.ServerJSON(ctx)
}

func serveInvalidPaymentRequest(ctx echo.Context) error {
	resp := core.Response{}
	resp.Title = "Invalid payment request"
	resp.Status = http.StatusForbidden
	resp.Code = errors.PaymentProcessingFailed
	return resp.ServerJSON(ctx)
}

func revertOrderPayment(ctx echo.Context) error {
	orderID := ctx.Param("order_id")

	resp := core.Response{}

	db := app.DB()

	ou := data.NewOrderRepository()
	m, err := ou.GetDetails(db, orderID)
	if err != nil {
		resp.Title = "Order not found"
		resp.Status = http.StatusNotFound
		resp.Code = errors.OrderNotFound
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if m.PaymentStatus == models.PaymentReverted {
		resp.Title = "Order payment already reverted"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.OrderPaymentAlreadyReverted
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if m.PaymentStatus != models.PaymentCompleted {
		resp.Title = "Order not paid yet"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.OrderNotPaidYet
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	switch m.PaymentGateway {
	case payment_gateways.StripePaymentGatewayName:
		return revertOrderPaymentForAny(ctx, m)
	case payment_gateways.BrainTreePaymentGatewayName:
		return revertOrderPaymentForAny(ctx, m)
	case payment_gateways.TwoCheckoutPaymentGatewayName:
		return revertOrderPaymentForAny(ctx, m)
	case payment_gateways.SSLCommerzPaymentGatewayName:
		return revertOrderPaymentForAny(ctx, m)
	}
	return serveInvalidPaymentRequest(ctx)
}

type reqRevertPayment struct {
	Reason string `json:"reason"`
	Type   int    `json:"type"`
}

func revertOrderPaymentForAny(ctx echo.Context, details *models.OrderDetailsView) error {
	resp := core.Response{}

	body := reqRevertPayment{}
	if err := ctx.Bind(&body); err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.OrderPaymentRevertDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB().Begin()
	or := data.NewOrderRepository()

	pg, err := payment_gateways.GetPaymentGatewayByName(details.PaymentGateway)
	if err != nil {
		db.Rollback()

		resp.Title = "Invalid payment gateway"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.PaymentGatewayFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if err := pg.VoidTransaction(details, map[string]interface{}{
		"reason": body.Reason,
		"type":   body.Type,
	}); err != nil {
		db.Rollback()

		log.Log().Errorln(err)

		resp.Title = "Failed to revert payment"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.PaymentGatewayFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	} else {
		details.PaymentStatus = models.PaymentReverted
	}

	if err := or.UpdatePaymentInfo(db, details); err != nil {
		db.Rollback()

		resp.Title = "Failed to update payment info"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	ol := models.OrderLog{
		ID:        utils.NewUUID(),
		OrderID:   details.ID,
		Action:    string(details.PaymentStatus),
		Details:   fmt.Sprintf("Payment reverted for : %s", body.Reason),
		CreatedAt: time.Now().UTC(),
	}
	if err := or.CreateLog(db, &ol); err != nil {
		db.Rollback()

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if err := queue.SendPaymentRevertedEmail(details.ID); err != nil {
		db.Rollback()

		resp.Title = "Failed to enqueue task"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.FailedToEnqueueTask
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

	resp.Status = http.StatusOK
	resp.Title = "Payment successfully reverted"
	return resp.ServerJSON(ctx)
}
