package payment_gateways

import (
	"fmt"
	"github.com/nahid/gohttp"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/log"
	"github.com/shopicano/shopicano-backend/models"
	"net/http"
	url2 "net/url"
	"strconv"
)

const (
	SSLCommerzPaymentGatewayName = "ssl"
)

type sslCommerzPaymentGateway struct {
	Host            string
	SuccessCallback string
	FailureCallback string
	CancelCallback  string
	StoreID         string
	StorePassword   string
}

func NewSSLCommerzPaymentGateway(cfg map[string]interface{}) (*sslCommerzPaymentGateway, error) {
	return &sslCommerzPaymentGateway{
		SuccessCallback: cfg["success_callback"].(string),
		FailureCallback: cfg["failure_callback"].(string),
		CancelCallback:  cfg["failure_callback"].(string),
		StoreID:         cfg["store_id"].(string),
		StorePassword:   cfg["store_password"].(string),
		Host:            cfg["host"].(string),
	}, nil
}

func (ssl *sslCommerzPaymentGateway) GetName() string {
	return SSLCommerzPaymentGatewayName
}

func (ssl *sslCommerzPaymentGateway) Pay(orderDetails *models.OrderDetailsView) (*PaymentGatewayResponse, error) {
	url := fmt.Sprintf("%s/gwprocess/v4/api.php", ssl.Host)

	payload := fmt.Sprintf("store_id=%s&", ssl.StoreID)
	payload += fmt.Sprintf("store_passwd=%s&", ssl.StorePassword)
	payload += fmt.Sprintf("tran_id=%s&", orderDetails.ID)
	payload += fmt.Sprintf("currency=%s&", "BDT")
	payload += fmt.Sprintf("product_profile=%s&", "general")
	payload += fmt.Sprintf("cus_add1=%s&", orderDetails.BillingAddress)
	payload += fmt.Sprintf("cus_city=%s&", orderDetails.BillingCity)
	payload += fmt.Sprintf("cus_postcode=%s&", orderDetails.BillingPostcode)
	payload += fmt.Sprintf("cus_country=%s&", orderDetails.BillingCountry)
	payload += fmt.Sprintf("cus_phone=%s&", orderDetails.BillingPhone)
	payload += fmt.Sprintf("cus_email=%s&", orderDetails.BillingEmail)
	payload += fmt.Sprintf("cus_name=%s&", orderDetails.BillingName)
	payload += fmt.Sprintf("shipping_method=%s&", "no")
	payload += fmt.Sprintf("success_url=%s&", fmt.Sprintf(ssl.SuccessCallback, orderDetails.ID))
	payload += fmt.Sprintf("fail_url=%s&", fmt.Sprintf(ssl.FailureCallback, orderDetails.ID))
	payload += fmt.Sprintf("cancel_url=%s&", fmt.Sprintf(ssl.FailureCallback, orderDetails.ID))

	grandTotal := float64(orderDetails.GrandTotal) / 100

	payload += fmt.Sprintf("total_amount=%f&", grandTotal)

	log.Log().Infoln("Grand Total : ", grandTotal)

	payload += fmt.Sprintf("product_category=%s&", "general")
	payload += fmt.Sprintf("product_name=%s&", fmt.Sprintf("Payment for Order %s", orderDetails.Hash))

	resp, err := gohttp.NewRequest().Body([]byte(payload)).Headers(map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}).Post(url)

	if err != nil {
		log.Log().Errorln(err)
		return nil, err
	}

	if !(resp.GetStatusCode() >= 200 && resp.GetStatusCode() < 300) {
		return nil, errors.NewError(fmt.Sprintf("Request failed with status code : %d", resp.GetStatusCode()))
	}

	result := map[string]interface{}{}
	if err := resp.UnmarshalBody(&result); err != nil {
		return nil, err
	}

	log.Log().Infoln(result)

	if result["status"].(string) != "SUCCESS" {
		return nil, errors.NewError("Payment gateway failed")
	}

	return &PaymentGatewayResponse{
		Nonce:  result["GatewayPageURL"].(string),
		Result: result["sessionkey"].(string),
	}, nil
}

func (ssl *sslCommerzPaymentGateway) GetConfig() (map[string]interface{}, error) {
	cfg := map[string]interface{}{
		"success_callback_url": ssl.SuccessCallback,
		"failure_callback_url": ssl.FailureCallback,
	}
	return cfg, nil
}

type resSSLValidateTransaction struct {
	Status            string `json:"status"`
	SessionKey        string `json:"sessionKey"`
	TranID            string `json:"tran_id"`
	Amount            string `json:"amount"`
	BankTransactionID string `json:"bank_tran_id"`
}

func (ssl *sslCommerzPaymentGateway) ValidateTransaction(orderDetails *models.OrderDetailsView) error {
	if orderDetails.TransactionID == nil {
		return errors.NewError("invalid transactionID")
	}

	url := fmt.Sprintf("%s/validator/api/merchantTransIDvalidationAPI.php?sessionkey=%s&store_id=%s&store_passwd=%s&format=json",
		ssl.Host, *orderDetails.TransactionID, ssl.StoreID, ssl.StorePassword)

	req := gohttp.NewRequest().
		Headers(map[string]string{
			"Accept": "application/json",
		})

	resp, err := req.Post(url)
	if err != nil {
		return err
	}

	if resp.GetStatusCode() != http.StatusOK {
		return errors.NewError("invalid response code")
	}

	body := resSSLValidateTransaction{}
	if err := resp.UnmarshalBody(&body); err != nil {
		return err
	}

	if body.Status != "VALID" && body.Status != "VALIDATED" {
		return errors.NewError("Transaction isn't valid")
	}

	capturedAmount := int64(0)
	orderID := ""

	am, _ := strconv.ParseFloat(body.Amount, 64)
	capturedAmount += int64(am * 100)
	orderID = body.TranID

	if orderID != orderDetails.ID {
		return errors.NewError("transaction isn't valid for the order")
	}

	log.Log().Infoln("Amount : ", orderDetails.GrandTotal)
	log.Log().Infoln("Target : ", capturedAmount)

	if capturedAmount != orderDetails.GrandTotal {
		return errors.NewError("invalid transaction amount")
	}

	return nil
}

func (ssl *sslCommerzPaymentGateway) VoidTransaction(orderDetails *models.OrderDetailsView, params map[string]interface{}) error {
	if orderDetails.TransactionID == nil {
		return errors.NewError("invalid transactionID")
	}

	url := fmt.Sprintf("%s/validator/api/merchantTransIDvalidationAPI.php?sessionkey=%s&store_id=%s&store_passwd=%s&format=json",
		ssl.Host, *orderDetails.TransactionID, ssl.StoreID, ssl.StorePassword)

	req := gohttp.NewRequest().
		Headers(map[string]string{
			"Accept": "application/json",
		})

	resp, err := req.Post(url)
	if err != nil {
		return err
	}

	if resp.GetStatusCode() != http.StatusOK {
		return errors.NewError("invalid response code")
	}

	body := resSSLValidateTransaction{}
	if err := resp.UnmarshalBody(&body); err != nil {
		return err
	}

	if body.Status != "VALID" && body.Status != "VALIDATED" {
		return errors.NewError("Transaction isn't valid")
	}

	comment := params["reason"].(string)
	comment = url2.QueryEscape(comment)
	refundAmount := orderDetails.GrandTotal - orderDetails.PaymentProcessingFee
	amountToAdjust := float64(refundAmount) / 100

	url = fmt.Sprintf("%s/validator/api/merchantTransIDvalidationAPI.php?", ssl.Host) +
		fmt.Sprintf("store_id=%s", ssl.StoreID) +
		fmt.Sprintf("&store_passwd=%s", ssl.StorePassword) +
		fmt.Sprintf("&bank_tran_id=%s", body.BankTransactionID) +
		fmt.Sprintf("&refund_amount=%.2f", amountToAdjust) +
		fmt.Sprintf("&refund_remarks=%s", comment) +
		fmt.Sprintf("&format=%s", "json")

	resp, err = gohttp.NewRequest().Headers(map[string]string{
		"Accept": "application/json",
	}).Get(url)
	if err != nil {
		return err
	}

	if resp.GetStatusCode() != http.StatusOK {
		return errors.NewError(fmt.Sprintf("invalid response status code : %d", resp.GetStatusCode()))
	}

	result := map[string]interface{}{}
	if err := resp.UnmarshalBody(&result); err != nil {
		return err
	}

	if result["status"].(string) != "success" {
		return errors.NewError("Refund request failed")
	}

	return nil
}

func (ssl *sslCommerzPaymentGateway) DisplayName() string {
	return "SSLCommerz"
}
