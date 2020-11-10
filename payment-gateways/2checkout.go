package payment_gateways

import (
	"errors"
	"fmt"
	"github.com/nahid/gohttp"
	"github.com/shopicano/shopicano-backend/log"
	"github.com/shopicano/shopicano-backend/models"
	"io/ioutil"
	"net/http"
	url2 "net/url"
	"strconv"
	"strings"
)

const (
	TwoCheckoutPaymentGatewayName = "2co"
)

type twoCheckoutPaymentGateway struct {
	Host            string
	SuccessCallback string
	FailureCallback string
	PublicKey       string
	PrivateKey      string
	MerchantCode    string
	Username        string
	Password        string
	SecretKey       string
}

func NewTwoCheckoutPaymentGateway(cfg map[string]interface{}) (*twoCheckoutPaymentGateway, error) {
	publicKey := cfg["public_key"].(string)
	privateKey := cfg["private_key"].(string)
	merchantCode := cfg["merchant_code"].(string)
	secretKey := cfg["secret_key"].(string)
	username := cfg["username"].(string)
	password := cfg["password"].(string)
	host := cfg["host"].(string)

	return &twoCheckoutPaymentGateway{
		SuccessCallback: cfg["success_callback"].(string),
		FailureCallback: cfg["failure_callback"].(string),
		PublicKey:       publicKey,
		PrivateKey:      privateKey,
		MerchantCode:    merchantCode,
		SecretKey:       secretKey,
		Username:        username,
		Password:        password,
		Host:            host,
	}, nil
}

func (tco *twoCheckoutPaymentGateway) GetName() string {
	return TwoCheckoutPaymentGatewayName
}

func (tco *twoCheckoutPaymentGateway) Pay(orderDetails *models.OrderDetailsView) (*PaymentGatewayResponse, error) {
	url := fmt.Sprintf("%s/checkout/purchase", tco.Host)

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

	grandTotal := float64(orderDetails.GrandTotal) / 100

	log.Log().Infoln("Grand Total : ", grandTotal)

	payload += fmt.Sprintf("li_0_type=%s&", "product")
	payload += fmt.Sprintf("li_0_name=%s&", fmt.Sprintf("Payment for Order %s", orderDetails.Hash))
	payload += fmt.Sprintf("li_0_price=%s&", fmt.Sprintf("%.2f", grandTotal))
	payload += fmt.Sprintf("li_0_quantity=%s&", fmt.Sprintf("%d", 1))
	payload += fmt.Sprintf("li_0_tangible=%s&", "N")

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

type resInvoice struct {
	Status        string `json:"status"`
	USDTotal      string `json:"usd_total"`
	VendorOrderID string `json:"vendor_order_id"`
}

type resSale struct {
	InvoiceID string       `json:"invoice_id"`
	Invoices  []resInvoice `json:"invoices"`
}

type resValidateTransaction struct {
	Sale *resSale `json:"sale"`
}

func (tco *twoCheckoutPaymentGateway) ValidateTransaction(orderDetails *models.OrderDetailsView) error {
	if orderDetails.TransactionID == nil {
		return errors.New("invalid transactionID")
	}

	url := fmt.Sprintf("%s/api/sales/detail_sale?invoice_id=%s", tco.Host, *orderDetails.TransactionID)
	req := gohttp.NewRequest().
		BasicAuth(tco.Username, tco.Password).
		Headers(map[string]string{
			"Accept": "application/json",
		})

	resp, err := req.Get(url)
	if err != nil {
		return err
	}

	if resp.GetStatusCode() != http.StatusOK {
		return errors.New("invalid response code")
	}

	body := resValidateTransaction{}
	if err := resp.UnmarshalBody(&body); err != nil {
		return err
	}

	if body.Sale == nil {
		return errors.New("invalid transaction")
	}

	capturedAmount := int64(0)
	orderID := ""

	for _, in := range body.Sale.Invoices {
		log.Log().Infoln(in)
		log.Log().Infoln(in.Status)

		if in.Status != "deposited" && in.Status != "approved" {
			return errors.New("invalid transaction status")
		}

		am, _ := strconv.ParseFloat(in.USDTotal, 64)
		capturedAmount += int64(am * 100)
		orderID = in.VendorOrderID
	}

	if orderID != orderDetails.ID {
		return errors.New("transaction isn't valid for the order")
	}

	log.Log().Infoln("Amount : ", orderDetails.GrandTotal)
	log.Log().Infoln("Target : ", capturedAmount)

	if capturedAmount != orderDetails.GrandTotal {
		return errors.New("invalid transaction amount")
	}

	return nil
}

func (tco *twoCheckoutPaymentGateway) VoidTransaction(orderDetails *models.OrderDetailsView, params map[string]interface{}) error {
	if orderDetails.TransactionID == nil {
		return errors.New("invalid transactionID")
	}

	category := 5

	typ := params["type"].(int)
	switch typ {
	case 1:
		category = 17 // Duplicate
	case 2:
		category = 4 // Fraud
	}

	comment := params["reason"].(string)
	comment = url2.QueryEscape(comment)
	refundAmount := orderDetails.GrandTotal - orderDetails.PaymentProcessingFee
	amountToAdjust := float64(refundAmount) / 100

	url := fmt.Sprintf("%s/api/sales/refund_invoice?", tco.Host) +
		fmt.Sprintf("invoice_id=%s", *orderDetails.TransactionID) +
		fmt.Sprintf("&amount=%.2f", amountToAdjust) +
		"&currency=usd" +
		fmt.Sprintf("&category=%d", category) +
		fmt.Sprintf("&comment=%s", comment)

	method := "POST"

	payload := strings.NewReader("")

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return err
	}

	req.SetBasicAuth(tco.Username, tco.Password)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	fmt.Println(string(body))

	if res.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("invalid response status code : %d", res.StatusCode))
	}

	return nil
}

func (tco *twoCheckoutPaymentGateway) DisplayName() string {
	return "2Checkout"
}
