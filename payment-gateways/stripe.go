package payment_gateways

import (
	"errors"
	"fmt"
	"github.com/shopicano/shopicano-backend/log"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/checkout/session"
	"github.com/stripe/stripe-go/client"
)

const (
	StripePaymentGatewayName = "stripe"
)

type stripePaymentGateway struct {
	SecretKey       string
	SuccessCallback string
	FailureCallback string
	PublicKey       string
	client          *client.API
}

func NewStripePaymentGateway(cfg map[string]interface{}) (*stripePaymentGateway, error) {
	return &stripePaymentGateway{
		SecretKey:       cfg["secret_key"].(string),
		SuccessCallback: cfg["success_callback"].(string),
		FailureCallback: cfg["failure_callback"].(string),
		PublicKey:       cfg["public_key"].(string),
		client:          client.New(cfg["secret_key"].(string), nil),
	}, nil
}

func (spg *stripePaymentGateway) GetName() string {
	return StripePaymentGatewayName
}

func (spg *stripePaymentGateway) Pay(orderDetails *models.OrderDetailsView) (*PaymentGatewayResponse, error) {
	stripe.Key = spg.SecretKey

	var lineItems []*stripe.CheckoutSessionLineItemParams

	lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
		Name:     stripe.String(fmt.Sprintf("Payment for Order #%s", orderDetails.Hash)),
		Amount:   stripe.Int64(orderDetails.GrandTotal),
		Currency: stripe.String("usd"),
		Quantity: stripe.Int64(int64(1)),
	})

	successUrl := fmt.Sprintf("%s?session_id={CHECKOUT_SESSION_ID}&status=success", spg.SuccessCallback)
	failureUrl := fmt.Sprintf("%s?session_id={CHECKOUT_SESSION_ID}&status=failure", spg.FailureCallback)

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems:          lineItems,
		Params: stripe.Params{
			IdempotencyKey: stripe.String(stripe.NewIdempotencyKey()),
		},
		CustomerEmail:     stripe.String(orderDetails.BillingEmail),
		SuccessURL:        stripe.String(fmt.Sprintf(successUrl, orderDetails.ID)),
		CancelURL:         stripe.String(fmt.Sprintf(failureUrl, orderDetails.ID)),
		ClientReferenceID: stripe.String(orderDetails.ID),
	}

	ss, err := session.New(params)
	if err != nil {
		return nil, err
	}

	return &PaymentGatewayResponse{
		Result: ss.PaymentIntent.ID,
		Nonce:  ss.ID,
	}, nil
}

func (spg *stripePaymentGateway) GetConfig() (map[string]interface{}, error) {
	cfg := map[string]interface{}{
		"success_callback_url": spg.SuccessCallback,
		"failure_callback_url": spg.FailureCallback,
		"public_key":           spg.PublicKey,
	}

	return cfg, nil
}

func (spg *stripePaymentGateway) ValidateTransaction(orderDetails *models.OrderDetailsView) error {
	if orderDetails.TransactionID == nil {
		return errors.New("invalid transactionID")
	}

	result, err := spg.client.PaymentIntents.Get(*orderDetails.TransactionID, &stripe.PaymentIntentParams{})
	if err != nil {
		return err
	}

	if result.Status != stripe.PaymentIntentStatusSucceeded {
		return errors.New("payment intent status isn't succeed")
	}

	capturedAmount := int64(0)

	for _, c := range result.Charges.Data {
		_, err := spg.client.Charges.Capture(c.ID, &stripe.CaptureParams{})
		formattedErr := err.(*stripe.Error)
		if err != nil && formattedErr.Code != stripe.ErrorCodeChargeAlreadyCaptured {
			log.Log().Infoln(formattedErr)
			return err
		}

		capturedAmount += c.Amount
	}

	log.Log().Infoln("Captured : ", capturedAmount)
	log.Log().Infoln("Grand Total : ", orderDetails.GrandTotal)

	if capturedAmount != orderDetails.GrandTotal {
		return errors.New("paid amount is invalid")
	}
	return nil
}

func (spg *stripePaymentGateway) VoidTransaction(orderDetails *models.OrderDetailsView, params map[string]interface{}) error {
	if orderDetails.TransactionID == nil {
		return errors.New("invalid transactionID")
	}

	result, err := spg.client.PaymentIntents.Get(*orderDetails.TransactionID, &stripe.PaymentIntentParams{})
	if err != nil {
		return err
	}

	if result.Status != stripe.PaymentIntentStatusSucceeded {
		return errors.New("payment isn't paid yet")
	}

	for _, c := range result.Charges.Data {
		typ := params["type"].(int)
		reason := stripe.RefundReasonRequestedByCustomer

		switch typ {
		case 1:
			reason = stripe.RefundReasonDuplicate
		case 2:
			reason = stripe.RefundReasonFraudulent
		}

		refundAmount := orderDetails.GrandTotal - orderDetails.PaymentProcessingFee
		_, err := spg.client.Refunds.New(&stripe.RefundParams{
			Amount:               stripe.Int64(refundAmount),
			Reason:               stripe.String(string(reason)),
			Charge:               stripe.String(c.ID),
			ReverseTransfer:      stripe.Bool(false),
			RefundApplicationFee: stripe.Bool(false),
		})
		if err != nil {
			return errors.New("failed to issue refund")
		}
	}

	return nil
}

func (spg *stripePaymentGateway) DisplayName() string {
	return "Stripe"
}
