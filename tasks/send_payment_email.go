package tasks

import (
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/data"
	"github.com/shopicano/shopicano-backend/log"
	"github.com/shopicano/shopicano-backend/services"
	"time"
)

const (
	SendPaymentConfirmationEmailTaskName = "send_payment_confirmation_email"
	SendPaymentRevertedEmailTaskName     = "send_payment_reverted_email"
)

func SendPaymentConfirmationEmailFn(orderID string) error {
	db := app.DB()

	orderDao := data.NewOrderRepository()
	o, err := orderDao.GetDetails(db, orderID)
	if err != nil {
		log.Log().Errorln(err)
		return tasks.NewErrRetryTaskLater(err.Error(), time.Second*30)
	}

	userDao := data.NewUserRepository()
	u, err := userDao.Get(db, o.UserID)
	if err != nil {
		log.Log().Errorln(err)
		return tasks.NewErrRetryTaskLater(err.Error(), time.Second*30)
	}

	if err := services.SendOrderDetailsEmail(u.Email, "Your payment has been received.", o); err != nil {
		log.Log().Errorln(err)
		return tasks.NewErrRetryTaskLater(err.Error(), time.Second*30)
	}
	return nil
}

func SendPaymentRevertedEmailFn(orderID string) error {
	db := app.DB()

	orderDao := data.NewOrderRepository()
	order, err := orderDao.GetDetails(db, orderID)
	if err != nil {
		log.Log().Errorln(err)
		return tasks.NewErrRetryTaskLater(err.Error(), time.Second*30)
	}

	userDao := data.NewUserRepository()
	u, err := userDao.Get(db, order.UserID)
	if err != nil {
		log.Log().Errorln(err)
		return tasks.NewErrRetryTaskLater(err.Error(), time.Second*30)
	}

	if err := services.SendOrderDetailsEmail(u.Email, "Your payment has been refunded.", order); err != nil {
		log.Log().Errorln(err)
		return tasks.NewErrRetryTaskLater(err.Error(), time.Second*30)
	}
	return nil
}
