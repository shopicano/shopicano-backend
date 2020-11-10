package payment_gateways

import (
	"context"
	"errors"
	"fmt"
	"github.com/braintree-go/braintree-go"
	"github.com/shopicano/shopicano-backend/log"
	"github.com/shopicano/shopicano-backend/models"
)

const (
	BrainTreePaymentGatewayName = "brain_tree"
)

const (
	Sale BrainTreeTransactionType = "sale"
)

type BrainTreeTransactionType string

type brainTreePaymentGateway struct {
	SuccessCallback string
	FailureCallback string
	Token           string
	PublicKey       string
	PrivateKey      string
	client          *braintree.Braintree
}

func NewBrainTreePaymentGateway(cfg map[string]interface{}) (*brainTreePaymentGateway, error) {
	publicKey := cfg["public_key"].(string)
	privateKey := cfg["private_key"].(string)
	merchantID := cfg["merchant_id"].(string)
	mode := cfg["mode"].(string)

	gatewayMode := braintree.Sandbox
	if mode == "live" {
		gatewayMode = braintree.Production
	}

	c := braintree.New(gatewayMode, merchantID, publicKey, privateKey)

	return &brainTreePaymentGateway{
		client:          c,
		SuccessCallback: cfg["success_callback"].(string),
		FailureCallback: cfg["failure_callback"].(string),
		Token:           cfg["token"].(string),
		PublicKey:       publicKey,
		PrivateKey:      privateKey,
	}, nil
}

func (bt *brainTreePaymentGateway) GetName() string {
	return BrainTreePaymentGatewayName
}

func (bt *brainTreePaymentGateway) Pay(orderDetails *models.OrderDetailsView) (*PaymentGatewayResponse, error) {
	var items []*braintree.TransactionLineItemRequest

	grandTotalU := orderDetails.GrandTotal / 100
	grandTotalS := int(orderDetails.GrandTotal % 100)

	log.Log().Infoln("Grand Total U : ", grandTotalU)
	log.Log().Infoln("Grand Total S : ", grandTotalS)

	d := braintree.NewDecimal(0, 0)
	if err := d.UnmarshalText([]byte(fmt.Sprintf("%d.%d", grandTotalU, grandTotalS))); err != nil {
		return nil, err
	}

	log.Log().Infoln(d.String())

	items = append(items, &braintree.TransactionLineItemRequest{
		Name:        fmt.Sprintf("Payment for Order #%s", orderDetails.Hash),
		UnitAmount:  d,
		Quantity:    braintree.NewDecimal(int64(1), 0),
		TotalAmount: d,
		Kind:        braintree.TransactionLineItemKindDebit,
	})

	resp, err := bt.client.Transaction().Create(context.Background(), &braintree.TransactionRequest{
		PaymentMethodNonce: *orderDetails.Nonce,
		Amount:             d,
		LineItems:          items,
		BillingAddress: &braintree.Address{
			StreetAddress: fmt.Sprintf("%s", orderDetails.BillingAddress),
			Region:        orderDetails.BillingCity,
			PostalCode:    orderDetails.BillingPostcode,
			CountryName:   orderDetails.BillingCountry,
		},
		Options: &braintree.TransactionOptions{
			SubmitForSettlement: true,
		},
		Type: string(Sale),
	})

	if err != nil {
		log.Log().Errorln(err)
		return nil, err
	}

	return &PaymentGatewayResponse{
		Result:                     resp.Id,
		BrainTreeTransactionStatus: resp.Status,
	}, nil
}

func (bt *brainTreePaymentGateway) GetConfig() (map[string]interface{}, error) {
	token, err := bt.client.ClientToken().Generate(context.Background())
	if err != nil {
		return nil, err
	}

	cfg := map[string]interface{}{
		"client_token":         token,
		"success_callback_url": bt.SuccessCallback,
		"failure_callback_url": bt.FailureCallback,
		"public_key":           bt.PublicKey,
	}

	return cfg, nil
}

func (bt *brainTreePaymentGateway) ValidateTransaction(orderDetails *models.OrderDetailsView) error {
	if orderDetails.TransactionID == nil {
		return errors.New("invalid transactionID")
	}

	transaction, err := bt.client.Transaction().Find(context.Background(), *orderDetails.TransactionID)
	if err != nil {
		log.Log().Errorln(err)
		return err
	}

	log.Log().Infoln(transaction.Status)
	log.Log().Infoln(transaction.Amount)
	log.Log().Infoln(transaction.ServiceFeeAmount)

	if transaction.Status != braintree.TransactionStatusSettled && transaction.Status != braintree.TransactionStatusSubmittedForSettlement {
		return errors.New("transaction isn't settled yet")
	}

	log.Log().Infoln("Grand Total : ", orderDetails.GrandTotal)
	log.Log().Infoln("Unscaled : ", transaction.Amount.Unscaled)
	log.Log().Infoln("Scaled : ", transaction.Amount.Scale)

	if transaction.Amount.Unscaled != orderDetails.GrandTotal {
		return errors.New("invalid transaction amount")
	}

	return nil
}

func (bt *brainTreePaymentGateway) VoidTransaction(orderDetails *models.OrderDetailsView, params map[string]interface{}) error {
	if orderDetails.TransactionID == nil {
		return errors.New("invalid transactionID")
	}

	refundedAmount := orderDetails.GrandTotal - orderDetails.PaymentProcessingFee
	grandTotalU := refundedAmount / 100
	grandTotalS := int(refundedAmount % 100)

	log.Log().Infoln("Grand Total U : ", grandTotalU)
	log.Log().Infoln("Grand Total S : ", grandTotalS)

	d := braintree.NewDecimal(0, 0)
	if err := d.UnmarshalText([]byte(fmt.Sprintf("%d.%d", grandTotalU, grandTotalS))); err != nil {
		return err
	}

	if _, err := bt.client.Transaction().
		Refund(context.Background(), *orderDetails.TransactionID, d); err != nil {
		log.Log().Errorln(err)
		return err
	}
	return nil
}

func (bt *brainTreePaymentGateway) DisplayName() string {
	return "BrainTree"
}
