package payment_gateways

import (
	"errors"
	"fmt"
	"github.com/nahid/gohttp"
	"github.com/shopicano/shopicano-backend/config"
	"github.com/shopicano/shopicano-backend/log"
	"github.com/shopicano/shopicano-backend/models"
	"net/http"
	"strconv"
)

const (
	PaddlePaymentGatewayName = "paddle"
)

type paddlePaymentGateway struct {
	Host            string
	VendorID        string
	VendorAuthCode  string
	SuccessCallback string
	FailureCallback string
}

func NewPaddlePaymentGateway(cfg map[string]interface{}) (*paddlePaymentGateway, error) {
	return &paddlePaymentGateway{
		SuccessCallback: cfg["success_callback"].(string),
		FailureCallback: cfg["failure_callback"].(string),
		VendorID:        cfg["vendor_id"].(string),
		VendorAuthCode:  cfg["vendor_auth_code"].(string),
		Host:            cfg["host"].(string),
	}, nil
}

func (pd *paddlePaymentGateway) GetName() string {
	return PaddlePaymentGatewayName
}

type respGenerateLinkResponse struct {
	URL *string `json:"url"`
}

type respGenerateLink struct {
	Response *respGenerateLinkResponse `json:"response"`
}

func (pd *paddlePaymentGateway) Pay(orderDetails *models.OrderDetailsView) (*PaymentGatewayResponse, error) {
	url := fmt.Sprintf("%s/api/2.0/product/generate_pay_link", pd.Host)

	grandTotal := float64(orderDetails.GrandTotal) / 100

	orderPath := fmt.Sprintf(config.PathMappingCfg()["after_payment_completed"], orderDetails.ID)
	paymentCompletedCallback := fmt.Sprintf("%s%s", config.App().FrontStoreUrl, orderPath)

	params := map[string]string{
		"vendor_auth_code":  pd.VendorAuthCode,
		"vendor_id":         pd.VendorID,
		"title":             fmt.Sprintf("Payment for Order %s", orderDetails.Hash),
		"webhook_url":       fmt.Sprintf(pd.SuccessCallback, orderDetails.ID),
		"prices[0]":         fmt.Sprintf("USD:%.2f", grandTotal),
		"quantity":          "1",
		"quantity_variable": "0",
		"customer_email":    orderDetails.BillingEmail,
		//"customer_country":  "BD",
		"customer_postcode": orderDetails.BillingPostcode,
		"passthrough":       orderDetails.ID,
		"return_url":        paymentCompletedCallback,
	}

	resp, err := gohttp.NewRequest().FormData(params).Headers(map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}).Post(url)
	if err != nil {
		return nil, err
	}

	body := respGenerateLink{}
	if err := resp.UnmarshalBody(&body); err != nil {
		return nil, err
	}

	if !(body.Response != nil && body.Response.URL != nil) {
		return nil, errors.New("failed to generate checkout url")
	}

	return &PaymentGatewayResponse{
		Result: *body.Response.URL,
	}, nil
}

func (pd *paddlePaymentGateway) GetConfig() (map[string]interface{}, error) {
	cfg := map[string]interface{}{
		"success_callback_url": pd.SuccessCallback,
		"failure_callback_url": pd.FailureCallback,
	}
	return cfg, nil
}

type resTransactionResponsePaddle struct {
	OrderID        string `json:"order_id"`
	Amount         string `json:"amount"`
	Status         string `json:"status"`
	PassThrough    string `json:"passthrough"`
	IsSubscription bool   `json:"is_subscription"`
}

type resValidateTransactionPaddle struct {
	Response []resTransactionResponsePaddle `json:"response"`
}

func (pd *paddlePaymentGateway) ValidateTransaction(orderDetails *models.OrderDetailsView) error {
	if orderDetails.TransactionID == nil {
		return errors.New("invalid transactionID")
	}

	url := fmt.Sprintf("%s/api/2.0/order/%s/transactions", pd.Host, *orderDetails.TransactionID)
	req := gohttp.NewRequest().
		Headers(map[string]string{
			"Accept":       "application/json",
			"Content-Type": "application/x-www-form-urlencoded",
		}).
		FormData(map[string]string{
			"vendor_auth_code": pd.VendorAuthCode,
			"vendor_id":        pd.VendorID,
		})

	resp, err := req.Post(url)
	if err != nil {
		return err
	}

	if resp.GetStatusCode() != http.StatusOK {
		return errors.New("invalid response code")
	}

	body := resValidateTransactionPaddle{}
	if err := resp.UnmarshalBody(&body); err != nil {
		return err
	}

	if len(body.Response) == 0 {
		return errors.New("no transaction information found")
	}

	capturedAmount := int64(0)
	orderID := ""

	for _, in := range body.Response {
		log.Log().Infoln(in)
		log.Log().Infoln(in.Status)

		if in.Status != "completed" {
			return errors.New("invalid transaction status")
		}

		am, _ := strconv.ParseFloat(in.Amount, 64)
		capturedAmount += int64(am * 100)
		orderID = in.PassThrough
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

func (pd *paddlePaymentGateway) VoidTransaction(orderDetails *models.OrderDetailsView, params map[string]interface{}) error {
	//if orderDetails.TransactionID == nil {
	//	return errors.New("invalid transactionID")
	//}
	//
	//category := 5
	//
	//typ := params["type"].(int)
	//switch typ {
	//case 1:
	//	category = 17 // Duplicate
	//case 2:
	//	category = 4 // Fraud
	//}
	//
	//comment := params["reason"].(string)
	//comment = url2.QueryEscape(comment)
	//refundAmount := orderDetails.GrandTotal - orderDetails.PaymentProcessingFee
	//amountToAdjust := float64(refundAmount) / 100
	//
	//url := fmt.Sprintf("%s/api/sales/refund_invoice?", pd.Host) +
	//	fmt.Sprintf("invoice_id=%s", *orderDetails.TransactionID) +
	//	fmt.Sprintf("&amount=%.2f", amountToAdjust) +
	//	"&currency=usd" +
	//	fmt.Sprintf("&category=%d", category) +
	//	fmt.Sprintf("&comment=%s", comment)
	//
	//method := "POST"
	//
	//payload := strings.NewReader("")
	//
	//client := &http.Client{}
	//req, err := http.NewRequest(method, url, payload)
	//if err != nil {
	//	return err
	//}
	//
	//req.SetBasicAuth(pd.Username, pd.Password)
	//req.Header.Add("Accept", "application/json")
	//req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	//
	//res, err := client.Do(req)
	//if err != nil {
	//	return err
	//}
	//
	//defer res.Body.Close()
	//body, err := ioutil.ReadAll(res.Body)
	//fmt.Println(string(body))
	//
	//if res.StatusCode != http.StatusOK {
	//	return errors.New(fmt.Sprintf("invalid response status code : %d", res.StatusCode))
	//}

	return nil
}

func (pd *paddlePaymentGateway) DisplayName() string {
	return "Paddle"
}
