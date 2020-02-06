package payment_gateways

import (
	"fmt"
	"github.com/shopicano/shopicano-backend/models"
)

const (
	TwoCheckoutPaymentGatewayName = "2co"
)

type twoCheckoutPaymentGateway struct {
	SuccessCallback string
	FailureCallback string
	PublicKey       string
	PrivateKey      string
	MerchantCode    string
	SecretKey       string
}

func NewTwoCheckoutPaymentGateway(cfg map[string]interface{}) (*twoCheckoutPaymentGateway, error) {
	publicKey := cfg["public_key"].(string)
	privateKey := cfg["private_key"].(string)
	merchantCode := cfg["merchant_code"].(string)
	secretKey := cfg["secret_key"].(string)

	return &twoCheckoutPaymentGateway{
		SuccessCallback: cfg["success_callback"].(string),
		FailureCallback: cfg["failure_callback"].(string),
		PublicKey:       publicKey,
		PrivateKey:      privateKey,
		MerchantCode:    merchantCode,
		SecretKey:       secretKey,
	}, nil
}

func (tco *twoCheckoutPaymentGateway) GetName() string {
	return TwoCheckoutPaymentGatewayName
}

func (tco *twoCheckoutPaymentGateway) Pay(orderDetails *models.OrderDetailsView) (*PaymentGatewayResponse, error) {
	url := "https://sandbox.2checkout.com/checkout/purchase"

	payload := fmt.Sprintf("sid=%s&", tco.MerchantCode)
	payload += fmt.Sprintf("mode=%s&", "2CO")
	payload += fmt.Sprintf("submit=%s&", "Checkout")
	payload += fmt.Sprintf("merchant_order_id=%s&", orderDetails.ID)
	payload += fmt.Sprintf("currency_code=%s&", "USD")
	payload += fmt.Sprintf("street_address=%s&", orderDetails.BillingAddress)
	payload += fmt.Sprintf("city=%s&", orderDetails.BillingCity)
	payload += fmt.Sprintf("state=%s&", orderDetails.BillingCity)
	payload += fmt.Sprintf("zip=%s&", orderDetails.BillingPostcode)
	payload += fmt.Sprintf("country=%s&", orderDetails.BillingCountry)
	payload += fmt.Sprintf("phone=%s&", orderDetails.BillingPhone)
	payload += fmt.Sprintf("email=%s&", orderDetails.BillingEmail)

	for i, op := range orderDetails.Items {
		payload += fmt.Sprintf("li_%d_type=%s&", i, "product")
		payload += fmt.Sprintf("li_%d_name=%s&", i, op.Name)
		payload += fmt.Sprintf("li_%d_price=%s&", i, fmt.Sprintf("%.2f", float64(op.Price)))
		payload += fmt.Sprintf("li_%d_quantity=%s&", i, fmt.Sprintf("%d", op.Quantity))
		payload += fmt.Sprintf("li_%d_tangible=%s&", i, "N")
	}

	if orderDetails.PaymentProcessingFee != 0 {
		payload += fmt.Sprintf("li_%d_type=%s&", len(orderDetails.Items), "product")
		payload += fmt.Sprintf("li_%d_name=%s&", len(orderDetails.Items), "Payment Processing Fee")
		payload += fmt.Sprintf("li_%d_price=%s&", len(orderDetails.Items), fmt.Sprintf("%.2f", float64(orderDetails.PaymentProcessingFee)))
		payload += fmt.Sprintf("li_%d_quantity=%s&", len(orderDetails.Items), fmt.Sprintf("%d", 1))
		payload += fmt.Sprintf("li_%d_tangible=%s&", len(orderDetails.Items), "N")
	}

	payload += "purchase_step=payment-method"

	return &PaymentGatewayResponse{
		Result: fmt.Sprintf("%s?%s", url, payload),
	}, nil
}

func (tco *twoCheckoutPaymentGateway) GetConfig() (map[string]interface{}, error) {
	cfg := map[string]interface{}{
		"success_callback_url": tco.SuccessCallback,
		"failure_callback_url": tco.FailureCallback,
		"public_key":           tco.PublicKey,
	}
	return cfg, nil
}