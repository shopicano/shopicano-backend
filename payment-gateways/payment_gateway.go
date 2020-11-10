package payment_gateways

import (
	"errors"
	"github.com/braintree-go/braintree-go"
	"github.com/shopicano/shopicano-backend/config"
	"github.com/shopicano/shopicano-backend/models"
)

type PaymentGateway interface {
	GetName() string
	GetConfig() (map[string]interface{}, error)
	Pay(orderDetails *models.OrderDetailsView) (*PaymentGatewayResponse, error)
	ValidateTransaction(orderDetails *models.OrderDetailsView) error
	VoidTransaction(orderDetails *models.OrderDetailsView, params map[string]interface{}) error
	DisplayName() string
}

type PaymentGatewayResponse struct {
	Result                     string
	Nonce                      string
	BrainTreeTransactionStatus braintree.TransactionStatus
}

var activePaymentGateway PaymentGateway

func SetActivePaymentGateway(cfg config.PaymentGatewayCfg) error {
	gateway, err := GetPaymentGatewayByName(cfg.Name)
	if err != nil {
		return err
	}
	activePaymentGateway = gateway
	return nil
}

func GetActivePaymentGateway() PaymentGateway {
	return activePaymentGateway
}

func GetPaymentGatewayByName(name string) (PaymentGateway, error) {
	cfg := config.PaymentGateway()
	if name == StripePaymentGatewayName {
		stripe, err := NewStripePaymentGateway(cfg.Configs[StripePaymentGatewayName].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		return stripe, nil
	} else if name == BrainTreePaymentGatewayName {
		bt, err := NewBrainTreePaymentGateway(cfg.Configs[BrainTreePaymentGatewayName].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		return bt, nil
	} else if name == TwoCheckoutPaymentGatewayName {
		tco, err := NewTwoCheckoutPaymentGateway(cfg.Configs[TwoCheckoutPaymentGatewayName].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		return tco, nil
	} else if name == SSLCommerzPaymentGatewayName {
		ssl, err := NewSSLCommerzPaymentGateway(cfg.Configs[SSLCommerzPaymentGatewayName].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		return ssl, nil
	} else if name == PaddlePaymentGatewayName {
		pd, err := NewPaddlePaymentGateway(cfg.Configs[PaddlePaymentGatewayName].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		return pd, nil
	}
	return nil, errors.New("payment gateway not found")
}
