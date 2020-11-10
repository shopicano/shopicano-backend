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
	SendOrderDetailsEmailTaskName = "send_order_details_email"
)

func SendOrderDetailsEmailFn(orderID, subject string) error {
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

	if err := services.SendOrderDetailsEmail(u.Email, subject, o); err != nil {
		log.Log().Errorln(err)
		return tasks.NewErrRetryTaskLater(err.Error(), time.Second*30)
	}
	return nil
}
