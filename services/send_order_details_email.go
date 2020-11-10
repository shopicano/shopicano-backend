package services

import (
	"fmt"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/go-gomail/gomail"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/config"
	"github.com/shopicano/shopicano-backend/data"
	"github.com/shopicano/shopicano-backend/log"
	"github.com/shopicano/shopicano-backend/models"
	payment_gateways "github.com/shopicano/shopicano-backend/payment-gateways"
	"github.com/shopicano/shopicano-backend/templates"
	"github.com/shopicano/shopicano-backend/utils"
)

func SendOrderDetailsEmail(email, subject string, order *models.OrderDetailsView) error {
	pu := data.NewMarketplaceRepository()
	settings, err := pu.GetSettings(app.DB())
	if err != nil {
		return err
	}

	params := map[string]interface{}{}
	params["greetings"] = fmt.Sprintf("Hi %s,", order.UserName)
	params["intros"] = subject
	params["orderHash"] = order.Hash
	params["billingAddress"] = fmt.Sprintf("%s, %s, %s - %s",
		order.BillingAddress, order.BillingCity, order.BillingCountry, order.BillingPostcode)
	params["isShippable"] = !order.IsAllDigitalProducts
	params["buyerName"] = order.UserName
	params["orderDate"] = order.CreatedAt.Format(utils.DateTimeFormatForDistribution)
	orderPath := fmt.Sprintf(config.PathMappingCfg()["after_payment_completed"], order.ID)
	params["orderUrl"] = fmt.Sprintf("%s%s", config.App().FrontStoreUrl, orderPath)

	if !order.IsAllDigitalProducts {
		params["shippingAddress"] = fmt.Sprintf("%s, %s, %s - %s",
			*order.ShippingAddress, *order.ShippingCity, *order.ShippingCountry, *order.ShippingPostcode)
	} else {
		params["shippingAddress"] = "N/A"
	}

	params["shippingCharge"] = fmt.Sprintf("%.2f", float64(order.ShippingCharge)/100)
	params["paymentProcessingFee"] = fmt.Sprintf("%.2f", float64(order.PaymentProcessingFee)/100)
	params["subTotal"] = fmt.Sprintf("%.2f", float64(order.SubTotal)/100)
	params["grandTotal"] = fmt.Sprintf("%.2f", float64(order.GrandTotal)/100)
	params["isCouponApplied"] = false
	params["isDigitalPayment"] = !order.PaymentMethodIsOffline
	params["assetsUrl"] = fmt.Sprintf("%s/assets/", settings.Website)
	params["siteUrl"] = fmt.Sprintf("%s", settings.Website)
	params["platformName"] = settings.Name

	pg, err := payment_gateways.GetPaymentGatewayByName(order.PaymentGateway)
	if err != nil {
		params["paymentGateway"] = "None"
	} else {
		params["paymentGateway"] = pg.DisplayName()
	}

	switch order.Status {
	case models.OrderPending:
		params["status"] = "Pending"
	case models.OrderConfirmed:
		params["status"] = "Confirmed"
	case models.OrderCancelled:
		params["status"] = "Cancelled"
	case models.OrderShipping:
		params["status"] = "Shipping"
	case models.OrderDelivered:
		params["status"] = "Delivered"
	}

	switch order.PaymentStatus {
	case models.PaymentPending:
		params["paymentStatus"] = "Pending"
	case models.PaymentCompleted:
		params["paymentStatus"] = "Completed"
	case models.PaymentFailed:
		params["paymentStatus"] = "Failed"
	case models.PaymentReverted:
		params["paymentStatus"] = "Reverted"
	}

	if order.DiscountedAmount != 0 {
		params["couponCode"] = order.CouponCode
		params["discount"] = fmt.Sprintf("%.2f", float64(order.DiscountedAmount)/100)
		params["isCouponApplied"] = true
	}

	var items []map[string]interface{}

	for _, v := range order.Items {
		items = append(items, map[string]interface{}{
			"name":     v.Name,
			"quantity": v.Quantity,
			"price":    fmt.Sprintf("%.2f", float64(v.Price)/100),
			"subTotal": fmt.Sprintf("%.2f", float64(v.SubTotal)/100),
		})
	}

	params["orderedItems"] = items

	body, err := templates.GenerateInvoiceEmailHTML(params)
	if err != nil {
		log.Log().Errorln(err)
		return tasks.NewErrRetryTaskLater(err.Error(), TaskRetryDelay)
	}

	if err := SendEmail(subject, email, body); err != nil {
		log.Log().Errorln(err)
		return tasks.NewErrRetryTaskLater(err.Error(), TaskRetryDelay)
	}
	return nil
}

func SendEmail(subject, email, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s", config.EmailService().FromEmailAddress))
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	return EmailDialer().DialAndSend(m)
}
